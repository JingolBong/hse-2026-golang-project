package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const defaultBindPort = 8000

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type Link struct {
	Href string `json:"href"`
}

type envelope struct {
	Links    map[string]Link `json:"_links"`
	Data     interface{}     `json:"data"`
	Message  string          `json:"message"`
	Name     string          `json:"name"`
	PageInfo *PageInfo       `json:"pageInfo,omitempty"`
	Status   bool            `json:"status"`
}

type LinkBuilder struct {
	fallbackBaseURL string
}

func NewLinkBuilder(bindAddress string, bindPort int) *LinkBuilder {
	host := strings.TrimSpace(bindAddress)
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}
	if bindPort == 0 {
		bindPort = defaultBindPort
	}

	return &LinkBuilder{
		fallbackBaseURL: "http://" + net.JoinHostPort(host, strconv.Itoa(bindPort)),
	}
}

func defaultLinkBuilder() *LinkBuilder {
	return NewLinkBuilder("localhost", defaultBindPort)
}

func linkBuilderOrDefault(builders []*LinkBuilder) *LinkBuilder {
	if len(builders) > 0 && builders[0] != nil {
		return builders[0]
	}

	return defaultLinkBuilder()
}

func normalizeForwardedValue(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.Split(value, ",")
	return strings.TrimSpace(parts[0])
}

func requestScheme(r *http.Request) string {
	if proto := normalizeForwardedValue(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		return proto
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func requestHost(r *http.Request) string {
	if host := normalizeForwardedValue(r.Header.Get("X-Forwarded-Host")); host != "" {
		return host
	}
	return r.Host
}

func (b *LinkBuilder) baseURLFromRequest(r *http.Request) string {
	host := requestHost(r)
	if host == "" {
		return b.fallbackBaseURL
	}
	return requestScheme(r) + "://" + host
}

func (b *LinkBuilder) Rel(path string) Link {
	return Link{Href: b.fallbackBaseURL + "/" + strings.TrimLeft(path, "/")}
}

func (b *LinkBuilder) RelForRequest(r *http.Request, path string) Link {
	return Link{Href: b.baseURLFromRequest(r) + "/" + strings.TrimLeft(path, "/")}
}

func (b *LinkBuilder) Self(r *http.Request) Link {
	return b.RelForRequest(r, r.URL.RequestURI())
}

func (b *LinkBuilder) links(r *http.Request, relations map[string]string) map[string]Link {
	links := make(map[string]Link, len(relations)+1)
	links["self"] = b.Self(r)
	for name, path := range relations {
		links[name] = b.RelForRequest(r, path)
	}

	return links
}

func (b *LinkBuilder) ResourceLinks(r *http.Request) map[string]Link {
	return b.links(r, map[string]string{
		"projects": "/api/v1/projects",
		"issues":   "/api/v1/issues",
		"services": "/api/v1/resource/services",
	})
}

func (b *LinkBuilder) ConnectorLinks(r *http.Request) map[string]Link {
	return b.links(r, map[string]string{
		"connectorProjects": "/api/v1/connector/projects",
		"connectorUpdate":   "/api/v1/connector/updateProject",
		"projects":          "/api/v1/projects",
		"services":          "/api/v1/resource/services",
	})
}

func (b *LinkBuilder) AnalyticsLinks(r *http.Request) map[string]Link {
	return b.links(r, map[string]string{
		"graphGet":    "/api/v1/graph/get/{task}",
		"graphMake":   "/api/v1/graph/make/{task}",
		"graphDelete": "/api/v1/graph/delete",
		"isAnalyzed":  "/api/v1/isAnalyzed",
		"compare":     "/api/v1/compare/{task}",
		"projects":    "/api/v1/projects",
		"services":    "/api/v1/resource/services",
	})
}

func (b *LinkBuilder) ServiceLinks(r *http.Request) map[string]Link {
	return b.links(r, map[string]string{
		"projects":          "/api/v1/projects",
		"projectStat":       "/api/v1/projects/{id}",
		"issues":            "/api/v1/issues",
		"connectorProjects": "/api/v1/connector/projects",
		"connectorUpdate":   "/api/v1/connector/updateProject",
		"graphGet":          "/api/v1/graph/get/{task}",
		"graphMake":         "/api/v1/graph/make/{task}",
		"graphDelete":       "/api/v1/graph/delete",
		"isAnalyzed":        "/api/v1/isAnalyzed",
		"compare":           "/api/v1/compare/{task}",
	})
}

func encode(w http.ResponseWriter, statusCode int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(payload)
}

func BuildSelfLink(r *http.Request) map[string]Link {
	return map[string]Link{
		"self": defaultLinkBuilder().Self(r),
	}
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var urlErr *url.Error
	return errors.As(err, &urlErr) && errors.Is(urlErr.Err, context.DeadlineExceeded)
}

func writeRequestTimeout(w http.ResponseWriter, r *http.Request, links map[string]Link) {
	writeError(w, r, http.StatusRequestTimeout, "Request Timeout", links)
}

func responseName(r *http.Request) string {
	return "Jira Analyzer REST API " + r.Method + " " + r.URL.Path
}

func writeData(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}, links map[string]Link) error {
	return encode(w, statusCode, envelope{
		Links:   links,
		Data:    data,
		Message: "OK",
		Name:    responseName(r),
		Status:  true,
	})
}

func writeDataPaged(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}, page *PageInfo, links map[string]Link) error {
	return encode(w, statusCode, envelope{
		Links:    links,
		Data:     data,
		Message:  "OK",
		Name:     responseName(r),
		PageInfo: page,
		Status:   true,
	})
}

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, message string, links ...map[string]Link) {
	responseLinks := BuildSelfLink(r)
	if len(links) > 0 {
		responseLinks = links[0]
	}

	_ = encode(w, statusCode, envelope{
		Links:   responseLinks,
		Message: message,
		Name:    responseName(r),
		Status:  false,
	})
}
