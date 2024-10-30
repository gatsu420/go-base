// Package admin ties together administration resources and handlers.
package admin

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"

	"github.com/go-chi/chi/v5"

	"github.com/dhax/go-base/auth/authorize"
	"github.com/dhax/go-base/database"
	"github.com/dhax/go-base/logging"
)

const (
	roleAdmin = "admin"
)

type ctxKey int

const (
	ctxAccount ctxKey = iota
)

// API provides admin application resources and handlers.
type API struct {
	Accounts *AccountResource
	Names    *NameResource
	NameAry  *NameAryResource
}

// NewAPI configures and returns admin application API.
func NewAPI(db *bun.DB) (*API, error) {
	accountStore := database.NewAdmAccountStore(db)
	accounts := NewAccountResource(accountStore)
	names := NewNameResource(accountStore)
	names_ary := NewNameAryResource(accountStore)

	api := &API{
		Accounts: accounts,
		Names:    names,
		NameAry:  names_ary,
	}
	return api, nil
}

// Router provides admin application routes.
func (a *API) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(authorize.RequiresRole(roleAdmin))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Admin"))
	})

	r.Mount("/accounts", a.Accounts.router())

	r.Mount("/names", a.Names.router())

	r.Mount("/names_ary", a.NameAry.router())

	return r
}

func log(r *http.Request) logrus.FieldLogger {
	return logging.GetLogEntry(r)
}
