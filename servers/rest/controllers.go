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

// NewController - returns new rest definition controller
func NewController() *Definitions {
	return &Definitions{}
}

// AddController - adds a new controller
func (r *Definitions) AddController(controller *Definition) {
	r.Controllers = append(r.Controllers, controller)
}

// CreateController - createa a new rest controller
func (r *Definitions) CreateController() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/v1", func(route chi.Router) {
		for _, controllerIn := range r.Controllers {
			route.Mount("/", controllerIn.Controller)
		}
	})

	return router
}

// CreateControllerByName - createa a new rest controller
func (r *Definitions) CreateControllerByName() *chi.Mux {
	router := chi.NewRouter()

	router.Route("/v1", func(route chi.Router) {
		for _, controllerIn := range r.Controllers {
			route.Mount(fmt.Sprintf("/%s", controllerIn.Name), controllerIn.Controller)
		}
	})

	return router
}
