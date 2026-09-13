package rest

import (
	"context"
	"net/http"

	"github.com/booscaaa/observability-go-example/pkg/di"
)

// Initialize configures and returns an HTTP server and its shutdown function.
func Initialize() (*http.Server, func(context.Context) error, error) {
	exampleController := di.NewExampleController()
	mux := http.NewServeMux()

	mux.HandleFunc("/", exampleController.GetExample)

	// Create the HTTP server instance
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Return the server, its shutdown function, and no error
	return server, server.Shutdown, nil
}
