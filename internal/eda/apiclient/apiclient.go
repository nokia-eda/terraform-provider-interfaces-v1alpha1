package apiclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nokia/eda/apps/terraform-provider-interfaces/internal/eda/rest"
)

const (
	// Constants
	KEY_CLIENT_ID      = "client_id"
	KEY_CLIENT_SECRET  = "client_secret"
	KEY_USERNAME       = "username"
	KEY_PASSWORD       = "password"
	KEY_GRANT_TYPE     = "grant_type"
	KEY_PASSWORD_GRANT = "password"
	KEY_REFRESH_GRANT  = "refresh_token"

	// URLs
	KEYCLOAK_URL = "/core/httpproxy/v1/keycloak"
	OAUTH_URL    = KEYCLOAK_URL + "/realms/%s/protocol/openid-connect/token"
	CLIENT_URL   = KEYCLOAK_URL + "/admin/realms/{realm}/clients"
)

type grant struct {
	AccessToken   string  `json:"access_token"`
	RefreshToken  string  `json:"refresh_token"`
	Scope         string  `json:"scope"`
	TokenType     string  `json:"token_type"`
	ExpiresInSecs float64 `json:"expires_in"`
	timestamp     *time.Time
}

type clientCredentials struct {
	authUrl      string
	clientId     string
	clientSecret string
	username     string
	password     string
}

type EdaApiClient struct {
	tokenLock     sync.Mutex
	cfg           *Config
	restClient    *rest.ApiClient
	edaCred       *clientCredentials
	keyCloakGrant *grant
	edaGrant      *grant
	logCtx        context.Context
}

type Config struct {
	BaseURL           string        `json:"baseURL"`
	KcUsername        string        `json:"kcUsername"`
	KcPassword        string        `json:"kcPassword"`
	KcRealm           string        `json:"kcRealm"`
	KcClientID        string        `json:"kcClientId"`
	EdaUsername       string        `json:"edaUsername"`
	EdaPassword       string        `json:"edaPassword"`
	EdaRealm          string        `json:"edaRealm"`
	EdaClientID       string        `json:"edaClientId"`
	EdaClientSecret   string        `json:"edaClientSecret"`
	TlsSkipVerify     bool          `json:"tlsSkipVerify"`
	RestDebug         bool          `json:"restDebug"`
	RestTimeout       time.Duration `json:"restTimeout"`
	RestRetries       int           `json:"restRetries"`
	RestRetryInterval time.Duration `json:"restRetryInterval"`
}

func (cfg *Config) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s: %s, ", "baseURL", cfg.BaseURL))
	sb.WriteString(fmt.Sprintf("%s: %s, ", "kcUsername", cfg.KcUsername))
	sb.WriteString(fmt.Sprintf("%s: %s, ", "kcRealm", cfg.KcRealm))
	sb.WriteString(fmt.Sprintf("%s: %s, ", "kcClientId", cfg.KcClientID))
	sb.WriteString(fmt.Sprintf("%s: %s, ", "edaUsername", cfg.EdaUsername))
	sb.WriteString(fmt.Sprintf("%s: %s, ", "edaRealm", cfg.EdaRealm))
	sb.WriteString(fmt.Sprintf("%s: %s, ", "edaClientId", cfg.EdaClientID))
	sb.WriteString(fmt.Sprintf("%s: %t, ", "tlsSkipVerify", cfg.TlsSkipVerify))
	sb.WriteString(fmt.Sprintf("%s: %t, ", "restDebug", cfg.RestDebug))
	sb.WriteString(fmt.Sprintf("%s: %s, ", "restTimeout", cfg.RestTimeout))
	sb.WriteString(fmt.Sprintf("%s: %d, ", "restRetries", cfg.RestRetries))
	sb.WriteString(fmt.Sprintf("%s: %s", "restRetryInterval", cfg.RestRetryInterval))
	return sb.String()
}

func NewEdaApiClient(logCtx context.Context, cfg *Config) (*EdaApiClient, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	client := &EdaApiClient{
		cfg: cfg,
		edaCred: &clientCredentials{
			authUrl:  fmt.Sprintf(OAUTH_URL, cfg.EdaRealm),
			clientId: cfg.EdaClientID,
			username: cfg.EdaUsername,
			password: cfg.EdaPassword,
		},
		keyCloakGrant: &grant{},
		edaGrant:      &grant{},
		logCtx:        logCtx,
	}
	client.restClient = rest.CreateApiClient().
		WithBaseURL(cfg.BaseURL).
		WithTimeout(cfg.RestTimeout).
		WithRetryCount(cfg.RestRetries).
		WithRetryInterval(cfg.RestRetryInterval).
		WithTlsConfig(&tls.Config{InsecureSkipVerify: cfg.TlsSkipVerify}).
		WithDebug(cfg.RestDebug)

	if cfg.EdaClientSecret != "" {
		client.edaCred.clientSecret = cfg.EdaClientSecret
		return client, nil
	}
	var err error
	client.edaCred.clientSecret, err = client.getClientSecret(cfg.EdaClientID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *EdaApiClient) getEdaAccessToken() (string, error) {
	return c.getAccessToken(c.edaCred, c.edaGrant)
}

// Attempt login with retries and exponential backoff
func (c *EdaApiClient) login(authUrl string, oauthBody map[string]string, grnt *grant) error {
	tflog.Trace(c.logCtx, "login()", map[string]any{"authUrl": authUrl, "oauthBody": fmt.Sprintf("%v", oauthBody)})
	var resp *resty.Response
	var err error
	maxRetries := 5
	baseDelay := time.Second

	for attempt := range maxRetries {
		resp, err = c.restClient.DoLogin(authUrl, oauthBody, grnt)
		if err == nil && !resp.IsError() {
			timestamp := time.Now()
			grnt.timestamp = &timestamp
			tflog.Info(c.logCtx, "login()", map[string]any{"authUrl": authUrl, "status": resp.Status(),
				"resp": resp.String(), "timeTaken": resp.Time().String()})
			return nil
		}

		// Log the error and response for debugging
		tflog.Error(c.logCtx, "Login attempt failed", map[string]any{
			"attempt": attempt + 1,
			"error":   err,
			"status":  resp.Status(),
			"body":    resp.String(),
		})

		// Exponential backoff before the next retry
		if attempt < maxRetries-1 { // Don’t sleep after last attempt
			time.Sleep(baseDelay * (1 << attempt))
		}
	}

	if err != nil {
		return fmt.Errorf("login failed after %d attempts: %w", maxRetries, err)
	}
	return fmt.Errorf("login failed after %d attempts: %s", maxRetries, resp.String())
}

func (c *EdaApiClient) getAccessToken(cred *clientCredentials, grnt *grant) (string, error) {
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	expired := false
	if grnt.timestamp != nil && grnt.ExpiresInSecs != 0 {
		elapsed := time.Since(*grnt.timestamp).Seconds()
		tflog.Debug(c.logCtx, "getAccessToken()", map[string]any{"authUrl": cred.authUrl, "timeElapsed": elapsed})
		expired = elapsed > grnt.ExpiresInSecs
	}
	tflog.Trace(c.logCtx, "getAccessToken()", map[string]any{"authUrl": cred.authUrl,
		"grantExpired": expired, "accessToken": grnt.AccessToken, "refreshToken": grnt.RefreshToken})

	if !expired && grnt.AccessToken != "" {
		tflog.Trace(c.logCtx, "getAccessToken()", map[string]any{"authUrl": cred.authUrl, "existingToken": grnt.AccessToken})
		return grnt.AccessToken, nil
	}
	var err error
	if expired && grnt.RefreshToken != "" {
		err = c.login(cred.authUrl, c.getOauthBody(cred, grnt.RefreshToken), grnt)
	} else {
		err = c.login(cred.authUrl, c.getOauthBody(cred, ""), grnt)
	}
	if err != nil {
		return "", err
	}
	if grnt.AccessToken == "" {
		return "", fmt.Errorf("access token is empty")
	}
	tflog.Trace(c.logCtx, "getAccessToken()", map[string]any{"authUrl": cred.authUrl, "newToken": grnt.AccessToken})
	return grnt.AccessToken, nil
}

func (c *EdaApiClient) getOauthBody(cred *clientCredentials, refreshToken string) map[string]string {
	oauthBody := make(map[string]string)
	oauthBody[KEY_CLIENT_ID] = cred.clientId
	oauthBody[KEY_CLIENT_SECRET] = cred.clientSecret
	if refreshToken != "" {
		oauthBody[KEY_GRANT_TYPE] = KEY_REFRESH_GRANT
		oauthBody[KEY_REFRESH_GRANT] = refreshToken
	} else {
		oauthBody[KEY_GRANT_TYPE] = KEY_PASSWORD_GRANT
		oauthBody[KEY_USERNAME] = cred.username
		oauthBody[KEY_PASSWORD] = cred.password
	}
	return oauthBody
}

func (c *EdaApiClient) getClientSecret(id string) (string, error) {
	keyCloakCred := &clientCredentials{
		authUrl:  fmt.Sprintf(OAUTH_URL, c.cfg.KcRealm),
		clientId: c.cfg.KcClientID,
		username: c.cfg.KcUsername,
		password: c.cfg.KcPassword,
	}
	accessToken, err := c.getAccessToken(keyCloakCred, c.keyCloakGrant)
	if err != nil {
		return "", err
	}

	result := []map[string]any{}
	resp, err := c.restClient.DoQuery(accessToken, CLIENT_URL, &result,
		map[string]string{"realm": c.cfg.EdaRealm},
		map[string]string{"clientId": id})
	if err != nil {
		return "", err
	}
	tflog.Info(c.logCtx, "getClientSecret()", map[string]any{"url": CLIENT_URL, "status": resp.Status(),
		"resp": resp.String(), "timeTaken": resp.Time().String()})

	if len(result) == 0 {
		return "", fmt.Errorf("client not found: %s", id)
	}
	secret, ok := result[0]["secret"]
	if !ok {
		return "", fmt.Errorf("client secret not found for client: %s", id)
	}
	tflog.Trace(c.logCtx, "getClientSecret()", map[string]any{"secret": secret})
	return secret.(string), nil
}

func (c *EdaApiClient) Create(ctx context.Context, pathUrl string, pathParams map[string]string, body any, result any) error {
	return c.Execute(ctx, pathUrl, rest.HTTP_POST, pathParams, nil, body, result)
}

func (c *EdaApiClient) Get(ctx context.Context, pathUrl string, pathParams map[string]string, result any) error {
	return c.Execute(ctx, pathUrl, rest.HTTP_GET, pathParams, nil, nil, result)
}

func (c *EdaApiClient) GetByQuery(ctx context.Context, pathUrl string, pathParams, queryParams map[string]string, result any) error {
	return c.Execute(ctx, pathUrl, rest.HTTP_GET, pathParams, queryParams, nil, result)
}

func (c *EdaApiClient) Update(ctx context.Context, pathUrl string, pathParams map[string]string, body any, result any) error {
	return c.Execute(ctx, pathUrl, rest.HTTP_PUT, pathParams, nil, body, result)
}

func (c *EdaApiClient) Delete(ctx context.Context, pathUrl string, pathParams map[string]string, result any) error {
	return c.Execute(ctx, pathUrl, rest.HTTP_DELETE, pathParams, nil, nil, result)
}

func (c *EdaApiClient) Execute(ctx context.Context, pathUrl, method string,
	pathParams, queryParams map[string]string, body, result any) error {
	accessToken, err := c.getEdaAccessToken()
	if err != nil {
		return err
	}
	tflog.Debug(c.logCtx, "Invoking DoExecute()::"+method+" "+pathUrl, map[string]any{
		"pathParams":  pathParams,
		"queryParams": queryParams,
	})
	resp, err := c.restClient.DoExecute(method, pathUrl, accessToken, body, result, pathParams, queryParams, nil)
	if err != nil {
		return err
	}
	tflog.Debug(c.logCtx, "After DoExecute()::"+method+" "+pathUrl, map[string]any{
		"status":    resp.Status(),
		"timeTaken": resp.Time().String(),
	})
	if resp.IsError() {
		return fmt.Errorf("%s %s", resp.Status(), resp.String())
	}
	return nil
}
