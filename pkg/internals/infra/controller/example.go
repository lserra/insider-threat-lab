package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/booscaaa/observability-go-example/internals/domain/core/domain"
	"github.com/booscaaa/observability-go-example/pkg/adapter/instrumentation"
	"github.com/booscaaa/observability-go-example/pkg/adapter/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type exampleController struct {
	exampleUseCase domain.ExampleUseCase
}

// GetExample implements domain.ExampleController.
func (controller *exampleController) GetExample(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, propagation.HeaderCarrier(request.Header))

	ctx, span := instrumentation.Tracer.Start(ctx, "controller.GetExample")
	defer span.End()

	start := time.Now()
	metric.RequestInFlight.Add(ctx, 1) // contar requisições em voo.
	metric.RequestCounter.Add(ctx, 1)  // contar o total de requisições.
	defer func() {
		// contar requisições em voo.
		metric.RequestInFlight.Add(ctx, -1)
		// registar a duração da requisição.
		metric.RequestDuration.Record(ctx, time.Since(start).Seconds())
	}()

	instrumentation.Logger.InfoContext(ctx, "Processing request",
		"method", request.Method,
		"path", request.URL.Path,
	)

	example, err := controller.exampleUseCase.GetExample(ctx)
	if err != nil {
		instrumentation.Logger.ErrorContext(ctx, "Failed to get example",
			"error", err,
			"method", request.Method,
			"path", request.URL.Path,
			"status", http.StatusInternalServerError,
		)
		span.RecordError(err)
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(err.Error()))
		return
	}

	// usado após a chamada bem-sucedida ao caso de uso.
	metric.LetterCounter.Add(ctx, 1)
	// registar o comprimento da palavra.
	metric.WordLengthGauge.Record(ctx, int64(len(example.Word)))

	response.WriteHeader(http.StatusOK)
	err = json.NewEncoder(response).Encode(example)
	if err != nil {
		instrumentation.Logger.ErrorContext(ctx, "Failed to encode response",
			"error", err,
			"method", request.Method,
			"path", request.URL.Path,
			"status", http.StatusInternalServerError,
		)
		span.RecordError(err)
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(err.Error()))
		return
	}
}

func NewExampleController(exampleUseCase domain.ExampleUseCase) domain.ExampleController {
	return &exampleController{
		exampleUseCase: exampleUseCase,
	}
}
