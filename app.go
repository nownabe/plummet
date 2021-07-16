package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type app struct {
}

func newApp() (*app, error) {
	return &app{}, nil
}

func (a *app) handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)

	r.Get("/", a.handle)

	return r
}

func (a *app) handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}
