package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kataras/iris/context"
)

/* go get github.com/kataras/iris */

const (
	param    = ":" // is segment a named parameter?
	wildcard = "*" // allow everything else after that path prefix but it checks for static paths and named parameters before that in order to support everything that other implementations do not.
)

type trieNode struct {
	parent *trieNode

	children               map[string]*trieNode
	hasDynamicChild        bool // does one of the children contains a parameter or wildcard?
	childNamedParameter    bool // is the child a named parameter (single segmnet)
	childWildcardParameter bool // or it is a wildcard (can be more than one path segments) ?

	paramKeys []string // the param keys without : or *.
	end       bool     // it is a complete node, here we stop and we can say that the node is valid.
	key       string   // if end == true then key is filled with the original value of the insertion's key.
	// if key != "" && its parent has childWildcardParameter == true,
	// we need it to track the static part for the closest-wildcard's parameter storage.
	staticKey string

	// insert data.
	Handlers  context.Handlers
	RouteName string
}

func newTrieNode() *trieNode {
	n := new(trieNode)
	return n
}

func (tn *trieNode) hasChild(s string) bool {
	return tn.getChild(s) != nil
}

func (tn *trieNode) getChild(s string) *trieNode {
	if tn.children == nil {
		tn.children = make(map[string]*trieNode)
	}

	return tn.children[s]
}

func (tn *trieNode) addChild(s string, n *trieNode) {
	if tn.children == nil {
		tn.children = make(map[string]*trieNode)
	}

	if _, exists := tn.children[s]; exists {
		return
	}

	n.parent = tn
	tn.children[s] = n
}

func (tn *trieNode) findClosestParentWildcardNode() *trieNode {
	tn = tn.parent
	for tn != nil {
		if tn.childWildcardParameter {
			return tn.getChild(wildcard)
		}

		tn = tn.parent
	}

	return nil
}

func (tn *trieNode) isEnd() bool {
	return tn.end
}

func (tn *trieNode) getKeys(sorted bool) (list []string) {
	if tn.isEnd() {
		list = append(list, tn.key)
	}

	if tn.children != nil {
		for _, child := range tn.children {
			list = append(list, child.getKeys(sorted)...)
		}
	}

	// 1:
	//	sort.Strings(list)
	//
	// 2:
	// sort.Slice(list, func(i, j int) bool {
	// 	return len(list[i]) < len(list[j])
	// })
	//
	// 3:
	// first the "key (the path)" with the lowest number of slashes.
	if sorted {
		sort.Slice(list, func(i, j int) bool {
			return len(strings.Split(list[i], pathSep)) < len(strings.Split(list[j], pathSep))
		})
	}

	return
}

func (tn *trieNode) String() string {

	return ""
}

type trie struct {
	root *trieNode

	// if true then it will handle any path if not other parent wildcard exists,
	// so even 404 (on http services) is up to it, see trie#insert.
	hasRootWildcard bool
}

func newTrie() *trie {
	return &trie{
		root: newTrieNode(),
	}
}

const (
	pathSep  = "/"
	pathSepB = '/'
)

func slowPathSplit(path string) []string {
	if path == pathSep {
		return []string{pathSep}
	}

	return strings.Split(path, pathSep)[1:]
}

func (tr *trie) insert(path, routeName string, handlers context.Handlers) {
	input := slowPathSplit(path)

	n := tr.root
	var paramKeys []string

	for _, s := range input {
		c := s[0]

		if isParam, isWildcard := c == param[0], c == wildcard[0]; isParam || isWildcard {
			n.hasDynamicChild = true
			paramKeys = append(paramKeys, s[1:]) // without : or *.

			// if node has already a wildcard, don't force a value, check for true only.
			if isParam {
				n.childNamedParameter = true
				s = param
			}

			if isWildcard {
				n.childWildcardParameter = true
				s = wildcard
				if tr.root == n {
					tr.hasRootWildcard = true
				}
			}
		}

		if !n.hasChild(s) {
			child := newTrieNode()
			n.addChild(s, child)
		}

		n = n.getChild(s)
	}

	n.RouteName = routeName
	n.Handlers = handlers
	n.paramKeys = paramKeys
	n.key = path
	n.end = true

	i := strings.Index(path, param)
	if i == -1 {
		i = strings.Index(path, wildcard)
	}
	if i == -1 {
		i = len(n.key)
	}

	n.staticKey = path[:i]
}

func (tr *trie) searchPrefix(prefix string) *trieNode {
	input := slowPathSplit(prefix)
	n := tr.root

	for i := 0; i < len(input); i++ {
		s := input[i]
		if child := n.getChild(s); child != nil {
			n = child
			continue
		}

		return nil
	}

	return n
}

func (tr *trie) search(q string, params *context.RequestParams) *trieNode {
	end := len(q)
	n := tr.root
	if end == 1 && q[0] == pathSepB {
		return n.getChild(pathSep)
	}

	start := 1
	i := 1
	var paramValues []string

	for {
		if i == end || q[i] == pathSepB {
			if child := n.getChild(q[start:i]); child != nil {
				n = child
			} else if n.childNamedParameter { // && n.childWildcardParameter == false {
				//	println("dynamic NAMED element for: " + q[start:i] + " found ")
				n = n.getChild(param)
				if ln := len(paramValues); cap(paramValues) > ln {
					paramValues = paramValues[:ln+1]
					paramValues[ln] = q[start:i]
				} else {
					paramValues = append(paramValues, q[start:i])
				}
			} else if n.childWildcardParameter {
				//	println("dynamic WILDCARD element for: " + q[start:i] + " found ")
				n = n.getChild(wildcard)
				if ln := len(paramValues); cap(paramValues) > ln {
					paramValues = paramValues[:ln+1]
					paramValues[ln] = q[start:]
				} else {
					paramValues = append(paramValues, q[start:])
				}
				break
			} else {
				n = n.findClosestParentWildcardNode()
				if n != nil {
					// means that it has :param/static and *wildcard, we go trhough the :param
					// but the next path segment is not the /static, so go back to *wildcard
					// instead of not found.
					//
					// Fixes:
					// /hello/*p
					// /hello/:p1/static/:p2
					// req: http://localhost:8080/hello/dsadsa/static/dsadsa => found
					// req: http://localhost:8080/hello/dsadsa => but not found!
					// and
					// /second/wild/*p
					// /second/wild/static/otherstatic/
					// req: /second/wild/static/otherstatic/random => but not found!
					params.Set(n.paramKeys[0], q[len(n.staticKey):])
					return n
				}

				return nil
			}

			if i == end {
				break
			}

			i++
			start = i
			continue
		}

		i++

		// if i == end {
		// same...
		// 	break
		// }
	}

	if n == nil || !n.isEnd() {
		if tr.hasRootWildcard {
			// that's the case for root wildcard, tests are passing
			// even without it but stick with it for reference.
			// Note ote that something like:
			// Routes: /other2/*myparam and /other2/static
			// Reqs: /other2/staticed will be handled
			// the /other2/*myparam and not the root wildcard, which is what we want.
			//
			n = tr.root.getChild(wildcard)
			params.Set(n.paramKeys[0], q[1:])
			return n
		}

		return nil
	}

	for i, paramValue := range paramValues {
		if len(n.paramKeys) > i {
			params.Set(n.paramKeys[i], paramValue)
		}
	}

	return n
}

func (tr *trie) hasPrefix(s string) bool {
	return tr.searchPrefix(s) != nil
}

func (tr *trie) autocomplete(s string, sorted bool) (list []string) {
	n := tr.searchPrefix(s)
	if n != nil {
		list = n.getKeys(sorted)
	}
	return
}

func main() {
	tree := newTrie()

	tests := []struct {
		Path      string
		RouteName string
	}{
		{"/first", "first_data"},
		{"/first/one", "first/one_data"},
		{"/first/one/two", "first/one/two_data"},
		{"/firstt", "firstt_data"},
		{"/second", "second_data"},
		{"/second/one", "second/one_data"},
		{"/second/one/two", "second/one/two_data"},
		{"/second/one/two/three", "second/one/two/three_data"},

		// named parameters.
		{"/first/one/with/:param1/:param2/:param3/static", "first/one/with/static/_data_otherparams_with_static_end"},
		{"/first/one/with/:param1/:param2/:param3", "first/one/with/with_data_threeparams"},
		{"/first/one/with/:param/static/:otherparam", "first/one/with/static/_data_otherparam"},
		{"/first/one/with/:param", "first/one/with_data_param"},

		// wildcard named parameters.
		{"/second/wild/*mywildcardparam", "second/wildcard_1"},
		// no wildcard but same prefix.
		{"/second/wild/static", "second/no_wild"},
		// no wildcard, parameter instead with same prefix.
		{"/sectond/wild/:param", "second/no_wild_but_param"},
		// root wildcard.
		{"/*anything", "root_wildcard"},
	}

	for _, tt := range tests {
		tree.insert(tt.Path, tt.RouteName, nil)

		// if n := tree.search(s); n == nil {
		// 	panic(fmt.Sprintf("expected %s to be found", s))
		// } else {
		// 	if expected, got := routeName, n.RouteName; expected != got {
		// 		panic(fmt.Sprintf("expected %s key to has route name: '%s' but got: '%s'", s, expected, got))
		// 	}
		// }
	}

	if tree.hasPrefix("/first/one/three") {
		panic(fmt.Sprintf("[0] expected to not has prefix"))
	}

	if !tree.hasPrefix("/first/one") {
		panic(fmt.Sprintf("[1] expected to has prefix"))
	}

	if !tree.hasPrefix("/first") {
		panic(fmt.Sprintf("[2] expected to has prefix"))
	}

	if !tree.hasPrefix("/second") {
		panic(fmt.Sprintf("[3] expected to has found"))
	}

	keyToTest := "/first"
	describe("autocomplete of \"%s\"", keyToTest)
	keyStartsWith1 := tree.autocomplete(keyToTest, true)
	for _, s := range keyStartsWith1 {
		fmt.Println(s)
	}

	keyToTest = "/second/one/two/three"
	describe("find final parents of \"%s\"", keyToTest)
	n := tree.searchPrefix(keyToTest).parent
	for {
		if n == nil {
			break
		}

		if n.isEnd() {
			fmt.Println(n.key)
		}

		n = n.parent
	}

	// n := tree.root
	params := new(context.RequestParams)

	keyToTest = "/first/one/with/myparam"
	describe("search all nodes against \"%s\"", keyToTest)

	if n = tree.search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.RouteName)
		paramName := "param"
		if expected, got := "myparam", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()

	keyToTest = "/first/one/with/myparam1/static/myparam2"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param/static/:otherparam"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.RouteName)

		paramName := "param"
		if expected, got := "myparam1", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}

		paramName = "otherparam"
		if expected, got := "myparam2", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()

	keyToTest = "/first/one/with/myparam1/myparam2/myparam3"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param1/:param2/:param3"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.RouteName)

		paramName := "param1"
		if expected, got := "myparam1", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}

		paramName = "param2"
		if expected, got := "myparam2", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}

		paramName = "param3"
		if expected, got := "myparam3", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()

	keyToTest = "/first/one/with/myparameter1/myparameter2/myparameter3/static"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param1/:param2/:param3/static"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.RouteName)

		paramName := "param1"
		if expected, got := "myparameter1", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}

		paramName = "param2"
		if expected, got := "myparameter2", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}

		paramName = "param3"
		if expected, got := "myparameter3", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()

	keyToTest = "/second/wild/everything/else/can/go/here"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/second/wild/*mywildcardparam"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.RouteName)

		paramName := "mywildcardparam"
		if expected, got := "everything/else/can/go/here", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()

	keyToTest = "/second/wild/static"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, keyToTest))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.RouteName)

		if params.Len() > 0 {
			panic("expected 0 parameters")
		}
	}
	params.Reset()

	keyToTest = "/sectond/wild/parameter1"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/sectond/wild/:param"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.RouteName)

		paramName := "param"
		if expected, got := "parameter1", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()

	keyToTest = "/something/here/to/match/root/wildcard"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/*anything"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.RouteName)

		paramName := "anything"
		if expected, got := "something/here/to/match/root/wildcard", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()
}

func describe(title string, args ...interface{}) {
	fmt.Printf("🌀  %s ⤵️\n", fmt.Sprintf(title, args...))
}

/*
	🌀  autocomplete of "/first" ⤵️
	/first
	/first/one
	/first/one/two
	/first/one/with/:param
	/first/one/with/:param/:otherparam/:otherparam2
	/first/one/with/:param/static/:otherparam
	/first/one/with/:param/:otherparam/:otherparam2/static
	🌀  find final parents of "/second/one/two/three" ⤵️
	/second/one/two
	/second/one
	/second
	🌀  search all nodes against "/first/one/with/myparam" ⤵️
	found: '/first/one/with/:param': first/one/with_data_param
	🌀  search all nodes against "/first/one/with/myparam1/static/myparam2" ⤵️
	found: '/first/one/with/:param/static/:otherparam': first/one/with/static/_data_otherparam
	🌀  search all nodes against "/first/one/with/myparam1/myparam2/myparam3" ⤵️
	found: '/first/one/with/:param/:otherparam/:otherparam2': first/one/with/static/_data_otherparams
	🌀  search all nodes against "/first/one/with/myparameter1/myparameter2/myparameter3/static" ⤵️
	found: '/first/one/with/:param/:otherparam/:otherparam2/static': first/one/with/static/_data_otherparams_with_static_end
	🌀  search all nodes against "/second/wild/everything/else/can/go/here" ⤵️
	found: '/second/wild/*mywildcardparam': second/wildcard_1
	🌀  search all nodes against "/second/wild/static" ⤵️
	found: '/second/wild/static': second/no_wild
	🌀  search all nodes against "/sectond/wild/parameter1" ⤵️
	found: '/sectond/wild/:param': second/no_wild_but_param
	🌀  search all nodes against "/something/here/to/match/root/wildcard" ⤵️
	found: '/*anything': root_wildcard
*/
