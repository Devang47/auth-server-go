package rest

import (
	"auth-server-go/internal/middlewares"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gorm.io/gorm"
)

type REST struct {
	DB     *gorm.DB
	Router *chi.Mux
}

func SetupREST(db *gorm.DB) REST {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(render.SetContentType(render.ContentTypeJSON))

	rest := REST{
		DB:     db,
		Router: router,
	}

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/"))
	})

	// Health check
	router.Get("/health", rest.GetHealth)

	public := rest.Router // public routes

	secureRoutes := rest.Router.Route("/secure", func(r chi.Router) {})
	secureRoutes.Use(middlewares.ProtectHandler)

	AddGoogleAuthRoutes(rest, public)
	AddAccountRoutes(rest, public)

	return rest
}
