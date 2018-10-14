package main

import (
	"fmt"

	"github.com/kataras/iris/context"
	"github.com/kataras/trie-examples-to-remember-again/9/muxie"
)

/* go get github.com/kataras/iris */

func main() {
	tree := muxie.NewTrie()

	tests := []struct {
		Path      string
		RouteName string
	}{
		{"/first", "first_data"},
		{"/first/one", "first/one_data"},
		{"/first/one/two", "first/one/two_data"},
		{"/firstt", "firstt_data"},
		{"/second", "second_data"},
		{"/second/one", "second/one_data"},
		{"/second/one/two", "second/one/two_data"},
		{"/second/one/two/three", "second/one/two/three_data"},

		// named parameters.
		{"/first/one/with/:param1/:param2/:param3/static", "first/one/with/static/_data_otherparams_with_static_end"},
		{"/first/one/with/:param1/:param2/:param3", "first/one/with/with_data_threeparams"},
		{"/first/one/with/:param/static/:otherparam", "first/one/with/static/_data_otherparam"},
		{"/first/one/with/:param", "first/one/with_data_param"},

		// wildcard named parameters.
		{"/second/wild/*mywildcardparam", "second/wildcard_1"},
		// no wildcard but same prefix.
		{"/second/wild/static", "second/no_wild"},
		// no wildcard, parameter instead with same prefix.
		{"/sectond/wild/:param", "second/no_wild_but_param"},
		// root wildcard.
		{"/*anything", "root_wildcard"},
	}

	for _, tt := range tests {
		tree.InsertRoute(tt.Path, tt.RouteName, nil)

		// if n := tree.Search(s); n == nil {
		// 	panic(fmt.Sprintf("expected %s to be found", s))
		// } else {
		// 	if expected, got := routeName, n.RouteName; expected != got {
		// 		panic(fmt.Sprintf("expected %s key to has route name: '%s' but got: '%s'", s, expected, got))
		// 	}
		// }
	}

	if tree.HasPrefix("/first/one/three") {
		panic(fmt.Sprintf("[0] expected to not has prefix"))
	}

	if !tree.HasPrefix("/first/one") {
		panic(fmt.Sprintf("[1] expected to has prefix"))
	}

	if !tree.HasPrefix("/first") {
		panic(fmt.Sprintf("[2] expected to has prefix"))
	}

	if !tree.HasPrefix("/second") {
		panic(fmt.Sprintf("[3] expected to has found"))
	}

	keyToTest := "/first"
	describe("autocomplete of \"%s\"", keyToTest)
	keyStartsWith1 := tree.Autocomplete(keyToTest, muxie.DefaultKeysSorter)
	for _, s := range keyStartsWith1 {
		fmt.Println(s)
	}

	keyToTest = "/second/one/two/three"
	describe("find final parents of \"%s\"", keyToTest)
	n := tree.SearchPrefix(keyToTest).Parent()
	for {
		if n == nil {
			break
		}

		if n.IsEnd() {
			fmt.Println(n.String())
		}

		n = n.Parent()
	}

	// n := tree.root
	params := new(context.RequestParams)

	keyToTest = "/first/one/with/myparam"
	describe("search all nodes against \"%s\"", keyToTest)

	if n = tree.Search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.String(), n.Tag)
		paramName := "param"
		if expected, got := "myparam", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()

	keyToTest = "/first/one/with/myparam1/static/myparam2"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.Search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param/static/:otherparam"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.String(), n.Tag)

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
	if n = tree.Search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param1/:param2/:param3"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.String(), n.Tag)

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
	if n = tree.Search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/first/one/with/:param1/:param2/:param3/static"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.String(), n.Tag)

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
	if n = tree.Search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/second/wild/*mywildcardparam"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.String(), n.Tag)

		paramName := "mywildcardparam"
		if expected, got := "everything/else/can/go/here", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()

	keyToTest = "/second/wild/static"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.Search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, keyToTest))
	} else {
		fmt.Printf("found: '%s': %s\n", n.String(), n.Tag)

		if params.Len() > 0 {
			panic("expected 0 parameters")
		}
	}
	params.Reset()

	keyToTest = "/sectond/wild/parameter1"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.Search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/sectond/wild/:param"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.String(), n.Tag)

		paramName := "param"
		if expected, got := "parameter1", params.Get(paramName); expected != got {
			panic(fmt.Sprintf("expected param: '%s' to be filled with value '%s' but got '%s' instead", paramName, expected, got))
		}
	}
	params.Reset()

	keyToTest = "/something/here/to/match/root/wildcard"
	describe("search all nodes against \"%s\"", keyToTest)
	if n = tree.Search(keyToTest, params); n == nil {
		panic(fmt.Sprintf("expected '%s' to be matched with: '%s' but nothing found\n", keyToTest, "/*anything"))
	} else {
		fmt.Printf("found: '%s': %s\n", n.String(), n.Tag)

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
