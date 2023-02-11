package multiclient

import "net/http"

type ClientPool interface {
	LazyClient(string) *http.Client
}
