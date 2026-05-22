package app

import (
	"net/http"

	"github.com/gorilla/mux"

	"hse-2026-golang-project/jira-backend/internal/handler"
)

func cors(allowedOrigin string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

func NewRouter(
	projectHandler *handler.ProjectHandler,
	issueHandler *handler.IssueHandler,
	graphHandler *handler.GraphHandler,
	allowedOrigin string,
) *mux.Router {
	r := mux.NewRouter()
	r.Use(cors(allowedOrigin))

	r.HandleFunc("/api/v1/projects", projectHandler.GetCatalog).Methods("GET")
	r.HandleFunc("/api/v1/myprojects", projectHandler.GetAll).Methods("GET")
	r.HandleFunc("/api/v1/myprojects/{id:[0-9]+}/stat", projectHandler.Stat).Methods("GET")
	r.HandleFunc("/api/v1/projects/{id:[0-9]+}", projectHandler.Delete).Methods("DELETE")
	r.HandleFunc("/api/v1/projects/{project}/update", projectHandler.Update).Methods("POST")

	r.HandleFunc("/api/v1/issues", issueHandler.GetByProject).Methods("GET")
	r.HandleFunc("/api/v1/projects/{project}/issues", issueHandler.GetByProject).Methods("GET")

	r.HandleFunc("/api/v1/graph", graphHandler.Get).Methods("GET")
	r.HandleFunc("/api/v1/graph", graphHandler.Make).Methods("POST")
	r.HandleFunc("/api/v1/graph/compare", graphHandler.Compare).Methods("GET")

	r.HandleFunc("/api/v1/graph/make/{task:[0-9]+}", graphHandler.Make).Methods("POST")
	r.HandleFunc("/api/v1/graph/get/{task:[0-9]+}", graphHandler.Get).Methods("GET")
	r.HandleFunc("/api/v1/projects/{project}/graph/{task:[0-9]+}", graphHandler.Get).Methods("GET")
	r.HandleFunc("/api/v1/projects/{project}/graph/{task:[0-9]+}", graphHandler.Make).Methods("POST")
	r.HandleFunc("/api/v1/projects/{project}/graph/{task:[0-9]+}/make", graphHandler.Make).Methods("POST")
	r.HandleFunc("/api/v1/isAnalyzed", graphHandler.IsAnalyzed).Methods("GET")
	r.HandleFunc("/api/v1/projects/{project}/analyzed", graphHandler.IsAnalyzed).Methods("GET")
	r.HandleFunc("/api/v1/isEmpty", graphHandler.IsEmpty).Methods("GET")
	r.HandleFunc("/api/v1/graphs", graphHandler.DeleteGraphs).Methods("DELETE")

	return r
}
