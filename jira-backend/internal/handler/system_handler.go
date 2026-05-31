package handler

import "net/http"

type SystemHandler struct {
	links *LinkBuilder
}

func NewSystemHandler(builders ...*LinkBuilder) *SystemHandler {
	return &SystemHandler{
		links: linkBuilderOrDefault(builders),
	}
}

func (h *SystemHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	_ = writeData(w, r, http.StatusOK, nil, h.links.ServiceLinks(r))
}
