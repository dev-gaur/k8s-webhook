package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dev-gaur/k8s-webhook/internal/handler"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	debug := flag.Bool("debug", false, "enable for verbose logs")

	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Msg("starting server...")

	r := mux.NewRouter()
	r.Use(logger)
	r.Use(recovery)
	r.HandleFunc("/", handler.Root)
	r.HandleFunc("/mutate", handler.Mutate)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

	err := s.ListenAndServe()
	//err := s.ListenAndServeTLS("/etc/secrets/tls.crt", "/etc/secrets/tls.key")
	if err != nil {
		panic(fmt.Sprintf("server crashed: %s", err.Error()))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Shutdown(ctx)

	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("server failed to exit gracefully. error: '%v'", err))
	}
	log.Info().Msg("server exiting gracefully")
}
