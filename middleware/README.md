## Usage example

### oauth2 authenticator
```go
	import (
		"github.com/topfreegames/extensions/oauth2"
		o2 "golang.org/x/oauth2"
	)

	authenticator := oauth2.Google(
		&oauth2.PGTokenStorage{
			TableName: "oauth2_tokens",
			DB:        a.DB,
		},
	).
		ClientID(a.Config.GetString("oauth2.clientID")).
		ClientSecret(a.Config.GetString("oauth2.clientSecret")).
		RedirectURL(a.Config.GetString("oauth2.redirectURL")).
		AllowedEmailDomains(
			a.Config.GetStringSlice("oauth2.allowedEmailDomains")...,
		).
		AuthCodeOptions(o2.AccessTypeOffline, o2.ApprovalForce)
}
```

### *mux.Router
```go
	r := mux.NewRouter()
	r.Use(middleware.Version(models.AppInfo.Version))
	r.Use(middleware.Logging(a.Logger))
	r.Use(middleware.Metrics(a.MetricsReporter))
	r.Use(middleware.DB(a.DB))
	r.Use(
		middleware.OAuth2(
			config.GetBool("oauth2.enabled"),
			a.configureAuthenticator(),
			middleware.OAuth2Paths{
				Whitelist: []string{"/healthcheck"},
				Login:     "/login",
				Logout:    "/logout",
			},
		),
	)

	r.Handle("/healthcheck", Chain(
		NewHealthcheckHandler(a),
	)).Methods("GET").Name("healthcheck")

	r.HandleFunc("/login", BlankHandler).Methods("GET").Name("login")
	r.HandleFunc("/logout", BlankHandler).Methods("GET").Name("logout")
```
