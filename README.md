# Rue
Minimalistic Go http router based on [Mat Ryer's way](https://github.com/matryer/way)

It differs from _way_ by adding the path parameters to the form values.
To retrieve them, use the rue.Param function :
```go
value := rue.Param(req, parameter)
```

For exampe, with the url pattern _/path/:param1_ :
```go
value := rue.Param(req,"param1")
```

You can also retrieve all form values in the same manner.

Multiple values are concatenate with the semi-colon separator. 
For example, for the url query _/path?m=a&m=b&m=c_ :
```go
result := rue.Param(req, "m")
fmt.Println(result) // a;b;c
```

The request host prefix is also added to the form values with the *_host* key.
For example, for a complete url like _https://something.domain.ltd/query_ :
```go
h := rue.Param(req, "_host")
fmt.Printf("The host is %s", h) // The host is something
```

## Examples

A minimal hello world handler with a name as path parameter :

```go
package main

import (
	"net/http"

	"github.com/sipkg/rue"
)

func main() {
	// create the router
	router := rue.NewRouter()
	// add the handler functions with the method and the parameter path
	router.HandleFunc("GET", "hello/:name",
		func(w http.ResponseWriter, r *http.Request) {
			message := "Hello " + rue.Param(r, "name")
			w.Write([]byte(message))
		})
	println("listening on port 8080...")
	panic(http.ListenAndServe(":8080", router))
}
```

A more complicated example with mixed origins. 

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
	params := []string{"para1", "para2", "multi", "post", "get", "_host"}
	data := ""
	for _, p := range params {
		data = data + p + ": \t" + rue.Param(r, p) + "\n"
	}

	w.Write([]byte(data))
}
```

Simulate an http POST with :
```sh
curl --data 'post=foo&multi=one' http://host.localhost:8080/path/foo/bar\?get=\bar\&multi\=two\&multi\=three
```

**Reminder** : Don't forget to inialize the go modules before running 
this examples. Someting like :

```sh
go mod init rue_example && go mod tidy && go run .
```
