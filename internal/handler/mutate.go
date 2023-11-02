package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/dev-gaur/k8s-webhook/internal/mutate"
)

func Mutate(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		panic(fmt.Errorf("err reading the request body: %w", err))
	}

	mutated, err := mutate.Mutate(body)
	if err != nil {
		panic(fmt.Errorf("error in mutation: %w", err))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(mutated)

}
