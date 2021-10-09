# Usage

```go
package main

import (
	"net/http"

	"github.com/sipkg/rue"
)

func main() {
	router := rue.NewRouter()
	router.HandleFunc("GET", "/path/:para1/:para2", handle)
	router.HandleFunc("POST", "/path/:para1/:para2", handle)
	panic(http.ListenAndServe(":8080", router))
}

// curl --data 'post=foo&multi=one' http://host.localhost:8080/path/foo/bar\?get=\bar\&multi\=two\&multi\=three

func handle(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	params := []string{"para1", "para2", "multi", "post", "get", "_host"}
	data := ""
	for _, p := range params {
		data = data + p + ": \t" + rue.Param(c, p) + "\n"
	}

	w.Write([]byte(data))
}
```