package retina

import (
	"html"
	"io"
	"net/http"
	"strings"
	"sync"
)

type Mux struct {
	PathCorrection bool
	Routes         *Trie

	paramsPool *sync.Pool
	root       string

	// TODO: somehow make the separator to be able to chagne by mux or by some options, configs...
}

func NewMux() *Mux {
	return &Mux{
		Routes: NewTrie(),
		paramsPool: &sync.Pool{
			New: func() interface{} {
				return &paramsWriter{}
			},
		},
		root: "",
	}
}

func (m *Mux) Handle(pattern string, handler http.Handler) {
	m.Routes.Insert(m.root+pattern, WithHandler(handler))
}

func (m *Mux) HandleFunc(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	m.Handle(pattern, http.HandlerFunc(handlerFunc))
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if m.PathCorrection {
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			// Remove trailing slash and client-permanent rule for redirection,
			// if confgiuration allows that and path has an extra slash.

			// update the new path and redirect.
			// use Trim to ensure there is no open redirect due to two leading slashes
			r.URL.Path = pathSep + strings.Trim(path, pathSep)
			url := r.URL.String()
			method := r.Method
			// Fixes https://github.com/kataras/iris/issues/921
			// This is caused for security reasons, imagine a payment shop,
			// you can't just permantly redirect a POST request, so just 307 (RFC 7231, 6.4.7).
			if method == http.MethodPost || method == http.MethodPut {
				http.Redirect(w, r, url, http.StatusTemporaryRedirect)
				return
			}
			http.Redirect(w, r, url, http.StatusMovedPermanently)

			// RFC2616 recommends that a short note "SHOULD" be included in the
			// response because older user agents may not understand 301/307.
			// Shouldn't send the response for POST or HEAD; that leaves GET.
			if method == http.MethodGet {
				io.WriteString(w, "<a href=\""+html.EscapeString(url)+"\">Moved Permanently</a>.\n")
			}
			return
		}
	}

	// r.URL.Query() is slow and will allocate a lot, although
	// the first idea was to not introduce a new type to the end-developers
	// so they are using this library as the std one, but we will have to do it
	// for the params, we keep that rule so a new ResponseWriter, which is an interface,
	// and it will be compatible with net/http will be introduced to store the params at least,
	// we don't want to add a third parameter or a global state to this library.

	pw := m.paramsPool.Get().(*paramsWriter)
	pw.reset(w)
	n := m.Routes.Search(path, pw)
	if n != nil && n.Handler != nil {
		n.Handler.ServeHTTP(pw, r)
	}

	m.paramsPool.Put(pw)
}

type SubMux interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handlerFunc func(http.ResponseWriter, *http.Request))
	Of(prefix string) SubMux
}

func (m *Mux) Of(prefix string) SubMux {
	if prefix == "" || prefix == pathSep {
		return m
	}

	if prefix == m.root {
		return m
	}

	// modify prefix if it's already there on the parent.
	if strings.HasPrefix(m.root, prefix) {
		prefix = prefix[0:strings.LastIndex(m.root, prefix)]
	}

	// remove last slash "/", if any.
	if lidx := len(prefix) - 1; prefix[lidx] == pathSepB {
		prefix = prefix[0:lidx]
	}

	// remove any duplication of slashes "/".
	prefix = pathSep + strings.Trim(m.root+prefix, pathSep)

	return &Mux{
		Routes: m.Routes,
		root:   prefix,
	}
}
