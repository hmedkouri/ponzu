package api

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/hmedkouri/ponzu/system/db"
)

// Gzip wraps a HandlerFunc to compress responses when possible
func Gzip(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if db.ConfigCache("gzip_disabled").(bool) == true {
			next.ServeHTTP(res, req)
			return
		}

		// check if req header content-encoding supports gzip
		if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			// gzip response data
			res.Header().Set("Content-Encoding", "gzip")
			gzWriter := gzip.NewWriter(res)
			defer gzWriter.Close()
			var gzres gzipResponseWriter
			if pusher, ok := res.(http.Pusher); ok {
				gzres = gzipResponseWriter{res, pusher, gzWriter}
			} else {
				gzres = gzipResponseWriter{res, nil, gzWriter}
			}

			next.ServeHTTP(gzres, req)
			return
		}

		next.ServeHTTP(res, req)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	pusher http.Pusher

	gw *gzip.Writer
}

func (gzw gzipResponseWriter) Write(p []byte) (int, error) {
	return gzw.gw.Write(p)
}

func (gzw gzipResponseWriter) Push(target string, opts *http.PushOptions) error {
	if gzw.pusher == nil {
		return nil
	}

	if opts == nil {
		opts = &http.PushOptions{
			Header: make(http.Header),
		}
	}

	opts.Header.Set("Accept-Encoding", "gzip")

	return gzw.pusher.Push(target, opts)
}
