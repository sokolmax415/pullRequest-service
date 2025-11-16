package router

import (
	handler "pullrequest-service/internal/api/http/handlers"

	"github.com/go-chi/chi/v5"
)

func NewPRRouter(teamHandler *handler.PRHandler) chi.Router {
	r := chi.NewRouter()
	r.Post("/create", teamHandler.CreatePR)
	r.Post("/merge", teamHandler.MergePR)
	r.Post("/reassign", teamHandler.ReAssign)

	return r
}
