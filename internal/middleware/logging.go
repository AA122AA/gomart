package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type LoggedResponse struct {
	statusCode int
	size       int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	lResp *LoggedResponse
}

func (lr *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lr.ResponseWriter.Write(b)
	lr.lResp.size += size
	return size, err
}

func (lr *loggingResponseWriter) WriteHeader(statusCode int) {
	lr.ResponseWriter.WriteHeader(statusCode)
	lr.lResp.statusCode = statusCode
}

func WithLogger(lg *zap.Logger) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lResp := LoggedResponse{
				statusCode: 0,
				size:       0,
			}

			lw := loggingResponseWriter{
				ResponseWriter: w,
				lResp:          &lResp,
			}

			next.ServeHTTP(&lw, r)

			duration := time.Since(start)

			lg.Info(
				"got request",
				zap.String("url", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status code", lResp.statusCode),
				zap.String("duration", duration.String()),
				zap.Int("size", lResp.size),
			)
		})
	}

}
