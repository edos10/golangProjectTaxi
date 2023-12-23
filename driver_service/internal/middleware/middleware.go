package middleware

import (
	"fmt"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Printf("Request: %s %s %s\n", r.Method, r.RequestURI, start.Format(time.RFC822))
		defer fmt.Printf("Response: %s %s %s %s\n", r.Method, r.RequestURI, time.Now().Format(time.RFC822), time.Since(start))

		next.ServeHTTP(w, r)
	})
}
