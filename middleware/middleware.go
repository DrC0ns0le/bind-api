package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func CorsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, Depth, User-Agent, X-File-Size, X-Requested-With, If-Modified-Since, X-File-Name, Cache-Control")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length,Content-Range")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RESTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		buf := new(bytes.Buffer)

		forwardedFor := r.Header.Get("X-Forwarded-For")
		forwardedPort := r.Header.Get("X-Forwarded-Port")

		if forwardedFor == "" {
			forwardedFor = r.RemoteAddr
		}

		fmt.Fprintf(buf, "%s %s %s", net.JoinHostPort(forwardedFor, forwardedPort), r.Method, r.URL.Path)

		if r.Method != "GET" && r.Header.Get("Content-Type") == "application/json" {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
			} else {
				fmt.Fprintf(buf, " %s", string(bodyBytes))
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		rbw := &responseBodyWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			status:         http.StatusOK,
		}

		next.ServeHTTP(rbw, r)

		fmt.Fprintf(buf, " %d", rbw.status)

		// If status is not 2xx, log additional information
		if rbw.status >= 300 || rbw.status < 200 {
			fmt.Fprintf(buf, " %s", rbw.body.String())
		}

		fmt.Fprintf(buf, " %s", time.Since(start))

		log.Println(buf.String())
	})
}

// customResponseWriter wraps http.ResponseWriter to capture the status code
type responseBodyWriter struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseBodyWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func HealthCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
