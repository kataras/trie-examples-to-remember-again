package main

import (
	"fmt"
	"sort"
	"strings"
)

type trieNode struct {
	parent *trieNode

	children         map[string]*trieNode
	segment          string // the part of the node path without the slash.
	isNamedParameter bool   // is segment a named parameter?
	paramKey         string // does one of the children contains a parameter name and if so then which key does its node belongs to, starts with the ':'?

	end bool   // it is a complete node, here we stop and we can say that the node is valid.
	key string // if end == true then key is filled with the original value of the insertion's key.
	//
	data string // any data here, on the future the Route or each handlers will be sticked.
}

func newTrieNode() *trieNode {
	n := new(trieNode)
	return n
}

func (tn *trieNode) hasChild(s string) (has bool) {
	_, has = tn.children[s]
	return
}

func (tn *trieNode) hasNoRecord() bool { // or "leaf" as we used to describe nodes without children.
	return len(tn.children) == 0
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

func (tr *trie) insert(key string, data string) {
	input := strings.Split(key, pathSep)[1:]
	n := tr.root

	for _, s := range input {
		if !n.hasChild(s) {
			child := newTrieNode()
			if s[0] == ':' {
				n.paramKey = s
				child.isNamedParameter = true
			}
			n.addChild(s, child)
		}
		n = n.getChild(s)
	}

	n.data = data
	n.key = key
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

func (tr *trie) searchAgainst(q string) *trieNode {
	input := strings.Split(q, pathSep)[1:]

	n := tr.root
	for i := 0; i < len(input); i++ {
		s := input[i]
		if child := n.getChild(s); child != nil {
			n = child
			continue
		}

		if n.paramKey != "" {
			n = n.getChild(n.paramKey)
			// if n.isNamedParameter {
			// 	println(n.segment + " is named parameter")
			// }
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

func (tr *trie) get(s string) string {
	if n := tr.search(s); n != nil {
		return n.data
	}

	return ""
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
		"/first/one/with/:param":                                 "first/one/with_data_param",
		"/first/one/with/:param/static/:otherparam":              "first/one/with/static/_data_otherparam",
		"/first/one/with/:param/:otherparam/:otherparam2":        "first/one/with/static/_data_otherparams",
		"/first/one/with/:param/:otherparam/:otherparam2/static": "first/one/with/static/_data_otherparams_with_static_end",
	}

	for s, data := range tests {
		tree.insert(s, data)

		if n := tree.search(s); n == nil {
			panic(fmt.Sprintf("expected %s to be found", s))
		}

		if expected, got := data, tree.get(s); expected != got {
			panic(fmt.Sprintf("expected %s key to has data: '%s' but got: '%s'", s, expected, got))
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

	keyToTest = "/first/one/with/myparam"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.searchAgainst(keyToTest); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.data)
	}

	keyToTest = "/first/one/with/myparam1/static/myparam2"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.searchAgainst(keyToTest); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param/static/:otherparam"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.data)
	}

	keyToTest = "/first/one/with/myparam1/myparam2/myparam3"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.searchAgainst(keyToTest); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param/:otherparam/:otherparam2"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.data)
	}

	keyToTest = "/first/one/with/myparameter1/myparameter2/myparameter3/static"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.searchAgainst(keyToTest); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param/:otherparam/:otherparam2/static"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.key, n.data)
	}
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
	/first/one/with/:param/static/:otherparam
	/first/one/with/:param/:otherparam/:otherparam2
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
*/
