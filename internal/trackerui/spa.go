/*
MIT License
-----------

Copyright (c) 2020 Steve McDaniel

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the "Software"), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
package trackerui
*/

package trackerui

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type HttpdHandler struct {
	root http.FileSystem
	fs   http.Handler
}

func Httpd(root http.Dir) *HttpdHandler {
	return &HttpdHandler{
		root: root,
		fs:   http.FileServer(root),
	}
}

func cut(name string) string {
	name = strings.TrimSuffix(name, "/")
	dir, _ := path.Split(name)
	return dir
}

func HttpServeSPA(listen string, path string) {
	var srv *http.Server

	httpdServer := Httpd(http.Dir(path))

	srv = &http.Server{Addr: listen, Handler: httpdServer.fs}
	log.Printf("Starting HTTPD on %s", listen)
	/*
	   err := srv.ListenAndServeTLS("server.crt","server.key")
	*/
	err := srv.ListenAndServe()

	if err != nil {
		log.Printf("ListenAndServe failed: %s\n", err)
		os.Exit(1)
	}
}

func (dh *HttpdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}

	f, err := dh.root.Open(upath)
	if err == nil {
		f.Close()
	}

	if err != nil && os.IsNotExist(err) {
		for upath != "/" {
			upath = cut(upath)
			f, err := dh.root.Open(path.Join(upath, "index.html"))
			switch {
			case err == nil:
				f.Close()
				fallthrough
			case !os.IsNotExist(err):
				r.URL.Path = upath
				dh.fs.ServeHTTP(w, r)
				return
			}
		}
	}
	dh.fs.ServeHTTP(w, r)
}
