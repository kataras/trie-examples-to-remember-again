package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kataras/iris/context"
)

/* go get github.com/kataras/iris */

type trieNode struct {
	parent *trieNode

	children         map[string]*trieNode
	segment          string // the part of the node path without the slash.
	isNamedParameter bool   // is segment a named parameter?
	isWildcard       bool   // allow everything else after that path prefix but it checks for static paths and named parameters before that in order to support everything that other implementations do not.
	// paramKey         string // does one of the children contains a parameter name and if so then which key does its node belongs to, starts with ':' or '*'?

	// paramKeys []string
	paramKeys map[int]string

	// on insert.
	end bool   // it is a complete node, here we stop and we can say that the node is valid.
	key string // if end == true then key is filled with the original value of the insertion's key.
	// insert data.
	Handlers  context.Handlers
	RouteName string
}

func newTrieNode() *trieNode {
	n := new(trieNode)
	return n
}

func (tn *trieNode) hasChild(s string) (has bool) {
	_, has = tn.children[s]
	return
}

func (tn *trieNode) getChild(s string) *trieNode {
	if n, ok := tn.children[s]; ok {
		return n
	}

	return nil
}

func (tn *trieNode) addChild(s string, n *trieNode) {
	if tn.children == nil {
		tn.children = make(map[string]*trieNode)
	}

	n.parent = tn
	n.segment = s
	tn.children[s] = n
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
}

func newTrie() *trie {
	return &trie{
		root: newTrieNode(),
	}
}

const pathSep = "/"

const dynamicPseudoPath = ">param"

func (tr *trie) insert(path, routeName string, handlers context.Handlers) {
	input := strings.Split(path, pathSep)[1:]
	// input := strings.FieldsFunc(key, func(r rune) bool {
	// 	return r == '/'
	// })
	n := tr.root

	for i, s := range input {
		// if c := s[0]; c == ':' || c == '*' {
		// 	if !n.hasChild(dynamicPseudoPath) {
		// 		child := newTrieNode()
		// 		child.segment = s
		// 		paramKeys = append(paramKeys, s)
		// 		n.addChild(dynamicPseudoPath, child)
		// 	}
		// 	n = n.getChild(dynamicPseudoPath) //.addChild(s, newTrieNode())

		// 	continue
		// }
		if !n.hasChild(s) {
			child := newTrieNode()
			if c := s[0]; c == ':' {
				if n.paramKeys == nil {
					n.paramKeys = make(map[int]string)
				}
				n.paramKeys[i] = s
				child.isNamedParameter = true
			} else if c == '*' {
				if n.paramKeys == nil {
					n.paramKeys = make(map[int]string)
				}
				n.paramKeys[i] = s
				// or on the parent 'n'?
				child.isWildcard = true
			}
			n.addChild(s, child)
		}
		n = n.getChild(s)
	}

	n.RouteName = routeName
	n.Handlers = handlers
	n.key = path
	n.end = true
}

func (tr *trie) searchPrefix(prefix string) *trieNode {
	input := strings.Split(prefix, pathSep)[1:]
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

func (tr *trie) search(s string) *trieNode {
	if n := tr.searchPrefix(s); n != nil && n.end {
		return n
	}

	return nil
}

func (tr *trie) searchAgainst(q string, params *context.RequestParams) *trieNode {
	n := tr.root

	start := 1
	i := 1
	pathCount := 0
	for n != nil {
		c := q[i]
		if c == '/' {
			// word := q[start:i]
			// ("c == /: " + word)
			if child := n.getChild(q[start:i]); child != nil {
				n = child
			} else {
				if paramKey, ok := n.paramKeys[pathCount]; ok {
					n = n.getChild(paramKey)
					if n.isWildcard {
						// println("wildcard: " + paramKey[1:] + " = " + q[start:])
						params.Set(paramKey[1:], q[start:])

						break
					} else {
						// println(paramKey[1:] + " = " + q[start:i])
						params.Set(paramKey[1:], q[start:i])
					}
					// 	n = n.getChild(paramKey)
				}
			}

			i++
			start = i
			pathCount++
			continue
		}

		i++

		// if end and no slash...
		if i == len(q) {
			// word := q[start:]
			// println("i == len(q): " + word)
			if child := n.getChild(q[start:]); child != nil {
				n = child
			} else {
				if paramKey, ok := n.paramKeys[pathCount]; ok {
					// println("ending... " + paramKey[1:] + " = " + q[start:])
					params.Set(paramKey[1:], q[start:])
					n = n.getChild(paramKey)
				}
			}

			break
		}
	}

	if n.isEnd() {
		return n
	}

	return nil
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
	tests := map[string]string{
		"/first":                "first_data",
		"/first/one":            "first/one_data",
		"/first/one/two":        "first/one/two_data",
		"/firstt":               "firstt_data",
		"/second":               "second_data",
		"/second/one":           "second/one_data",
		"/second/one/two":       "second/one/two_data",
		"/second/one/two/three": "second/one/two/three_data",

		// named parameters.
		"/first/one/with/:param":                         "first/one/with_data_param",
		"/first/one/with/:param/static/:otherparam":      "first/one/with/static/_data_otherparam",
		"/first/one/with/:param1/:param2/:param3":        "first/one/with/with_data_threeparams",
		"/first/one/with/:param1/:param2/:param3/static": "first/one/with/static/_data_otherparams_with_static_end",
		// wildcard named parameters.
		"/second/wild/*mywildcardparam": "second/wildcard_1",
		// no wildcard but same prefix.
		"/second/wild/static": "second/no_wild",
		// no wildcard, parameter instead with same prefix.
		"/sectond/wild/:param": "second/no_wild_but_param",
		// root wildcard.
		"/*anything": "root_wildcard",
	}

	for s, routeName := range tests {
		tree.insert(s, routeName, nil)

		if n := tree.search(s); n == nil {
			panic(fmt.Sprintf("expected %s to be found", s))
		} else {
			if expected, got := routeName, n.RouteName; expected != got {
				panic(fmt.Sprintf("expected %s key to has route name: '%s' but got: '%s'", s, expected, got))
			}
		}
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
	n := tree.search(keyToTest).parent
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

	if n = tree.searchAgainst(keyToTest, params); n == nil {
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
	if n = tree.searchAgainst(keyToTest, params); n == nil {
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
	if n = tree.searchAgainst(keyToTest, params); n == nil {
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
	if n = tree.searchAgainst(keyToTest, params); n == nil {
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
	if n = tree.searchAgainst(keyToTest, params); n == nil {
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
	if n = tree.searchAgainst(keyToTest, params); n == nil {
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
	if n = tree.searchAgainst(keyToTest, params); n == nil {
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
	if n = tree.searchAgainst(keyToTest, params); n == nil {
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
