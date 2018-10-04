package main

import (
	"fmt"
	"sort"
	"strings"
)

type trieNode struct {
	parent *trieNode

	records map[rune]*trieNode

	end bool
	key string // if end == true then key is filled with the original value of the insertion's key.

	//
	data string
}

func newTrieNode() *trieNode {
	n := new(trieNode)
	return n
}

func (tn *trieNode) hasRecord(c rune) (has bool) {
	_, has = tn.records[c]
	return
}

func (tn *trieNode) hasNoRecord() bool { // or "leaf" as we used to describe nodes without children.
	return len(tn.records) == 0
}

func (tn *trieNode) getRecord(c rune) *trieNode {
	if n, ok := tn.records[c]; ok {
		return n
	}

	return nil
}

func (tn *trieNode) addRecord(c rune, n *trieNode) {
	if tn.records == nil {
		tn.records = make(map[rune]*trieNode)
	}

	n.parent = tn
	tn.records[c] = n
}

func (tn *trieNode) isEnd() bool {
	return tn.end
}

func (tn *trieNode) getKeys(sorted bool) (list []string) {
	if tn.isEnd() {
		list = append(list, tn.key)
	}

	if tn.records != nil {
		for _, child := range tn.records {
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
			return len(strings.Split(list[i], "/")) < len(strings.Split(list[j], "/"))
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

func (tr *trie) insert(s string, data string) {
	input := []rune(s)
	n := tr.root

	for _, c := range input {
		if !n.hasRecord(c) {
			child := newTrieNode()
			n.addRecord(c, child)
		}
		n = n.getRecord(c)
	}

	n.data = data
	n.key = s
	n.end = true
}

func (tr *trie) searchPrefix(s string) *trieNode {
	input := []rune(s)
	n := tr.root
	for i := 0; i < len(input); i++ {
		c := input[i]
		if n.hasRecord(c) {
			n = n.getRecord(c)
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

	if !tree.hasPrefix("/first/one/") {
		panic(fmt.Sprintf("[1] expected to has prefix"))
	}

	if !tree.hasPrefix("/fir") {
		panic(fmt.Sprintf("[2] expected to has prefix"))
	}

	if !tree.hasPrefix("/f") {
		panic(fmt.Sprintf("[3] expected to has found"))
	}

	keyToTest := "/firs"
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

}

func describe(title string, args ...interface{}) {
	fmt.Printf("🌀  %s ⤵️\n", fmt.Sprintf(title, args...))
}

/*
	🌀  autocomplete of "/firs" ⤵️
	/first
	/firstt
	/first/one
	/first/one/two
	🌀  find final parents of "/second/one/two/three" ⤵️
	/second/one/two
	/second/one
	/second
*/
