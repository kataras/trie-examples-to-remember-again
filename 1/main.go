package main

import "fmt"

type trieNode struct {
	records map[rune]*trieNode
	end     bool
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

func (tr *trie) insert(s string) {
	input := []rune(s)
	n := tr.root
	for _, c := range input {
		if !n.hasRecord(c) {
			n.addRecord(c, newTrieNode())
		}
		n = n.getRecord(c)
	}

	n.end = true
}

func (tr *trie) search(s string) *trieNode {
	input := []rune(s)
	n := tr.root
	for i := 0; i < len(input); i++ {
		c := input[i]
		if !n.hasRecord(c) {
			return nil
		}

		n = n.getRecord(c)
	}

	if n.end {
		return n
	}

	return nil
}

func main() {
	tree := newTrie()
	tests := []string{
		"/first",
		"/first/one",
		"/first/one/two",
		"/second",
		"/second/one",
		"/second/one/two/three",
	}

	for _, tt := range tests {
		tree.insert(tt)

		if n := tree.search(tt); n == nil {
			panic(fmt.Sprintf("expected %s to be found", tt))
		} else {
			// for _, record := range tree.root.records {
			// 	fmt.Printf("%s: %#+v\n", tt, record)
			// 	for _, childRecord := range record.records {
			// 		fmt.Printf("child: %#+v\n", childRecord)
			// 	}
			// }
		}

		if n := tree.search("/first/one/three"); n != nil {
			panic(fmt.Sprintf("expected to not be found"))
		}
		if n := tree.search("/first/one/"); n != nil {
			panic(fmt.Sprintf("expected to not be found"))
		}

		if n := tree.search("/f"); n != nil {
			panic(fmt.Sprintf("expected to not be found"))
		}
	}
}
