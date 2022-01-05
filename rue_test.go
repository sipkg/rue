package rue

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var tests = []struct {
	RouteMethod  string
	RoutePattern string

	Method string
	Path   string
	Match  bool
	Params map[string]string
}{
	// simple path matching
	{
		"GET", "/one",
		"GET", "/one", true, nil,
	},
	{
		"GET", "/two",
		"GET", "/two", true, nil,
	},
	{
		"GET", "/three",
		"GET", "/three", true, nil,
	},
	// methods
	{
		"get", "/methodcase",
		"GET", "/methodcase", true, nil,
	},
	{
		"Get", "/methodcase",
		"get", "/methodcase", true, nil,
	},
	{
		"GET", "/methodcase",
		"get", "/methodcase", true, nil,
	},
	{
		"GET", "/method1",
		"POST", "/method1", false, nil,
	},
	{
		"DELETE", "/method2",
		"GET", "/method2", false, nil,
	},
	{
		"GET", "/method3",
		"PUT", "/method3", false, nil,
	},
	// all methods
	{
		"*", "/all-methods",
		"GET", "/all-methods", true, nil,
	},
	{
		"*", "/all-methods",
		"POST", "/all-methods", true, nil,
	},
	{
		"*", "/all-methods",
		"PUT", "/all-methods", true, nil,
	},
	// nested
	{
		"GET", "/parent/child/one",
		"GET", "/parent/child/one", true, nil,
	},
	{
		"GET", "/parent/child/two",
		"GET", "/parent/child/two", true, nil,
	},
	{
		"GET", "/parent/child/three",
		"GET", "/parent/child/three", true, nil,
	},
	// slashes
	{
		"GET", "slashes/one",
		"GET", "/slashes/one", true, nil,
	},
	{
		"GET", "/slashes/two",
		"GET", "slashes/two", true, nil,
	},
	{
		"GET", "slashes/three/",
		"GET", "/slashes/three", true, nil,
	},
	{
		"GET", "/slashes/four",
		"GET", "slashes/four/", true, nil,
	},
	// prefix
	{
		"GET", "/prefix/",
		"GET", "/prefix/anything/else", true, nil,
	},
	{
		"GET", "/not-prefix",
		"GET", "/not-prefix/anything/else", false, nil,
	},
	{
		"GET", "/prefixdots...",
		"GET", "/prefixdots/anything/else", true, nil,
	},
	{
		"GET", "/prefixdots...",
		"GET", "/prefixdots", true, nil,
	},
	// path params
	{
		"GET", "/path-param/:id",
		"GET", "/path-param/123", true, map[string]string{"id": "123"},
	},
	{
		"GET", "/path-params/:era/:group/:member",
		"GET", "/path-params/60s/beatles/lennon", true, map[string]string{
			"era":    "60s",
			"group":  "beatles",
			"member": "lennon",
		},
	},
	{
		"GET", "/path-params-prefix/:era/:group/:member/",
		"GET", "/path-params-prefix/60s/beatles/lennon/yoko", true, map[string]string{
			"era":    "60s",
			"group":  "beatles",
			"member": "lennon",
		},
	},
	// misc no matches
	{
		"GET", "/not/enough",
		"GET", "/not/enough/items", false, nil,
	},
	{
		"GET", "/not/enough/items",
		"GET", "/not/enough", false, nil,
	},
}

func TestRue(t *testing.T) {
	for _, test := range tests {
		r := NewRouter()
		match := false
		// var ctx context.Context
		r.Handle(test.RouteMethod, test.RoutePattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			match = true
			// ctx = r.Context()
		}))
		req, err := http.NewRequest(test.Method, test.Path, nil)
		if err != nil {
			t.Errorf("NewRequest: %s", err)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if match != test.Match {
			t.Errorf("expected match %v but was %v: %s %s", test.Match, match, test.Method, test.Path)
		}
		if len(test.Params) > 0 {
			for expK, expV := range test.Params {
				// check using helper
				actualValStr := Param(req, expK)
				if actualValStr != expV {
					t.Errorf("Param: context value %s expected \"%s\" but was \"%s\"", expK, expV, actualValStr)
				}
			}
		}
	}
}

func TestMultipleRoutesDifferentMethods(t *testing.T) {
	r := NewRouter()
	var match string
	r.Handle(http.MethodGet, "/route", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		match = "GET /route"
	}))
	r.Handle(http.MethodDelete, "/route", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		match = "DELETE /route"
	}))
	r.Handle(http.MethodPost, "/route", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		match = "POST /route"
	}))

	req, err := http.NewRequest(http.MethodGet, "/route", nil)
	if err != nil {
		t.Errorf("NewRequest: %s", err)
	}
	r.ServeHTTP(httptest.NewRecorder(), req)
	if match != "GET /route" {
		t.Errorf("unexpected: %s", match)
	}

	req, err = http.NewRequest(http.MethodDelete, "/route", nil)
	if err != nil {
		t.Errorf("NewRequest: %s", err)
	}
	r.ServeHTTP(httptest.NewRecorder(), req)
	if match != "DELETE /route" {
		t.Errorf("unexpected: %s", match)
	}

	req, err = http.NewRequest(http.MethodPost, "/route", nil)
	if err != nil {
		t.Errorf("NewRequest: %s", err)
	}
	r.ServeHTTP(httptest.NewRecorder(), req)
	if match != "POST /route" {
		t.Errorf("unexpected: %s", match)
	}
}

func TestGetFormParameters(t *testing.T) {
	req, err := http.NewRequest("GET", "/happy/s?q=foo&z=bar", nil)
	if err != nil {
		t.Errorf("NewRequest: %s", err)
	}
	r := NewRouter()
	var h, q, z string
	r.Handle(http.MethodGet, "/:h/s", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h = Param(r, "h")
		q = Param(r, "q")
		z = Param(r, "z")
	}))
	r.ServeHTTP(httptest.NewRecorder(), req)
	if h != "happy" {
		t.Errorf(`Param("h") = %q, want "happy"`, h)
	}
	if q != "foo" {
		t.Errorf(`Param("q") = %q, want "foo"`, q)
	}
	if z != "bar" {
		t.Errorf(`Param("z") = %q, want "bar"`, z)
	}
}

func TestPostFormParameters(t *testing.T) {
	req, err := http.NewRequest("POST",
		"/happy/s?q=foo",
		strings.NewReader("z=bar"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	if err != nil {
		t.Errorf("NewRequest: %s", err)
	}
	r := NewRouter()
	var h, q, z string
	r.Handle(http.MethodPost, "/:h/s", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h = Param(r, "h")
		q = Param(r, "q")
		z = Param(r, "z")
	}))
	r.ServeHTTP(httptest.NewRecorder(), req)
	if h != "happy" {
		t.Errorf(`Param("h") = %q, want "happy"`, h)
	}
	if q != "foo" {
		t.Errorf(`Param("q") = %q, want "foo"`, q)
	}
	if z != "bar" {
		t.Errorf(`Param("z") = %q, want "bar"`, z)
	}
}

func TestGetHostFQDN(t *testing.T) {
	req, err := http.NewRequest("GET", "http://somehost.domain.tld/hello", nil)
	if err != nil {
		t.Errorf("NewRequest: %s", err)
	}
	r := NewRouter()
	var h, f string
	r.HandleFunc("GET", "/hello", func(w http.ResponseWriter, r *http.Request) {
		h = Param(r, "_host")
		f = Param(r, "_fqdn")
	})
	r.ServeHTTP(httptest.NewRecorder(), req)
	if h != "somehost" {
		t.Errorf(`Param("h") = %q, want "somehost"`, h)
	}
	if f != "somehost.domain.tld" {
		t.Errorf(`Param("f") = %q, want "somehost.domain.tld"`, f)
	}
}

//
// B E N C H M A R K S :
//

// data needed for benchmarking
type benchmarkData struct {
	RouteMethod  string
	RoutePattern string

	Method string
	Path   string
}

// only needed to prevent the compiler from optimizing the code too much
var benchmarkCounter int

// dummy routes used to fill up the routes table and make the benchmark more interesting
var benchmarkDummyRoutes = []string{
	// don't match at all:
	"/nomatch/one",
	"/nomatch/two",
	"/nomatch/three",
	"/nomatch/four",
	"/nomatch/five",
	"/nomatch/six",
	"/nomatch/seven",
	"/nomatch/eight",
	"/nomatch/nine",
	"/nomatch/ten",

	// matching one:
	"/one/nomatch/one",
	"/one/nomatch/two",
	"/one/nomatch/three",
	"/one/nomatch/four",
	"/one/nomatch/five",
	"/one/nomatch/six",
	"/one/nomatch/seven",
	"/one/nomatch/eight",
	"/one/nomatch/nine",
	"/one/nomatch/ten",

	// matching two:
	"/one/two/nomatch/one",
	"/one/two/nomatch/two",
	"/one/two/nomatch/three",
	"/one/two/nomatch/four",
	"/one/two/nomatch/five",
	"/one/two/nomatch/six",
	"/one/two/nomatch/seven",
	"/one/two/nomatch/eight",
	"/one/two/nomatch/nine",
	"/one/two/nomatch/ten",

	// matching three:
	"/one/two/three/nomatch/one",
	"/one/two/three/nomatch/two",
	"/one/two/three/nomatch/three",
	"/one/two/three/nomatch/four",
	"/one/two/three/nomatch/five",
	"/one/two/three/nomatch/six",
	"/one/two/three/nomatch/seven",
	"/one/two/three/nomatch/eight",
	"/one/two/three/nomatch/nine",
	"/one/two/three/nomatch/ten",
}

func BenchmarkStaticTwo(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/one/two",
		"GET", "/one/two",
	})
}
func BenchmarkStaticThree(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/one/two/three",
		"GET", "/one/two/three",
	})
}
func BenchmarkStaticFour(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/one/two/three/four",
		"GET", "/one/two/three/four",
	})
}
func BenchmarkParamsTwo(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/:one/:two",
		"GET", "/one/two",
	})
}
func BenchmarkParamsThree(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/:one/:two/:three",
		"GET", "/one/two/three",
	})
}
func BenchmarkParamsFour(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/:one/:two/:three/:four",
		"GET", "/one/two/three/four",
	})
}
func BenchmarkNoMatchStaticThree(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/one/two/three/four",
		"GET", "/one/two/three",
	})
}
func BenchmarkNoMatchParamsThree(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/:one/:two/:three/:four",
		"GET", "/one/two/three",
	})
}
func BenchmarkNoMatchStaticFour(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/one/two/three/four",
		"GET", "/one/two/three",
	})
}
func BenchmarkNoMatchParamsFour(b *testing.B) {
	benchmarkWay(b, benchmarkData{
		"GET", "/:one/:two/:three/:four",
		"GET", "/one/two/three",
	})
}

func benchmarkWay(b *testing.B, d benchmarkData) {
	r := NewRouter()
	count := 0
	for _, pattern := range benchmarkDummyRoutes {
		r.Handle(d.RouteMethod, pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b.Errorf("Routed to wrong method in benchmark!")
		}))
	}
	r.Handle(d.RouteMethod, d.RoutePattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
	}))
	req, err := http.NewRequest(d.Method, d.Path, nil)
	if err != nil {
		b.Errorf("NewRequest: %s", err)
	}
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
	benchmarkCounter = count
}
