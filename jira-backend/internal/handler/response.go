package handler

import (
	"encoding/json"
	"net/http"
)

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type envelope struct {
	Links    map[string]string `json:"_links"`
	Data     interface{}       `json:"data"`
	Message  string            `json:"message"`
	Name     string            `json:"name"`
	PageInfo *PageInfo         `json:"pageInfo"`
	Status   bool              `json:"status"`
}

func encode(w http.ResponseWriter, statusCode int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(payload)
}

func writeData(w http.ResponseWriter, statusCode int, data interface{}) error {
	return encode(w, statusCode, envelope{
		Links:   map[string]string{"href": ""},
		Data:    data,
		Message: "OK",
		Status:  true,
	})
}

func writeDataPaged(w http.ResponseWriter, statusCode int, data interface{}, page *PageInfo) error {
	return encode(w, statusCode, envelope{
		Links:    map[string]string{"href": ""},
		Data:     data,
		Message:  "OK",
		PageInfo: page,
		Status:   true,
	})
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	_ = encode(w, statusCode, envelope{
		Links:   map[string]string{"href": ""},
		Message: message,
		Status:  false,
	})
}
