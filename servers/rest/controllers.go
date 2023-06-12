package rest

import (
	"fmt"

	"github.com/go-chi/chi/v5"
)

// Definitions - definition for (multiple) rest controller
type Definitions struct {
	Controllers []*Definition
}

// Definition - holds a controller and the controllers name
type Definition struct {
	Controller *chi.Mux
	Name       string
}

// NewRestController - returns new rest definition
func NewRestController() *Definitions {
	return &Definitions{}
}

// AddController - adds a new controller
func (r *Definitions) AddController(controller *Definition) {
	r.Controllers = append(r.Controllers, controller)
}

// CreateRestController - createa a new rest controller
func (r *Definitions) CreateRestController() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/v1", func(route chi.Router) {
		for _, controllerIn := range r.Controllers {
			route.Mount("/", controllerIn.Controller)
		}
	})

	return router
}

// CreateRestControllerByName - createa a new rest controller
func (r *Definitions) CreateRestControllerByName() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/v1", func(route chi.Router) {
		for _, controllerIn := range r.Controllers {
			route.Mount(fmt.Sprintf("/%s", controllerIn.Name), controllerIn.Controller)
		}
	})

	return router
}
