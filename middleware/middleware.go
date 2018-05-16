package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/topfreegames/extensions/oauth2"
	pg "github.com/topfreegames/extensions/pg/interfaces"
)

type ctxKey string

var ctxKeys = struct {
	logger          ctxKey
	metricsReporter ctxKey
	db              ctxKey
	body            ctxKey
	oauth2          ctxKey
}{
	logger:          "logger",
	metricsReporter: "metricsReporter",
	db:              "db",
	body:            "body",
	oauth2:          "oauth2",
}

// GetLogger returns a logrus.FieldLogger from context
func GetLogger(ctx context.Context) logrus.FieldLogger {
	return ctx.Value(ctxKeys.logger).(logrus.FieldLogger)
}

// GetMetricsReporter returns a MetricsReporter from context
func GetMetricsReporter(ctx context.Context) MetricsReporter {
	return ctx.Value(ctxKeys.metricsReporter).(MetricsReporter)
}

// GetDB returns a pginterfaces.DB from context
func GetDB(ctx context.Context) pg.DB {
	return ctx.Value(ctxKeys.db).(pg.DB)
}

// GetBody returns a pginterfaces.DB from context
func GetBody(ctx context.Context) interface{} {
	return ctx.Value(ctxKeys.body)
}

// Version middleware
func Version(v string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Version", v)
			next.ServeHTTP(w, r)
		})
	}
}

// Logging middleware
func Logging(logger logrus.FieldLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := uuid.NewV4().String()
			l := logger.WithField("requestID", reqID)
			ctx := context.WithValue(r.Context(), ctxKeys.logger, l)
			start := time.Now()

			defer func() {
				status := GetStatusCode(w)
				route, _ := mux.CurrentRoute(r).GetPathTemplate()
				lf := GetLogger(ctx).WithFields(logrus.Fields{
					"operation":       "serveHTTP",
					"path":            r.URL.Path,
					"method":          r.Method,
					"route":           route,
					"requestDuration": int(time.Since(start).Nanoseconds() / 1e6),
					"status":          status,
				})

				if status > 399 && status < 500 {
					lf.Warn("Request failed.")
				} else if status > 499 {
					lf.Error("Response failed.")
				} else {
					lf.Debug("Request successful.")
				}
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Metrics middleware
func Metrics(mr MetricsReporter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxKeys.metricsReporter, mr)
			start := time.Now()

			defer func() {
				statusCode := GetStatusCode(w)
				errored := statusCode > 299
				elapsed := time.Since(start)
				route, _ := mux.CurrentRoute(r).GetPathTemplate()
				tags := []string{
					fmt.Sprintf("status:%d", statusCode),
					fmt.Sprintf("route:%s %s", r.Method, route),
					fmt.Sprintf("type:http"),
					fmt.Sprintf("error:%t", errored),
				}
				mr.Timing(MetricTypes.ResponseTimeMs, elapsed, tags...)
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// DB middleware
func DB(db pg.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(
				w, r.WithContext(context.WithValue(r.Context(), ctxKeys.db, db)),
			)
		})
	}
}

// BodyParser middleware
func BodyParser(holder interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := GetLogger(r.Context())
			if r.Body == nil {
				writeStatus(w, http.StatusBadRequest)
				return
			}
			defer r.Body.Close()
			bts, err := ioutil.ReadAll(r.Body)
			if err != nil {
				l.Error(err)
				writeStatus(w, http.StatusInternalServerError)
				return
			}
			h := reflect.New(
				reflect.Indirect(reflect.ValueOf(holder)).Type(),
			).Interface()
			err = json.Unmarshal(bts, h)
			if err != nil {
				l.Error(err)
				writeStatus(w, http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), ctxKeys.body, h)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Validator middleware
func Validator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := GetLogger(r.Context())
			body := GetBody(r.Context())
			_, err := govalidator.ValidateStruct(body)
			if err != nil {
				l.Error(err)
				write(w, http.StatusUnprocessableEntity, err.Error())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func oauth2login(a *oauth2.Authenticator, r *http.Request) (string, error) {
	state := r.FormValue("state")
	code := r.FormValue("code")

	if state == "" && code == "" {
		return "", errors.New("No state or code were sent")
	}

	if state != "" {
		return a.AuthCodeURL(state), nil
	}

	t, err := a.ExchangeCodeForToken(code)
	if err != nil {
		return "", err
	}
	return t.AccessToken, nil
}

func oauth2logout(a *oauth2.Authenticator, r *http.Request) error {
	accessToken :=
		strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return a.TS.Delete(accessToken)
}

// OAuth2Paths describe paths used by OAuth2 middleware
type OAuth2Paths struct {
	// Whitelist is in mux.CurrentRoute(r).GetPathTemplate() format
	Whitelist []string
	Login     string
	Logout    string
}

// OAuth2 middleware
func OAuth2(
	enabled bool, authenticator *oauth2.Authenticator, paths OAuth2Paths,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			l := GetLogger(r.Context())

			switch r.URL.Path {
			case paths.Login:
				r, err := oauth2login(authenticator, r)
				if err != nil {
					l.Error(err)
					writeStatus(w, http.StatusInternalServerError)
					return
				}
				write(w, http.StatusOK, r)
				return
			case paths.Logout:
				err := oauth2logout(authenticator, r)
				if err != nil {
					l.Error(err)
					writeStatus(w, http.StatusInternalServerError)
					return
				}
				writeStatus(w, http.StatusAccepted)
				return
			}

			pathTemplate, err := mux.CurrentRoute(r).GetPathTemplate()
			if err != nil {
				l.Error(err)
				writeStatus(w, http.StatusInternalServerError)
				return
			}

			for _, p := range paths.Whitelist {
				if p == pathTemplate {
					next.ServeHTTP(w, r)
					return
				}
			}

			accessToken :=
				strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

			token, err := authenticator.TS.Get(accessToken)
			if err != nil {
				l.Error(err)
				write(w, http.StatusForbidden, "Authorization token doesn't exist")
				return
			}

			email, err := authenticator.Authenticate(token)
			if err != nil {
				l.Error(err)
				write(w, http.StatusForbidden, "Authorization token is invalid")
				return
			}

			w.Header().Set("X-Access-Token", token.AccessToken)
			ctx := context.WithValue(r.Context(), ctxKeys.oauth2, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
