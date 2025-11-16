package router

import (
	handler "pullrequest-service/internal/api/http/handlers"

	"github.com/go-chi/chi/v5"
)

func NewUserRouter(userHandler *handler.UserHandler) chi.Router {
	r := chi.NewRouter()
	r.Post("/setIsActive", userHandler.SetActive)
	r.Get("/getReview", userHandler.GetPR)

	return r
}
