package retina

import (
	"net/http"
	"sort"
	"strings"
)

type Node struct {
	parent *Node

	children               map[string]*Node
	hasDynamicChild        bool // does one of the children contains a parameter or wildcard?
	childNamedParameter    bool // is the child a named parameter (single segmnet)
	childWildcardParameter bool // or it is a wildcard (can be more than one path segments) ?

	paramKeys []string // the param keys without : or *.
	end       bool     // it is a complete node, here we stop and we can say that the node is valid.
	key       string   // if end == true then key is filled with the original value of the insertion's key.
	// if key != "" && its parent has childWildcardParameter == true,
	// we need it to track the static part for the closest-wildcard's parameter storage.
	staticKey string

	// insert main data relative to http and a tag for things like route names.
	Handler http.Handler
	Tag     string

	// other insert data.
	Data interface{}
}

func newNode() *Node {
	n := new(Node)
	return n
}

func (n *Node) addChild(s string, child *Node) {
	if n.children == nil {
		n.children = make(map[string]*Node)
	}

	if _, exists := n.children[s]; exists {
		return
	}

	child.parent = n
	n.children[s] = child
}

func (n *Node) getChild(s string) *Node {
	if n.children == nil {
		n.children = make(map[string]*Node)
	}

	return n.children[s]
}

func (n *Node) hasChild(s string) bool {
	return n.getChild(s) != nil
}

func (n *Node) findClosestParentWildcardNode() *Node {
	n = n.parent
	for n != nil {
		if n.childWildcardParameter {
			return n.getChild(WildcardParamStart)
		}

		n = n.parent
	}

	return nil
}

// NodeKeysSorter is the type definition for the sorting logic
// that caller can pass on `GetKeys` and `Autocomplete`.
type NodeKeysSorter = func(list []string) func(i, j int) bool

// DefaultKeysSorter sorts as: first the "key (the path)" with the lowest number of slashes.
var DefaultKeysSorter = func(list []string) func(i, j int) bool {
	return func(i, j int) bool {
		return len(strings.Split(list[i], pathSep)) < len(strings.Split(list[j], pathSep))
	}
}

func (n *Node) Keys(sorter NodeKeysSorter) (list []string) {
	if n.end {
		list = append(list, n.key)
	}

	if n.children != nil {
		for _, child := range n.children {
			list = append(list, child.Keys(sorter)...)
		}
	}

	if sorter != nil {
		sort.Slice(list, sorter(list))
	}

	return
}

func (n *Node) Parent() *Node {
	return n.parent
}

func (n *Node) String() string {
	return n.key
}

func (n *Node) IsEnd() bool {
	return n.end
}
