package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"hse-2026-golang-project/jira-backend/internal/handler"
	"hse-2026-golang-project/jira-backend/internal/reqid"
)

const (
	defaultResourceTimeout  = 5 * time.Second
	defaultAnalyticsTimeout = 15 * time.Second
)

func cors(allowedOrigin string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func requestLogging(log *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			id := req.Header.Get(reqid.HeaderName)
			if id == "" {
				id = reqid.New()
			}
			ctx := reqid.WithID(req.Context(), id)
			req = req.WithContext(ctx)
			w.Header().Set(reqid.HeaderName, id)

			log.WithFields(logrus.Fields{
				"request_id": id,
				"method":     req.Method,
				"path":       req.URL.Path,
				"query":      req.URL.RawQuery,
			}).Debug("incoming request")

			started := time.Now()
			response := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(response, req)

			entry := log.WithFields(logrus.Fields{
				"request_id":  id,
				"method":      req.Method,
				"path":        req.URL.Path,
				"status":      response.status,
				"duration_ms": time.Since(started).Milliseconds(),
			})
			switch {
			case response.status >= http.StatusInternalServerError:
				entry.Error("request completed")
			case response.status >= http.StatusBadRequest:
				entry.Warn("request completed")
			default:
				entry.Info("request completed")
			}
		})
	}
}

func withRequestTimeout(timeout time.Duration) mux.MiddlewareFunc {
	if timeout <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, cancel := context.WithTimeout(req.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func NewRouter(
	projectHandler *handler.ProjectHandler,
	issueHandler *handler.IssueHandler,
	graphHandler *handler.GraphHandler,
	systemHandler *handler.SystemHandler,
	allowedOrigin string,
	log *logrus.Logger,
	timeouts ...time.Duration,
) *mux.Router {
	resourceTimeout := defaultResourceTimeout
	analyticsTimeout := defaultAnalyticsTimeout
	if len(timeouts) > 0 && timeouts[0] > 0 {
		resourceTimeout = timeouts[0]
	}
	if len(timeouts) > 1 && timeouts[1] > 0 {
		analyticsTimeout = timeouts[1]
	}

	r := mux.NewRouter()
	r.Use(requestLogging(log))
	r.Use(cors(allowedOrigin))

	resource := r.NewRoute().Subrouter()
	resource.Use(withRequestTimeout(resourceTimeout))
	resource.HandleFunc("/api/v1/projects", projectHandler.GetAll).Methods("GET")
	resource.HandleFunc("/api/v1/projects/{id:[0-9]+}", projectHandler.Stat).Methods("GET")
	resource.HandleFunc("/api/v1/projects/{id:[0-9]+}", projectHandler.Delete).Methods("DELETE")

	resource.HandleFunc("/api/v1/connector/projects", projectHandler.GetCatalog).Methods("GET")
	resource.HandleFunc("/api/v1/connector/updateProject", projectHandler.Update).Methods("POST")

	resource.HandleFunc("/api/v1/issues", issueHandler.GetByProject).Methods("GET")
	resource.HandleFunc("/api/v1/resource/services", systemHandler.GetServices).Methods("GET")

	analytics := r.NewRoute().Subrouter()
	analytics.Use(withRequestTimeout(analyticsTimeout))
	analytics.HandleFunc("/api/v1/graph/make/{task:[0-9]+}", graphHandler.Make).Methods("POST")
	analytics.HandleFunc("/api/v1/graph/get/{task:[0-9]+}", graphHandler.Get).Methods("GET")
	analytics.HandleFunc("/api/v1/compare/{task:[0-9]+}", graphHandler.Compare).Methods("GET")
	analytics.HandleFunc("/api/v1/graph/delete", graphHandler.DeleteGraphs).Methods("DELETE")
	analytics.HandleFunc("/api/v1/isAnalyzed", graphHandler.IsAnalyzed).Methods("GET")

	return r
}
