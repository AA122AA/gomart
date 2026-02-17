package middleware

import "net/http"

type Middleware func(next http.Handler) http.HandlerFunc

func Wrap(h http.Handler, mw Middleware) http.HandlerFunc {
	return mw(h)
}
