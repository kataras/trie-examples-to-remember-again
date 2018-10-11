package main

import (
	"strings"
	"testing"

	"github.com/kataras/iris/context"
)

func countParams(key string) int {
	return strings.Count(key, param) + strings.Count(key, "*")
}

// BenchmarkTrie runs a benchmark against the trie implementation with slices of children.
// TODO: same benchmark with different trie implementation vased on children with map[string]*trieNode instead (it should be even faster).
func TestTrie(t *testing.T) {
	// load.
	type request struct {
		path   string
		found  bool
		params map[string]string
	}

	tests := []struct {
		key       string
		routeName string
		requests  []request
	}{
		{"/first", "first_data", []request{ // 0
			{"/first", true, nil},
		}},
		{"/first/one", "first/one_data", []request{ // 1
			{"/first/one", true, nil},
		}},
		{"/first/one/two", "first/one/two_data", []request{ // 2
			{"/first/one/two", true, nil},
		}},
		{"/firstt", "firstt_data", []request{ // 3
			{"/firstt", true, nil},
		}},
		{"/second", "second_data", []request{ // 4
			{"/second", true, nil},
		}},
		{"/second/one", "second/one_data", []request{ // 5
			{"/second/one", true, nil},
		}},
		{"/second/one/two", "second/one/two_data", []request{ // 6
			{"/second/one/two", true, nil},
		}},
		{"/second/one/two/three", "second/one/two/three_data", []request{ // 7
			{"/second/one/two/three", true, nil},
		}},

		// named parameters.
		{"/first/one/with/:param1/:param2/:param3/static", "first/one/with/static/_data_otherparams_with_static_end", []request{ // 8
			{"/first/one/with/myparam1/myparam2/myparam3/static", true, map[string]string{
				"param1": "myparam1",
				"param2": "myparam2",
				"param3": "myparam3",
			}},
		}},
		{"/first/one/with/:param1/:param2/:param3", "first/one/with/with_data_threeparams", []request{ // 9
			{"/first/one/with/myparam1/myparam2/myparam3", true, map[string]string{
				"param1": "myparam1",
				"param2": "myparam2",
				"param3": "myparam3",
			}},
		}},
		{"/first/one/with/:param/static/:otherparam", "first/one/with/static/_data_otherparam", []request{ // 10
			{"/first/one/with/myparam1/static/myotherparam", true, map[string]string{
				"param":      "myparam1",
				"otherparam": "myotherparam",
			}},
		}},
		{"/first/one/with/:param", "first/one/with_data_param", []request{ // 11
			{"/first/one/with/singleparam", true, map[string]string{
				"param": "singleparam",
			}},
		}},

		// wildcard named parameters.
		{"/second/wild/*mywildcardparam", "second/wildcard_1", []request{ // 12
			{"/second/wild/everything/else/can/go/here", true, map[string]string{
				"mywildcardparam": "everything/else/can/go/here",
			}},
		}},
		// no wildcard but same prefix.
		{"/second/wild/static", "second/no_wild", []request{ // 13
			{"/second/wild/static", true, nil},
		}},
		// no wildcard, parameter instead with same prefix.
		{"/second/wild/:param", "second/no_wild_but_param", []request{ // 14
			{"/second/wild/myparam", true, map[string]string{
				"param": "myparam",
			}},
		}},
		// root wildcard.
		{"/*anything", "root_wildcard", []request{ // 15
			{"/something/or/anything/can/be/stored/here", true, map[string]string{
				"anything": "something/or/anything/can/be/stored/here",
			}},
			{"/justsomething", true, map[string]string{
				"anything": "justsomething",
			}},
		}},
	}

	tree := newTrie()
	// insert.
	for idx, tt := range tests {
		tree.insert(tt.key, tt.routeName, nil)
		for reqIdx, req := range tt.requests {
			if expected, got := countParams(tt.key), len(req.params); req.found && expected != got {
				t.Fatalf("before ran: [%d:%d]: registered parameters and expected parameters have not the same length, should be: %d but %d given", idx, reqIdx, expected, got)
			}
		}
	}

	// run.
	for idx, tt := range tests {
		params := new(context.RequestParams)
		for reqIdx, req := range tt.requests {
			params.Reset()
			n := tree.search(req.path, params)

			if req.found {
				if n == nil {
					t.Errorf("[%d:%d] expected node with key: %s and requested path: %s to be found", idx, reqIdx, tt.key, req.path)
					continue
				}

				if !n.isEnd() {
					t.Errorf("[%d:%d] expected node with key: %s and requested path: %s to be found (with end == true)", idx, reqIdx, tt.key, req.path)
					continue
				}
			}

			if !req.found && n != nil {
				t.Fatalf("[%s:%d:%d] expected node with key: %s to NOT be found for requested path: %s", tt.key, idx, reqIdx, tt.key, req.path)
			}

			if n != nil {
				if expected, got := tt.key, n.key; expected != got {
					t.Errorf("[%d:%d] %s:\n\texpected found node's key to be equal with: '%s' but got: '%s' instead", idx, reqIdx, req.path, expected, got)
				}
				if expected, got := n.RouteName, tt.routeName; expected != got {
					t.Errorf("[%s:%d:%d] %s:\n\texpected RouteName to be equal with: '%s' but got: '%s' instead", n.key, idx, reqIdx, req.path, expected, got)
				}

				if expected, got := len(req.params), len(params.Store); expected != got {
					t.Errorf("[%s:%d:%d] %s:\n\texpected request params length to be: %d  but got: %d instead", n.key, idx, reqIdx, req.path, expected, got)
				}

				if req.params != nil {
					for paramKey, expectedValue := range req.params {
						gotValue := params.Get(paramKey)
						if gotValue == "" {
							t.Errorf("[%s:%d:%d] %s:\n\texpected request param with key: '%s' to be found", n.key, idx, reqIdx, req.path, paramKey)
						}
						if expectedValue != gotValue {
							t.Errorf("[%s:%d:%d] %s:\n\texpected request param with key: '%s' to be equal with: '%s' but got: '%s' instead", n.key, idx, reqIdx, req.path, paramKey, expectedValue, gotValue)
						}
					}
				}
			}
		}
	}
}
