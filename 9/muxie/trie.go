package muxie

import (
	"net/http"
	"strings"
)

const (
	ParamStart         = ":" // is segment a named parameter?
	WildcardParamStart = "*" // allow everything else after that path prefix but it checks for static paths and named parameters before that in order to support everything that other implementations do not.
)

type Trie struct {
	root *Node

	// if true then it will handle any path if not other parent wildcard exists,
	// so even 404 (on http services) is up to it, see Trie#Insert.
	hasRootWildcard bool
}

func NewTrie() *Trie {
	return &Trie{
		root:            NewNode(),
		hasRootWildcard: false,
	}
}

type InsertOption func(*Node)

func WithHandler(handler http.Handler) InsertOption {
	return func(n *Node) {
		if n.Handler == nil {
			n.Handler = handler
		}
	}
}

func WithTag(tag string) InsertOption {
	return func(n *Node) {
		if n.Tag == "" {
			n.Tag = tag
		}
	}
}

func WithData(data interface{}) InsertOption {
	return func(n *Node) {
		// data can be replaced.
		n.Data = data
	}
}

func (t *Trie) Insert(key string, options ...InsertOption) {
	n := t.insert(key, "", nil, nil)
	for _, opt := range options {
		opt(n)
	}
}

func (t *Trie) InsertRoute(pattern, routeName string, handler http.Handler) {
	t.insert(pattern, routeName, nil, handler)
}

const (
	pathSep  = "/"
	pathSepB = '/'
)

func slowPathSplit(path string) []string {
	if path == pathSep {
		return []string{pathSep}
	}

	// remove last sep if any.
	if path[len(path)-1] == pathSepB {
		path = path[:len(path)-1]
	}

	return strings.Split(path, pathSep)[1:]
}

func resolveStaticPart(key string) string {
	i := strings.Index(key, ParamStart)
	if i == -1 {
		i = strings.Index(key, WildcardParamStart)
	}
	if i == -1 {
		i = len(key)
	}

	return key[:i]
}

func (t *Trie) insert(key, tag string, optionalData interface{}, handler http.Handler) *Node {
	input := slowPathSplit(key)

	n := t.root
	var paramKeys []string

	for _, s := range input {
		c := s[0]

		if isParam, isWildcard := c == ParamStart[0], c == WildcardParamStart[0]; isParam || isWildcard {
			n.hasDynamicChild = true
			paramKeys = append(paramKeys, s[1:]) // without : or *.

			// if node has already a wildcard, don't force a value, check for true only.
			if isParam {
				n.childNamedParameter = true
				s = ParamStart
			}

			if isWildcard {
				n.childWildcardParameter = true
				s = WildcardParamStart
				if t.root == n {
					t.hasRootWildcard = true
				}
			}
		}

		if !n.hasChild(s) {
			child := NewNode()
			n.addChild(s, child)
		}

		n = n.getChild(s)
	}

	n.Tag = tag
	n.Handler = handler
	n.Data = optionalData

	n.paramKeys = paramKeys
	n.key = key
	n.staticKey = resolveStaticPart(key)
	n.end = true

	return n
}

func (t *Trie) SearchPrefix(prefix string) *Node {
	input := slowPathSplit(prefix)
	n := t.root

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

func (t *Trie) HasPrefix(s string) bool {
	return t.SearchPrefix(s) != nil
}

func (t *Trie) Autocomplete(prefix string, sorter NodeKeysSorter) (list []string) {
	n := t.SearchPrefix(prefix)
	if n != nil {
		list = n.Keys(sorter)
	}
	return
}

type ParamsSetter interface {
	Set(string, string)
}

type Setter func(string, string)

func (fn Setter) Set(key, value string) {
	fn(key, value)
}

func (t *Trie) Search(q string, params ParamsSetter) *Node {
	end := len(q)
	n := t.root
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
				n = n.getChild(ParamStart)
				if ln := len(paramValues); cap(paramValues) > ln {
					paramValues = paramValues[:ln+1]
					paramValues[ln] = q[start:i]
				} else {
					paramValues = append(paramValues, q[start:i])
				}
			} else if n.childWildcardParameter {
				//	println("dynamic WILDCARD element for: " + q[start:i] + " found ")
				n = n.getChild(WildcardParamStart)
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
	}

	if n == nil || !n.end {
		if n != nil { // we need it on both places, on last segment (below) or on the first unnknown (above).
			if n = n.findClosestParentWildcardNode(); n != nil {
				params.Set(n.paramKeys[0], q[len(n.staticKey):])
				return n
			}
		}

		if t.hasRootWildcard {
			// that's the case for root wildcard, tests are passing
			// even without it but stick with it for reference.
			// Note ote that something like:
			// Routes: /other2/*myparam and /other2/static
			// Reqs: /other2/staticed will be handled
			// the /other2/*myparam and not the root wildcard, which is what we want.
			//
			n = t.root.getChild(WildcardParamStart)
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
