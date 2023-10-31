package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	m "github.com/dev-gaur/k8s-webhook/internal/mutate"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
	// read the body / request
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		panic(fmt.Errorf("err reading the request body: %v", err.Error()))
	}

	// mutate the request
	mutated, err := m.Mutate(body, true)
	if err != nil {
		panic(fmt.Errorf("error in mutation: %v", err))
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(mutated)
}

func main() {
	log.Info().Msg("starting server...")

	r := mux.NewRouter()
	r.Use(logMW)
	r.Use(recovery)
	r.HandleFunc("/", handleRoot)
	r.HandleFunc("/mutate", handleMutate)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

	err := s.ListenAndServeTLS("/etc/secrets/tls.crt", "/etc/secrets/tls.key")
	if err != nil {
		panic(fmt.Sprintf("server crashed: %s", err.Error()))
	}
}

// for global use (using a http.Handler!)
func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg(fmt.Sprintf("%s - %s (%s)", r.Method, r.URL.Path, r.RemoteAddr))

		// compare the return-value to the authMW
		next.ServeHTTP(w, r)
	})
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			err := recover()
			if err != nil {
				log.Error().Msg(fmt.Sprintf("%v", err))

				jsonBody, _ := json.Marshal(map[string]interface{}{
					"error":     "internal server error",
					"errorBody": err,
				})

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(jsonBody)
			}

		}()

		next.ServeHTTP(w, r)

	})
}
