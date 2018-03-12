package oauth2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	pg "github.com/topfreegames/extensions/pg/interfaces"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Token is a wrap over oauth2.Token
type Token struct {
	ID           string    `db:"id"`
	AccessToken  string    `db:"access_token"`
	RefreshToken string    `db:"refresh_token"`
	TokenType    string    `db:"token_type"`
	Email        string    `db:"email"`
	Expiry       time.Time `db:"expiry"`
}

// TokenStorage implementations are responsible to
// get, update and create tokens somewhere
type TokenStorage interface {
	Get(string) (*Token, error)
	Update(string, *oauth2.Token) error
	Create(*Token) error
	Delete(string) error
}

// Authenticator wrapper
type Authenticator struct {
	Config              oauth2.Config
	TS                  TokenStorage
	allowedEmailDomains []string
	authCodeOptions     []oauth2.AuthCodeOption
}

// Google authenticator ctor
func Google(ts TokenStorage) *Authenticator {
	return &Authenticator{
		Config: oauth2.Config{
			Endpoint: google.Endpoint,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
		},
		TS: ts,
	}
}

// RedirectURL is a builder method that sets RedirectURL
// and returns the same Authenticator ref
func (a *Authenticator) RedirectURL(url string) *Authenticator {
	a.Config.RedirectURL = url
	return a
}

// ClientSecret is a builder method that sets ClientSecret
// and returns the same Authenticator ref
func (a *Authenticator) ClientSecret(s string) *Authenticator {
	a.Config.ClientSecret = s
	return a
}

// ClientID is a builder method that sets ClientID
// and returns the same Authenticator ref
func (a *Authenticator) ClientID(id string) *Authenticator {
	a.Config.ClientID = id
	return a
}

// AllowedEmailDomains is a builder method that sets AllowedEmailDomains
// and returns the same Authenticator ref
func (a *Authenticator) AllowedEmailDomains(d ...string) *Authenticator {
	a.allowedEmailDomains = d
	return a
}

// AuthCodeOptions is a builder method that sets AuthCodeOptions and returns
// the same Authenticator ref
func (a *Authenticator) AuthCodeOptions(
	o ...oauth2.AuthCodeOption,
) *Authenticator {
	a.authCodeOptions = o
	return a
}

// Authenticate checks truthiness of token and if it's expired
// it'll be updated using RefreshToken information
// returns: (email, error)
func (a *Authenticator) Authenticate(token *Token) (string, error) {
	t, err := a.Config.TokenSource(oauth2.NoContext, &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}).Token()
	if err != nil || !t.Valid() {
		return "", errors.New("Authorization token is invalid")
	}
	if t.AccessToken != token.AccessToken {
		a.TS.Update(token.AccessToken, t)
	}
	return token.Email, nil
}

func (a *Authenticator) emailFromToken(t *oauth2.Token) (string, error) {
	client := a.Config.Client(oauth2.NoContext, t)
	res, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	status := res.StatusCode
	if status != http.StatusOK {
		return "", errors.New("Couldn't authorize token with provider")
	}
	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var m map[string]interface{}
	json.Unmarshal(bts, &m)
	return m["email"].(string), nil
}

func (a *Authenticator) isEmailDomainAllowed(email string) bool {
	d := strings.Split(email, "@")[1]
	for _, ad := range a.allowedEmailDomains {
		if d == ad {
			return true
		}
	}
	return false
}

// ExchangeCodeForToken exchange authorization code with access token
// Checks if email is in the allowed domains and calls TS.Create
func (a *Authenticator) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	token, err := a.Config.Exchange(oauth2.NoContext, code)
	if err != nil {
		err := errors.New("Couldn't exchange code")
		return nil, err
	}
	email, err := a.emailFromToken(token)
	if err != nil {
		return nil, err
	}
	if !a.isEmailDomainAllowed(email) {
		return nil, errors.New("Email isn't from an allowed domain")
	}
	err = a.TS.Create(&Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		TokenType:    token.TokenType,
		Email:        email,
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// AuthCodeURL generates URL to request sign-in w/ provider
func (a *Authenticator) AuthCodeURL(state string) string {
	return a.Config.AuthCodeURL(state, a.authCodeOptions...)
}

// PGTokenStorage implements TokenStorage over Postgres
type PGTokenStorage struct {
	TableName string
	DB        pg.DB
}

// Get searches for a Token in DB with access_token = ?
func (p *PGTokenStorage) Get(accessToken string) (*Token, error) {
	t := &Token{}
	_, err := p.DB.Query(
		t, fmt.Sprintf(`SELECT email, access_token, refresh_token, token_type, expiry
									FROM %s WHERE access_token = ?`, p.TableName),
		accessToken,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Update method
func (p *PGTokenStorage) Update(oldAccessToken string, t *oauth2.Token) error {
	_, err := p.DB.Exec(
		fmt.Sprintf(`UPDATE %s SET (access_token, expiry) = (?, ?)
								WHERE access_token = ?`, p.TableName),
		t.AccessToken, t.Expiry, oldAccessToken,
	)
	return err
}

// Create stores the token in postgres
func (p *PGTokenStorage) Create(t *Token) error {
	_, err := p.DB.Query(t, fmt.Sprintf(
		`INSERT INTO %s (access_token, refresh_token, token_type, expiry, email)
		VALUES (?access_token, ?refresh_token, ?token_type, ?expiry, ?email) RETURNING id`,
		p.TableName), t)
	return err
}

// Delete removes a token from storage
func (p *PGTokenStorage) Delete(accessToken string) error {
	_, err := p.DB.Exec(fmt.Sprintf(`DELETE FROM %s WHERE access_token = ?`,
		p.TableName), accessToken)
	return err
}
