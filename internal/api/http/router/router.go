package router

import (
	handler "pullrequest-service/internal/api/http/handlers"

	"github.com/go-chi/chi/v5"
)

func NewRouter(teamHandler *handler.TeamHandler, userHandler *handler.UserHandler, prHandler *handler.PRHandler) chi.Router {
	r := chi.NewRouter()

	r.Mount("/team", NewTeamRouter(teamHandler))
	r.Mount("/users", NewUserRouter(userHandler))
	r.Mount("/pullRequest", NewPRRouter(prHandler))

	return r
}
