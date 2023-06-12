package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v11"
	"github.com/tjarkmeyer/golang-toolkit/auth/v1/session"
	"github.com/tjarkmeyer/golang-toolkit/utils"
)

type KeycloakConfig struct {
	URL              string `default:"http://127.0.0.1:3005" envconfig:"KEYCLOAK_URL"`
	Base             string `default:"auth" envconfig:"KEYCLOAK_BASE"`
	Realm            string `default:"master" envconfig:"KEYCLOAK_REALM"`
	CliendID         string `default:"" envconfig:"KEYCLOAK_CLIENT_ID"`
	ClientSecret     string `default:"" envconfig:"KEYCLOAK_CLIENT_SECRET"`
	KeycloakUser     string `default:"admin" envconfig:"KEYCLOAK_USER"`
	KeycloakPassword string `default:"admin" envconfig:"KEYCLOAK_PASSWORD"`
}

type KcClient struct {
	c      gocloak.GoCloak
	realm  string
	id     string
	secret string
}

type KcSession struct {
	s        session.GoCloakSession
	realm    string
	clientID string
	secret   string
	username string
	password string
}

// CreateSession - creates gocloak session
func CreateSession(url, clientID, username, password, realm, clientSecret string) (*KcSession, error) {
	s, err := session.NewSession(clientID, clientSecret, username, password, realm, url)
	if err != nil {
		return nil, err
	}
	return &KcSession{
		s:        s,
		secret:   clientSecret,
		clientID: clientID,
		realm:    realm,
		username: username,
		password: password,
	}, nil
}

// CreateClient - creates gocloak client
func CreateClient(url, base, id, realm, secret string) *KcClient {
	return &KcClient{
		c: gocloak.NewClient(
			url,
			gocloak.SetAuthRealms(utils.MakeURL(base, "realms")),
			gocloak.SetAuthAdminRealms(utils.MakeURL(base, "admin", "realms")),
		),
		secret: secret,
		id:     id,
		realm:  realm,
	}
}

// GetGroupByID - get group by ID
func (client *KcSession) GetGroupByID(groupID string) (*gocloak.Group, error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err == nil {
		return client.s.GetGoCloakInstance().GetGroup(context.Background(), token.AccessToken, client.realm, groupID)
	}
	return nil, err
}

// CreateGroup - creates a keycloak group
func (client *KcSession) CreateGroup(group gocloak.Group) (string, error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err == nil {
		return client.s.GetGoCloakInstance().CreateGroup(context.Background(), token.AccessToken, client.realm, group)
	}
	return "", err
}

// UpdateGroup - update group
func (client *KcSession) UpdateGroup(group *gocloak.Group) error {
	token, err := client.s.GetKeycloakAuthToken()
	if err == nil {
		return client.s.GetGoCloakInstance().UpdateGroup(context.Background(), token.AccessToken, client.realm, *group)
	}
	return err
}

// GetFirstGroupByName - return first group with name (search query in DB: `name like %search%`)
func (client *KcSession) GetFirstGroupByName(name string) (*gocloak.Group, error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return nil, err
	}

	groupParams := gocloak.GetGroupsParams{
		Full:   gocloak.BoolP(true),
		Search: gocloak.StringP(name),
	}
	groups, err := client.s.GetGoCloakInstance().GetGroups(context.Background(), token.AccessToken, client.realm, groupParams)
	if err != nil {
		return nil, err
	}
	if len(groups) > 0 {
		return groups[0], nil
	}

	return &gocloak.Group{}, nil
}

// AddUserToGroup - add user to group
func (client *KcSession) AddUserToGroup(groupID string, userID string) error {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return err
	}
	return client.s.GetGoCloakInstance().AddUserToGroup(context.Background(), token.AccessToken, client.realm, userID, groupID)
}

// GetUser - get authorized user (From auth header)
func (client *KcSession) GetUser(ctx context.Context) (*gocloak.User, error) {
	emptyUser := gocloak.User{}

	userID, ok := GetUserID(ctx)
	if !ok {
		return &emptyUser, fmt.Errorf("can't get userID from auth header")
	}

	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return &emptyUser, err
	}

	return client.s.GetGoCloakInstance().GetUserByID(context.Background(), token.AccessToken, client.realm, userID)
}

// CreateUser - create new user
func (client *KcSession) CreateUser(user *gocloak.User) (string, error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return "", err
	}
	return client.s.GetGoCloakInstance().CreateUser(context.Background(), token.AccessToken, client.realm, *user)
}

// CreateUserWithMail - create new user and send mail
func (client *KcSession) CreateUserWithMail(user *gocloak.User) (string, error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return "", err
	}

	userID, err := client.s.GetGoCloakInstance().CreateUser(context.Background(), token.AccessToken, client.realm, *user)
	if err != nil {
		return "", err
	}

	return userID, client.ExecuteActionsEmail(userID, nil, 0)
}

func (client *KcSession) ExecuteActionsEmail(userID string, actions *[]string, lifespan int) error {
	if actions == nil || len(*actions) == 0 {
		actions = &[]string{"UPDATE_PASSWORD", "VERIFY_EMAIL"}
	}
	if lifespan == 0 {
		lifespan = 86400 * 30
	}

	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return err
	}

	params := gocloak.ExecuteActionsEmail{
		UserID:   &userID,
		Lifespan: gocloak.IntP(lifespan),
		Actions:  actions,
	}

	return client.s.GetGoCloakInstance().ExecuteActionsEmail(context.Background(), token.AccessToken, client.realm, params)
}

// AddRealmRolesToUser - add realm role to user
func (client *KcSession) AddRealmKeycloakRolesToUser(roles *[]gocloak.Role, userID string) error {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return err
	}

	return client.s.GetGoCloakInstance().AddRealmRoleToUser(context.Background(), token.AccessToken, client.realm, userID, *roles)
}

// AddRealmRolesToUser - adds a realm role to specified user
func (client *KcSession) AddRealmRolesToUser(roles []string, userID string) error {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return err
	}

	var kcRoles []gocloak.Role
	realmRoles, err := client.s.GetGoCloakInstance().GetRealmRoles(context.Background(), token.AccessToken, client.realm, gocloak.GetRoleParams{})
	if err != nil {
		return err
	}
	for _, realmRole := range realmRoles {
		if utils.Contains(roles, *realmRole.Name) {
			kcRoles = append(kcRoles, *realmRole)
		}
	}

	return client.s.GetGoCloakInstance().AddRealmRoleToUser(context.Background(), token.AccessToken, client.realm, userID, kcRoles)
}

// DeleteRealmRoleFromUser - deletes a realm role from user
func (client *KcSession) DeleteRealmRoleFromUser(role string, userID string) error {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return err
	}

	var roles []gocloak.Role
	realmRoles, err := client.s.GetGoCloakInstance().GetRealmRoles(context.Background(), token.AccessToken, client.realm, gocloak.GetRoleParams{})
	if err != nil {
		return err
	}
	for _, realmRole := range realmRoles {
		if role == *realmRole.Name {
			roles = append(roles, *realmRole)
		}
	}

	return client.s.GetGoCloakInstance().DeleteRealmRoleFromUser(context.Background(), token.AccessToken, client.realm, userID, roles)
}

func (client *KcSession) GetGroupMembers(groupName string, params ...gocloak.GetGroupsParams) ([]*gocloak.User, error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return []*gocloak.User{}, err
	}

	group, err := client.GetFirstGroupByName(groupName)
	if err != nil {
		return []*gocloak.User{}, err
	}

	var foundGroupId string
	for _, groupIn := range *group.SubGroups {
		if *groupIn.Name == groupName {
			foundGroupId = *groupIn.ID
		}
	}

	if foundGroupId == "" {
		return []*gocloak.User{}, errors.New("group does not exist or failed to find group id")
	}

	groupParams := gocloak.GetGroupsParams{BriefRepresentation: gocloak.BoolP(false), Max: gocloak.IntP(1000)}
	if len(params) > 0 {
		groupParams = params[0]
	}
	return client.s.GetGoCloakInstance().GetGroupMembers(context.Background(), token.AccessToken, client.realm, foundGroupId, groupParams)
}

func (client *KcSession) GetUserById(ctx context.Context, userID string) (*gocloak.User, error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return nil, err
	}

	user, err := client.s.GetGoCloakInstance().GetUserByID(ctx, token.AccessToken, client.realm, userID)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (client *KcSession) ReinviteUserById(ctx context.Context, userID string) error {
	user, err := client.GetUserById(ctx, userID)
	if err != nil {
		return err
	}
	if user != nil && *user.EmailVerified {
		return errors.New("user email is already verified, skip the action")
	}

	actions := &[]string{"UPDATE_PASSWORD", "VERIFY_EMAIL"}
	err = client.ExecuteActionsEmail(userID, actions, 0)
	if err != nil {
		return err
	}

	timeNow := time.Now().UTC().Unix() * 1000
	attributes := *user.Attributes
	attributes["invited"] = []string{strconv.FormatInt(timeNow, 10)}
	user.Attributes = &attributes
	*user.RequiredActions = append(*user.RequiredActions, *actions...)
	return client.UpdateUserProperties(user)
}

// DeleteUser - delete a given user
func (client *KcSession) DeleteUser(ctx context.Context, userID string) error {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return err
	}
	return client.s.GetGoCloakInstance().DeleteUser(ctx, token.AccessToken, client.realm, userID)
}

// UpdateUserProperties updates a user
func (client *KcSession) UpdateUserProperties(user *gocloak.User) error {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return err
	}
	return client.s.GetGoCloakInstance().UpdateUser(context.Background(), token.AccessToken, client.realm, *user)
}

// GetUsers - gets all user
func (client *KcSession) GetUsers(ctx context.Context, params ...gocloak.GetUsersParams) (users []*gocloak.User, err error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return
	}

	userParams := gocloak.GetUsersParams{BriefRepresentation: gocloak.BoolP(false)}
	if len(params) > 0 {
		userParams = params[0]
	}

	return client.s.GetGoCloakInstance().GetUsers(context.Background(), token.AccessToken, client.realm, userParams)
}

// GetUserGroups - gets all group memberships of a user
func (client *KcSession) GetUserGroups(ctx context.Context, userId string) (groups []*gocloak.Group, err error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return
	}
	return client.s.GetGoCloakInstance().GetUserGroups(context.Background(), token.AccessToken, client.realm, userId, gocloak.GetGroupsParams{})
}

// DeleteUserFromGroup - deletes a user from a group
func (client *KcSession) DeleteUserFromGroup(ctx context.Context, userId, groupId string) (err error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return
	}
	return client.s.GetGoCloakInstance().DeleteUserFromGroup(ctx, token.AccessToken, client.realm, userId, groupId)
}

// HasUserRealmRoleById returns if a user has a specific role
func (client *KcSession) HasUserRealmRoleById(ctx context.Context, userId, roleName string) (bool, error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return false, err
	}

	roles, err := client.s.GetGoCloakInstance().GetRealmRolesByUserID(ctx, token.AccessToken, client.realm, userId)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if *role.Name == roleName {
			return true, nil
		}
	}
	return false, nil
}

// GetEvents - returns all events (optional: GetEventsParams)
func (client *KcSession) GetEvents(config KeycloakConfig, params ...gocloak.GetEventsParams) ([]*gocloak.EventRepresentation, error) {
	var eventRep []*gocloak.EventRepresentation
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return eventRep, err
	}

	eventParams := gocloak.GetEventsParams{Max: gocloak.Int32P(250)}
	if len(params) > 0 {
		eventParams = params[0]
	}

	queryParams := "?"

	if eventParams.First != nil {
		queryParams += fmt.Sprintf("first=%d&", *eventParams.First)
	}
	if eventParams.Max != nil {
		queryParams += fmt.Sprintf("max=%d&", *eventParams.Max)
	}
	if eventParams.Client != nil {
		queryParams += fmt.Sprintf("client=%s&", *eventParams.Client)
	}
	if eventParams.DateFrom != nil {
		queryParams += fmt.Sprintf("dateFrom=%s&", *eventParams.DateFrom)
	}
	if eventParams.DateTo != nil {
		queryParams += fmt.Sprintf("dateTo=%s&", *eventParams.DateTo)
	}
	if eventParams.UserID != nil {
		queryParams += fmt.Sprintf("user=%s&", *eventParams.UserID)
	}

	if len(eventParams.Type) > 0 {
		for _, typeIn := range eventParams.Type {
			queryParams += fmt.Sprintf("&type=%s", typeIn)
		}
	}

	body := strings.NewReader("")
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/admin/realms/%s/events%s", config.URL, config.Base, config.Realm, queryParams), body)
	if err != nil {
		return eventRep, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return eventRep, err
	}
	defer resp.Body.Close()
	bodyOut, err := io.ReadAll(resp.Body)
	if err != nil {
		return eventRep, err
	}

	err = json.Unmarshal(bodyOut, &eventRep)

	return eventRep, err
}

// GetUserFederatedIdentities - returns all federated identities (IDPs) of a user
func (client *KcSession) GetUserFederatedIdentities(ctx context.Context, userID string) ([]*gocloak.FederatedIdentityRepresentation, error) {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return []*gocloak.FederatedIdentityRepresentation{}, err
	}

	return client.s.GetGoCloakInstance().GetUserFederatedIdentities(ctx, token.AccessToken, client.realm, userID)
}

// SetPassword - sets a new password for the user with the given id. Needs elevated privileges.
func (client *KcSession) SetPassword(ctx context.Context, userID, password string, temporary bool) error {
	token, err := client.s.GetKeycloakAuthToken()
	if err != nil {
		return err
	}
	return client.s.GetGoCloakInstance().SetPassword(ctx, token.AccessToken, userID, client.realm, password, temporary)
}
