package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/booscaaa/observability-go-example/pkg/adapter/instrumentation"
	"github.com/booscaaa/observability-go-example/pkg/adapter/metric"
	"github.com/booscaaa/observability-go-example/pkg/adapter/rest"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start an OpenTelemetry-instrumented HTTP service",
	Long: `Starts an HTTP service that is instrumented with OpenTelemetry for observability.

This service exposes metrics, traces, and logs that can be collected and analyzed
by OpenTelemetry collectors. The service responds with its name and demonstrates
distributed tracing capabilities when called through the Traefik reverse proxy.

Example usage:
  observability-go-example serve`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Initialize OpenTelemetry Instrumentation (TracerProvider)
		shutdownOtel, err := instrumentation.Initialize(cmd.Context())
		if err != nil {
			panic(fmt.Sprintf("failed to initialize OpenTelemetry instrumentation: %v", err))
		}

		// Defer OpenTelemetry shutdown
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) // Use background context for shutdown
			defer cancel()
			if err := shutdownOtel(ctx); err != nil {
				log.Printf("Error during OpenTelemetry shutdown: %v", err) // Use log instead of panic
			} else {
				log.Println("OpenTelemetry shutdown complete.")
			}
		}()

		// 2. Initialize OpenTelemetry Metrics (MeterProvider)
		metric.Initialize(otel.GetMeterProvider().Meter(instrumentation.Name))
		log.Println("OpenTelemetry metrics initialized.")

		// 3. Initialize and start the REST HTTP server
		server, shutdownServer, err := rest.Initialize() // Assuming rest.Initialize now returns server and shutdown func
		if err != nil {
			panic(fmt.Sprintf("failed to initialize REST server: %v", err))
		}
		log.Println("REST server initialized.")

		// Goroutine to start the HTTP server
		go func() {
			log.Printf("HTTP server listening on %s", server.Addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTP server ListenAndServe failed: %v", err)
			}
		}()

		// 4. Graceful shutdown handling
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit // Block until a signal is received
		log.Println("Received shutdown signal. Shutting down HTTP server...")

		ctxServer, cancelServer := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelServer()
		if err := shutdownServer(ctxServer); err != nil { // Call the server's shutdown function
			log.Fatalf("HTTP server forced to shutdown: %v", err)
		}
		log.Println("HTTP server shutdown complete.")

		// The deferred OpenTelemetry shutdown will be called next.
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
