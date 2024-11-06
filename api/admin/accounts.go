package admin

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/dhax/go-base/auth/pwdless"
	"github.com/dhax/go-base/database"
	validation "github.com/go-ozzo/ozzo-validation"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// The list of error types returned from account resource.
var (
	ErrAccountValidation = errors.New("account validation error")
)

// AccountStore defines database operations for account management.
type AccountStore interface {
	List(*database.AccountFilter) ([]pwdless.Account, int, error)
	Create(*pwdless.Account) error
	Get(id int) (*pwdless.Account, error)
	Update(*pwdless.Account) error
	Delete(*pwdless.Account) error
}

// NameStore defines database operations related to user names.
type NameStore interface {
	ListOnly() ([]pwdless.Account, error)
	Create(*pwdless.Account) error
	Get(id int) (*pwdless.Account, error)
	Update(*pwdless.Account) error
	Delete(*pwdless.Account) error
}

// NameAryStore defines database operations related to user names, but in array form.
type NameAryStore interface {
	ListOnly() ([]pwdless.Account, error)
}

// AccountResource implements account management handler.
type AccountResource struct {
	Store AccountStore
}

// NameResource implements user names management handler.
type NameResource struct {
	Store NameStore
}

// NameAryResource implements array user names management handler.
type NameAryResource struct {
	Store NameAryStore
}

// NewAccountResource creates and returns an account resource.
func NewAccountResource(store AccountStore) *AccountResource {
	return &AccountResource{
		Store: store,
	}
}

func NewNameResource(st NameStore) *NameResource {
	return &NameResource{
		Store: st,
	}
}

func NewNameAryResource(st NameAryStore) *NameAryResource {
	return &NameAryResource{
		Store: st,
	}
}

func (rs *AccountResource) router() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", rs.list)
	r.Post("/", rs.create)
	r.Route("/{accountID}", func(r chi.Router) {
		r.Use(rs.accountCtx)
		r.Get("/", rs.get)
		r.Put("/", rs.update)
		r.Delete("/", rs.delete)
	})
	return r
}

func (rs *NameResource) router() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", rs.listonly)
	r.Post("/", rs.createWithEnvelope)
	r.Get("/tes", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ngetes"))
	})
	r.Route("/{accountID}", func(r chi.Router) {
		r.Use(rs.accountCtx)
		r.Get("/", rs.getWithEnvelope)
		r.Put("/", rs.updateWithEnvelope)
		r.Delete("/", rs.deleteWithEnvelope)
	})

	return r
}

func (rs *NameAryResource) router() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", rs.listonly)

	return r
}

func (rs *AccountResource) accountCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "accountID"))
		if err != nil {
			render.Render(w, r, ErrBadRequest)
			return
		}
		account, err := rs.Store.Get(id)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), ctxAccount, account)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (rs *NameResource) accountCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "accountID"))
		if err != nil {
			render.Render(w, r, ErrBadRequest)
			return
		}
		account, err := rs.Store.Get(id)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), ctxAccount, account)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type accountRequest struct {
	*pwdless.Account
}

type nameRequest struct {
	Account *pwdless.Account `json:"account"`
}

func (d *accountRequest) Bind(r *http.Request) error {
	return nil
}

func (d *nameRequest) Bind(r *http.Request) error {
	return nil
}

type accountResponse struct {
	*pwdless.Account
}

type nameResponse struct {
	Account *pwdless.Account `json:"account"`
}

func newAccountResponse(a *pwdless.Account) *accountResponse {
	resp := &accountResponse{Account: a}
	return resp
}

func newNameResponse(a *pwdless.Account) *nameResponse {
	resp := &nameResponse{Account: a}

	return resp
}

type accountListResponse struct {
	Accounts *[]pwdless.Account `json:"accounts"`
	Names    []string           `json:"names"`
	Count    int                `json:"count"`
}

type nameListPreResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type nameListResponse struct {
	Names  []string              `json:"names"`
	Detail []nameListPreResponse `json:"detail"`
}

type nameAryListResponse struct {
	Name  []string `json:"name"`
	Email []string `json:"email"`
}

func newAccountListResponse(a *[]pwdless.Account, count int) *accountListResponse {
	names := make([]string, len(*a))
	for i, account := range *a {
		names[i] = account.Name
	}

	resp := &accountListResponse{
		Accounts: a,
		Names:    names,
		Count:    count,
	}
	return resp
}

func newNameListResponse(a *[]pwdless.Account) *nameListResponse {
	names := []string{}
	detail := []nameListPreResponse{}

	for _, acc := range *a {
		names = append(names, acc.Name)

		detail = append(detail, nameListPreResponse{
			Name:  acc.Name + " is the email",
			Email: acc.Email,
		})
	}

	resp := &nameListResponse{
		Names:  names,
		Detail: detail,
	}
	return resp
}

func newNameAryListResponse(a *[]pwdless.Account) *nameAryListResponse {
	names := []string{}
	emails := []string{}

	for _, acc := range *a {
		names = append(names, acc.Name)
		emails = append(emails, acc.Email)
	}

	resp := &nameAryListResponse{
		Name:  names,
		Email: emails,
	}

	return resp
}

func (rs *AccountResource) list(w http.ResponseWriter, r *http.Request) {
	f, err := database.NewAccountFilter(r.URL.Query())
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
	al, count, err := rs.Store.List(f)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
	render.Respond(w, r, newAccountListResponse(&al, count))
}

func (rs *NameResource) listonly(w http.ResponseWriter, r *http.Request) {
	names, err := rs.Store.ListOnly()
	if err != nil {
		log(r).Errorf("error %v", err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, newNameListResponse(&names))
}

func (rs *NameAryResource) listonly(w http.ResponseWriter, r *http.Request) {
	names, err := rs.Store.ListOnly()
	if err != nil {
		log(r).Errorf("error: %v", err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, newNameAryListResponse(&names))
}

func (rs *AccountResource) create(w http.ResponseWriter, r *http.Request) {
	data := &accountRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := rs.Store.Create(data.Account); err != nil {
		switch err := err.(type) {
		case validation.Errors:
			render.Render(w, r, ErrValidation(ErrAccountValidation, err))
			return
		}
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, newAccountResponse(data.Account))
}

func (rs *NameResource) createWithEnvelope(w http.ResponseWriter, r *http.Request) {
	data := &nameRequest{Account: &pwdless.Account{}}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := rs.Store.Create(data.Account); err != nil {
		switch err := err.(type) {
		case validation.Errors:
			render.Render(w, r, ErrValidation(ErrAccountValidation, err))
			return
		}
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, newAccountResponse(data.Account))
}

func (rs *AccountResource) get(w http.ResponseWriter, r *http.Request) {
	acc := r.Context().Value(ctxAccount).(*pwdless.Account)
	render.Respond(w, r, newAccountResponse(acc))
}

func (rs *NameResource) getWithEnvelope(w http.ResponseWriter, r *http.Request) {
	acc := r.Context().Value(ctxAccount).(*pwdless.Account)
	render.Respond(w, r, newNameResponse(acc))
}

func (rs *AccountResource) update(w http.ResponseWriter, r *http.Request) {
	acc := r.Context().Value(ctxAccount).(*pwdless.Account)
	data := &accountRequest{Account: acc}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := rs.Store.Update(acc); err != nil {
		switch err := err.(type) {
		case validation.Errors:
			render.Render(w, r, ErrValidation(ErrAccountValidation, err))
			return
		}
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, newAccountResponse(acc))
}

func (rs *NameResource) updateWithEnvelope(w http.ResponseWriter, r *http.Request) {
	acc := r.Context().Value(ctxAccount).(*pwdless.Account)
	data := &nameRequest{Account: acc}

	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := rs.Store.Update(acc); err != nil {
		switch err := err.(type) {
		case validation.Errors:
			render.Render(w, r, ErrValidation(ErrAccountValidation, err))
			return
		}

		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, newNameResponse(acc))
}

func (rs *AccountResource) delete(w http.ResponseWriter, r *http.Request) {
	acc := r.Context().Value(ctxAccount).(*pwdless.Account)
	if err := rs.Store.Delete(acc); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, http.NoBody)
}

func (rs *NameResource) deleteWithEnvelope(w http.ResponseWriter, r *http.Request) {
	acc := r.Context().Value(ctxAccount).(*pwdless.Account)
	if err := rs.Store.Delete(acc); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, newNameResponse(acc))
}
