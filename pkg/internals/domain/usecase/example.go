package usecase

import (
	"context"
	"os"

	"math/rand"

	"github.com/booscaaa/observability-go-example/internals/domain/core/domain"
	"github.com/booscaaa/observability-go-example/pkg/adapter/instrumentation"
)

type exampleUseCase struct {
	exampleHttpRepository domain.ExampleHttpRepository
}

// GetExample implements domain.ExampleUseCase.
func (usecase *exampleUseCase) GetExample(ctx context.Context) (*domain.Example, error) {
	ctx, span := instrumentation.Tracer.Start(ctx, "usecase.GetExample")
	defer span.End()

	instrumentation.Logger.InfoContext(ctx, "Processing example use case")

	url := os.Getenv("SERVICE_CALL_URL")

	length := rand.Intn(10) + 1
	letters := make([]rune, length)
	for i := range letters {
		letters[i] = rune('A' + rand.Intn(26))
	}
	word := string(letters)

	if url == "" {
		instrumentation.Logger.InfoContext(ctx, "No service URL configured, returning local letter",
			"word", word,
		)
		return &domain.Example{
			Word: word,
		}, nil
	}

	example, err := usecase.exampleHttpRepository.GetExample(ctx)
	if err != nil {
		instrumentation.Logger.ErrorContext(ctx, "Failed to get example from repository",
			"error", err,
			"url", url,
		)
		span.RecordError(err)
		return nil, err
	}

	example.ConcatenateWord(word)
	instrumentation.Logger.InfoContext(ctx, "Successfully processed example",
		"final_word", example.Word,
	)

	return example, nil
}

func NewExampleHttpRepository(exampleHttpRepository domain.ExampleHttpRepository) domain.ExampleUseCase {
	return &exampleUseCase{
		exampleHttpRepository: exampleHttpRepository,
	}
}
