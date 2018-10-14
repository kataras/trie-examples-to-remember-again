package muxie

import (
	"net/http"
)

func GetParam(w http.ResponseWriter, key string) string {
	if store, ok := w.(*paramsWriter); ok {
		return store.Get(key)
	}

	return ""
}

func GetParams(w http.ResponseWriter) []ParamEntry {
	if store, ok := w.(*paramsWriter); ok {
		return store.params
	}

	return nil
}

func SetParam(w http.ResponseWriter, key, value string) bool {
	if store, ok := w.(*paramsWriter); ok {
		store.Set(key, value)
		return true
	}

	return false
}

type paramsWriter struct {
	http.ResponseWriter
	params []ParamEntry
}

type ParamEntry struct {
	Key   string
	Value string
}

func (pw *paramsWriter) Set(key, value string) {
	if ln := len(pw.params); cap(pw.params) > ln {
		pw.params = pw.params[:ln+1]
		p := &pw.params[ln]
		p.Key = key
		p.Value = value
		return
	}

	pw.params = append(pw.params, ParamEntry{
		Key:   key,
		Value: value,
	})
}

func (pw *paramsWriter) Get(key string) string {
	n := len(pw.params)
	for i := 0; i < n; i++ {
		if kv := pw.params[i]; kv.Key == key {
			return kv.Value
		}
	}

	return ""
}

func (pw *paramsWriter) reset(w http.ResponseWriter) {
	pw.ResponseWriter = w
	pw.params = pw.params[0:0]
}
