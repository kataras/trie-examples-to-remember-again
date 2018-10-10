package main

import (
	"testing"

	"github.com/kataras/iris/context"
)

// load.
type request struct {
	path string
}

var tests = []struct {
	key       string
	routeName string
	requests  []request
}{
	{"/first", "first_data", []request{ // 0
		{"/first"},
	}},
	{"/first/one", "first/one_data", []request{ // 1
		{"/first/one"},
	}},
	{"/first/one/two", "first/one/two_data", []request{ // 2
		{"/first/one/two"},
	}},
	{"/firstt", "firstt_data", []request{ // 3
		{"/firstt"},
	}},
	{"/second", "second_data", []request{ // 4
		{"/second"},
	}},
	{"/second/one", "second/one_data", []request{ // 5
		{"/second/one"},
	}},
	{"/second/one/two", "second/one/two_data", []request{ // 6
		{"/second/one/two"},
	}},
	{"/second/one/two/three", "second/one/two/three_data", []request{ // 7
		{"/second/one/two/three"},
	}},

	// named parameters.
	{"/first/one/with/:param1/:param2/:param3/static", "first/one/with/static/_data_otherparams_with_static_end", []request{ // 8
		{"/first/one/with/myparam1/myparam2/myparam3/static"},
	}},
	{"/first/one/with/:param1/:param2/:param3", "first/one/with/with_data_threeparams", []request{ // 9
		{"/first/one/with/myparam1/myparam2/myparam3"},
	}},
	{"/first/one/with/:param/static/:otherparam", "first/one/with/static/_data_otherparam", []request{ // 10
		{"/first/one/with/myparam1/static/myotherparam"},
	}},
	{"/first/one/with/:param", "first/one/with_data_param", []request{ // 11
		{"/first/one/with/singleparam"},
	}},

	// wildcard named parameters.
	{"/second/wild/*mywildcardparam", "second/wildcard_1", []request{ // 12
		{"/second/wild/everything/else/can/go/here"},
	}},
	// no wildcard but same prefix.
	{"/second/wild/static", "second/no_wild", []request{ // 13
		{"/second/wild/static"},
	}},
	// no wildcard, parameter instead with same prefix.
	{"/second/wild/:param", "second/no_wild_but_param", []request{ // 14
		{"/second/wild/myparam"},
	}},

	// this is not possible ofc because of wildcard, we support param, wildcard and static
	// in the same path but we don't have a way to check the next children of an unknnown segment,
	// and for the best of all.
	// {"/second/wild/:param/static", "second/with_param_and_static_should_fail", []request{ // 14
	// 	{"/second/wild/myparam/static", false, nil},
	// }},
	// root wildcard.
	{"/*anything", "root_wildcard", []request{ // 15
		{"/something/or/anything/can/be/stored/here"},
		{"/justsomething"},
	}},
}

func initTree(tree *trie) {
	for _, tt := range tests {
		tree.insert(tt.key, tt.routeName, nil)
	}
}

// BenchmarkTrieInsert and BenchmarkTrieSearch runs benchmarks against the trie implementation with map[string]*trieNode for children.

// go test -run=XXX -v -bench=BenchmarkTrieInsert -count=3
func BenchmarkTrieInsert(b *testing.B) {
	tree := newTrie()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		initTree(tree)
	}
}

// go test -run=XXX -v -bench=BenchmarkTrieSearch -count=3
func BenchmarkTrieSearch(b *testing.B) {
	tree := newTrie()
	initTree(tree)
	params := new(context.RequestParams)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := range tests {
			for _, req := range tests[i].requests {
				n := tree.search(req.path, params)
				if n == nil {
					b.Fatalf("%s: node not found\n", req.path)
				}
				params.Store = params.Store[0:0]
			}
		}
	}
}
