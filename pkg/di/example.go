package di

import (
	"github.com/booscaaa/observability-go-example/internals/domain/core/domain"
	"github.com/booscaaa/observability-go-example/internals/domain/usecase"
	"github.com/booscaaa/observability-go-example/internals/infra/controller"
	"github.com/booscaaa/observability-go-example/internals/infra/repository"
)

func NewExampleController() domain.ExampleController {
	// Este é o "adaptador de saída" que faz chamadas para os outros serviços
	exampleHttpRepository := repository.NewExampleHttpRepository()
	// Cria o caso de uso (use case), injetando o repositório nele
	exampleUseCase := usecase.NewExampleHttpRepository(exampleHttpRepository)
	// Cria o controller, injetando o caso de uso nele.
	// Este é o "adaptador de entrada" que lida com a requisição HTTP.
	return controller.NewExampleController(exampleUseCase)
}
