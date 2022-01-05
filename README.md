# Rue

Minimalistic Go http router based on [Mat Ryer's way](https://github.com/matryer/way)

It differs from _way_ by adding the path parameters to the form values.
To retrieve them, use the rue.Param function :

```go
value := rue.Param(req, parameter)
```

For example, with the url pattern _/path/:param1_ :

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

The request FQDN (without port) and host prefix are also added to the form values with respectively the *_fqdn* and the *_host* key.
For example, for a complete url like _https://something.domain.tld/query_ :

```go
f := rue.Param(req, "_fqdn")
fmt.Printf("The FQDN is %s", f) // The FQDN is something.domain.tld
h := rue.Param(req, "_host")
fmt.Printf("The host is %s", h) // The host is something
```

Finally, you can also specify a static file folder to be served by the router :

```
router := rue.NewRouter()
router.HandleStatic("/", "./static/")
panic(http.ListenAndServe(":8080", router))
```

**Important rule : the static handler must be the last added handler to the router.**

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

## Usage

* Use `NewRouter` to make a new `Router`
* Call `Handle` and `HandleFunc` to add handlers.
* Call `HandleStatic` to add a file system handler.
* Specify HTTP method and path pattern for each route.
* Use `Param` function to get the path parameters from the context.

```go
func main() {
	router := rue.NewRouter()

	router.HandleFunc("GET", "/music/:band/:song", handleReadSong)
	router.HandleFunc("PUT", "/music/:band/:song", handleUpdateSong)
	router.HandleFunc("DELETE", "/music/:band/:song", handleDeleteSong)
	router.HandleStatic("/", "./static")

	panic(http.ListenAndServe(":8080", router))
}

func handleReadSong(w http.ResponseWriter, r *http.Request) {
	band := rue.Param(r, "band")
	song := rue.Param(r, "song")
	// use 'band' and 'song' parameters...
}
```

* Prefix matching

To match any path that has a specific prefix, use the `...` prefix indicator:

```go
func main() {
	router := rue.NewRouter()

	router.HandleFunc("GET", "/images...", handleImages)
	panic(http.ListenAndServe(":8080", router))
}
```

In the above example, the following paths will match:

* `/images`
* `/images/`
* `/images/one/two/three.jpg`

* Set `Router.NotFound` to handle 404 errors manually

```go
func main() {
	router := rue.NewRouter()
	
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "This is not the page you are looking for")
	})
	
	panic(http.ListenAndServe(":8080", router))
}
```