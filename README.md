# Rue
Minimalistic Go http router based on [Mat Ryer's way](https://github.com/matryer/way)

It extends way by adding all form parameters to the context and use the same function _rue.Param(parameter)_ to get a path or an url or a post parameter.

Multiple values for the same form parameter name are concatenate with the semi-colon separator. For example for a url query like _/path?m=a&m=b&m=c_ , a call to _rue.Param("m")_ will give _a;b;c_ as a result.

The request host prefix is also added to the context with the _host key. If the url is _something.domain.ltd_ , a call to _rue.Param("_host")_ will give _something_ as a result.

## Usage

Copy and paste the code below in a file, for example main.go :

```go
package main

import (
	"net/http"

	"github.com/sipkg/rue"
)

func main() {
	router := rue.NewRouter()
	router.HandleFunc("POST", "/path/:para1/:para2", handle)
    println("Listening on port 8080...")
	panic(http.ListenAndServe(":8080", router))
}

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

Initialize the go modules and execute the program with

```sh
go mod init rue_example && go mod tidy
go run .
```

You can simulate an http POST :

```sh
curl --data 'post=foo&multi=one' http://host.localhost:8080/path/foo/bar\?get=\bar\&multi\=two\&multi\=three
```