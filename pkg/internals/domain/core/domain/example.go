package domain

import (
	"context"
	"net/http"
)

type Example struct {
	Word string `json:"word"`
}

type ExampleHttpRepository interface {
	GetExample(context.Context) (*Example, error)
}

type ExampleUseCase interface {
	GetExample(context.Context) (*Example, error)
}

type ExampleController interface {
	GetExample(http.ResponseWriter, *http.Request)
}

func (example *Example) ConcatenateWord(word string) {
	example.Word += word
}
