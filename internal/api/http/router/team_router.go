package router

import (
	handler "pullrequest-service/internal/api/http/handlers"

	"github.com/go-chi/chi/v5"
)

func NewTeamRouter(teamHandler *handler.TeamHandler) chi.Router {
	r := chi.NewRouter()
	r.Post("/add", teamHandler.AddTeam)
	r.Get("/get", teamHandler.GetTeam)

	return r
}
