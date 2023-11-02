package handler

import (
	"fmt"
	"html"
	"net/http"

	"github.com/rs/zerolog/log"
)

func Root(w http.ResponseWriter, r *http.Request) {

	log.Debug().Msg("inside root handler")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("ping %q", html.EscapeString(r.URL.Path))))
}
