package handler

import (
	"fmt"
	"html"
	"net/http"
)

func Root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("ping %q", html.EscapeString(r.URL.Path))))
}
