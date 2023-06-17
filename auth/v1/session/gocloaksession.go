package session

import (
	"net/http"

	"github.com/Nerzal/gocloak/v11"
)

// GoCloakSession - the gocloak session
type GoCloakSession interface {
	// GetKeycloakAuthToken returns a JWT object, containing the AccessToken and more
	GetKeycloakAuthToken() (*gocloak.JWT, error)

	// AddAuthTokenToRequest sets the Authentication Header for the response
	AddAuthTokenToRequest(*http.Request) error

	// GetGoCloakInstance returns the currently used GoCloak instance.
	GetGoCloakInstance() gocloak.GoCloak

	// ForceAuthenticate ignores all checks and executes an authentication.
	ForceAuthenticate() error

	// ForceRefresh ignores all checks and executes a refresh.
	ForceRefresh() error
}
