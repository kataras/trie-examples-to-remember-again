package main

import (
	"fmt"
	"sort"
	"strings"
)

type trieNode struct {
	records map[rune]*trieNode
	end     bool

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

func (tn *trieNode) hasNoRecord() bool {
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

	tn.records[c] = n
}

func (tn *trieNode) isEnd() bool {
	return tn.end
}

func (tn *trieNode) getDatas() (list []string) {
	if tn.isEnd() {
		list = append(list, tn.data)
	}

	if tn.records != nil {
		for _, child := range tn.records {
			list = append(list, child.getDatas()...)
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
	// first the "data" with the lowest number of slashes.
	sort.Slice(list, func(i, j int) bool {
		return len(strings.Split(list[i], "/")) < len(strings.Split(list[j], "/"))
	})
	return
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
			n.addRecord(c, newTrieNode())
		}
		n = n.getRecord(c)
	}

	n.data = data
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

func (tr *trie) autocomplete(s string) (list []string) {
	n := tr.searchPrefix(s)
	if n != nil {
		list = n.getDatas()
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

	startsWith1 := tree.autocomplete("/first")
	for _, s := range startsWith1 {
		fmt.Println(s)
	}
}
