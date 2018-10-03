package main

import "fmt"

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

func (tr *trie) hasPrefix(s string) bool {
	if n := tr.searchPrefix(s); n != nil {
		return n.end
	}

	return false
}

func (tr *trie) get(s string) string {
	if n := tr.searchPrefix(s); n != nil {
		return n.data
	}

	return ""
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

		if n := tree.searchPrefix(s); n == nil {
			panic(fmt.Sprintf("expected %s to be found", s))
		}

		if expected, got := data, tree.get(s); expected != got {
			panic(fmt.Sprintf("expected %s key to has data: '%s' but got: '%s'", s, expected, got))
		}
	}

	if tree.hasPrefix("/first/one/three") {
		panic(fmt.Sprintf("expected to not be found"))
	}

	if tree.hasPrefix("/first/one/") {
		panic(fmt.Sprintf("expected to not be found"))
	}

	if tree.hasPrefix("/f") {
		panic(fmt.Sprintf("expected to not be found"))
	}
}
