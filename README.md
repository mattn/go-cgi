go-cgi
======

Usage
-----

```go
#! go-cgi
package main

import (
	"fmt"
	"net/http"
	"net/http/cgi"
)

func main() {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Hello %s", r.FormValue("name"))
	}))
}
// vim: ft=go:
```

Authors
-------

Yasuhiro Matsumoto

License
-------

under the MIT License: http://mattn.mit-license.org/2014
