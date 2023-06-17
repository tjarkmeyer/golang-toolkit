package session

import (
	"context"
	"net/http"
	"time"

	"github.com/Nerzal/gocloak/v11"
	"github.com/pkg/errors"
)

// CallOption configures a Session
type CallOption func(*goCloakSession) error

// PrematureRefreshThresholdOption sets the threshold for a premature token refresh
func PrematureRefreshThresholdOption(accessToken, refreshToken time.Duration) CallOption {
	return func(gcs *goCloakSession) error {
		gcs.prematureRefreshTokenRefreshThreshold = int(refreshToken.Seconds())
		gcs.prematureAccessTokenRefreshThreshold = int(accessToken.Seconds())
		return nil
	}
}

func SetGoCloak(gc gocloak.GoCloak) CallOption {
	return func(gcs *goCloakSession) error {
		gcs.gocloak = gc
		return nil
	}
}

type goCloakSession struct {
	clientID                              string
	clientSecret                          string
	username                              string
	password                              string
	realm                                 string
	gocloak                               gocloak.GoCloak
	token                                 *gocloak.JWT
	lastRequest                           *time.Time
	prematureRefreshTokenRefreshThreshold int
	prematureAccessTokenRefreshThreshold  int
}

// NewSession - new instance of a gocloak Session
func NewSession(clientID, clientSecret, username, password, realm, uri string, calloptions ...CallOption) (GoCloakSession, error) {
	session := &goCloakSession{
		clientID:                              clientID,
		clientSecret:                          clientSecret,
		username:                              username,
		password:                              password,
		realm:                                 realm,
		gocloak:                               gocloak.NewClient(uri),
		prematureAccessTokenRefreshThreshold:  0,
		prematureRefreshTokenRefreshThreshold: 0,
	}

	for _, option := range calloptions {
		err := option(session)
		if err != nil {
			return nil, errors.Wrap(err, "error while applying option")
		}
	}

	return session, nil
}

func (session *goCloakSession) ForceAuthenticate() error {
	return session.authenticate()
}

func (session *goCloakSession) ForceRefresh() error {
	return session.refreshToken()
}

func (session *goCloakSession) GetKeycloakAuthToken() (*gocloak.JWT, error) {
	if session.isAccessTokenValid() {
		return session.token, nil
	}

	if session.isRefreshTokenValid() {
		err := session.refreshToken()
		if err == nil {
			return session.token, nil
		}
	}

	err := session.authenticate()
	if err != nil {
		return nil, err
	}

	return session.token, nil
}

func (session *goCloakSession) isAccessTokenValid() bool {
	if session.token == nil {
		return false
	}

	if session.lastRequest.IsZero() {
		return false
	}

	sessionExpiry := session.token.ExpiresIn - session.prematureAccessTokenRefreshThreshold
	if int(time.Since(*session.lastRequest).Seconds()) > sessionExpiry {
		return false
	}

	token, _, err := session.gocloak.DecodeAccessToken(context.Background(), session.token.AccessToken, session.realm)
	return err == nil && token.Valid
}

func (session *goCloakSession) isRefreshTokenValid() bool {
	if session.token == nil {
		return false
	}

	if session.lastRequest.IsZero() {
		return false
	}

	sessionExpiry := session.token.RefreshExpiresIn - session.prematureRefreshTokenRefreshThreshold

	return int(time.Since(*session.lastRequest).Seconds()) <= sessionExpiry
}

func (session *goCloakSession) refreshToken() error {
	now := time.Now()
	session.lastRequest = &now

	jwt, err := session.gocloak.RefreshToken(context.Background(), session.token.RefreshToken, session.clientID, session.clientSecret, session.realm)
	if err != nil {
		return errors.Wrap(err, "could not refresh keycloak-token")
	}

	session.token = jwt

	return nil
}

func (session *goCloakSession) authenticate() error {
	now := time.Now()
	session.lastRequest = &now

	jwt, err := session.gocloak.LoginAdmin(context.Background(), session.username, session.password, session.realm)
	if err != nil {
		return errors.Wrap(err, "could not login to keycloak")
	}

	session.token = jwt

	return nil
}

func (session *goCloakSession) AddAuthTokenToRequest(request *http.Request) error {
	token, err := session.GetKeycloakAuthToken()
	if err != nil {
		return err
	}

	var tokenType string
	switch token.TokenType {
	case "bearer":
		tokenType = "Bearer"
	default:
		tokenType = token.TokenType
	}

	request.Header.Set("Authorization", tokenType+" "+token.AccessToken)

	return nil
}

func (session *goCloakSession) GetGoCloakInstance() gocloak.GoCloak {
	return session.gocloak
}
