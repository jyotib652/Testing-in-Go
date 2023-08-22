package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	// register recoverer middleware
	mux.Use(middleware.Recoverer)
	mux.Use(app.addIPToContext)
	mux.Use(app.Session.LoadAndSave)

	// register routes
	mux.Get("/", app.Home)
	mux.Post("/login", app.Login)

	// mux.Get("/user/profile", app.Profile)

	// Now we want to secure our "/user/profile" route with our custom auth middleware
	// So we would create a special route that would use the auth() method of middleware that we wrote
	// and we should use plain "mux.Get("/user/profile", app.Profile)" route anymore and we should use
	// subrouter along with a middleware
	mux.Route("/user", func(mux chi.Router) {
		mux.Use(app.auth)
		mux.Get("/profile", app.Profile)
	})

	// static assets (css, javascript, images)
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
