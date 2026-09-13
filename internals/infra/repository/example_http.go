package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/booscaaa/observability-go-example/internals/domain/core/domain"
	"github.com/booscaaa/observability-go-example/pkg/adapter/instrumentation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type exampleHttpRepository struct{}

// GetExample implements domain.ExampleHttpRepository.
func (repository *exampleHttpRepository) GetExample(ctx context.Context) (*domain.Example, error) {
	ctx, span := instrumentation.Tracer.Start(ctx, "repository.GetExample")
	defer span.End()

	time.Sleep(10 * time.Second)

	url := os.Getenv("SERVICE_CALL_URL")
	instrumentation.Logger.InfoContext(ctx, "Making HTTP request", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		instrumentation.Logger.ErrorContext(ctx, "Failed to create request",
			"error", err,
			"url", url,
		)
		span.RecordError(err)
		return nil, err
	}

	propagator := otel.GetTextMapPropagator()
	// Injete o contexto de trace nos cabeçalhos da requisição
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		instrumentation.Logger.ErrorContext(ctx, "Failed to execute request",
			"error", err,
			"url", url,
		)
		span.RecordError(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		instrumentation.Logger.ErrorContext(ctx, "Received non-200 status code",
			"error", err,
			"status_code", resp.StatusCode,
			"url", url,
		)
		span.RecordError(err)
		return nil, err
	}

	var exampleResponse domain.Example
	if err := json.NewDecoder(resp.Body).Decode(&exampleResponse); err != nil {
		instrumentation.Logger.ErrorContext(ctx, "Failed to decode response",
			"error", err,
			"url", url,
		)
		span.RecordError(err)
		return nil, err
	}

	instrumentation.Logger.InfoContext(ctx, "Successfully retrieved example",
		"word", exampleResponse.Word,
	)

	return &exampleResponse, nil
}

func NewExampleHttpRepository() domain.ExampleHttpRepository {
	return &exampleHttpRepository{}
}
