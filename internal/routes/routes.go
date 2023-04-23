package routes

import (
	"github.com/cf-sewe/cplace-cssc-operator/internal/environment"
  "github.com/go-chi/chi"
)

// Init initializes the routes for the given environment.
func Init(env environment.Environment) {
  // init chi router and add an example route
  r := chi.NewRouter()
  r.Get("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello world!"))
  }
}
