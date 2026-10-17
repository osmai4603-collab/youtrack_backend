package middlewares

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	perms2 "youtrack_backend/internal/services/permissions_services"

	"github.com/go-chi/chi/v5"
)

type Scope = perms2.ScopeLevel

const (
	ScopeGlobal       = perms2.ScopeGlobal
	ScopeOrganization = perms2.ScopeOrganization
	ScopeProject      = perms2.ScopeProject
)

func EntityScoped(s Scope) bool {
	return s == ScopeOrganization || s == ScopeProject
}

type IDExtractor func(r *http.Request) (string, error)

type ScopeBinding struct {
	Scope   Scope
	Entity  perms2.EntityType
	Extract IDExtractor
}

func (b ScopeBinding) Validate(permission perms2.Permission) error {
	if !EntityScoped(b.Scope) {
		return nil
	}
	if b.Extract == nil {
		return fmt.Errorf("scope %q requires an id extractor (permission %q)", b.Scope, permission.ID)
	}
	return nil
}

func URLParam(name string) IDExtractor {
	return func(r *http.Request) (string, error) {
		v := chi.URLParam(r, name)
		if v == "" {
			return "", fmt.Errorf("missing %q in route", name)
		}
		return v, nil
	}
}

func QueryParam(name string) IDExtractor {
	return func(r *http.Request) (string, error) {
		v := r.URL.Query().Get(name)
		if v == "" {
			return "", fmt.Errorf("missing %q in query", name)
		}
		return v, nil
	}
}

func ParamOrQuery(name string) IDExtractor {
	return func(r *http.Request) (string, error) {
		if v := chi.URLParam(r, name); v != "" {
			return v, nil
		}
		if v := r.URL.Query().Get(name); v != "" {
			if _, err := url.PathUnescape(v); err != nil {
				return "", fmt.Errorf("invalid %q: %w", name, err)
			}
			return v, nil
		}
		return "", fmt.Errorf("missing %q in route or query", name)
	}
}

func ChainIDExtractors(extractors ...IDExtractor) IDExtractor {
	return func(r *http.Request) (string, error) {
		if len(extractors) == 0 {
			return "", fmt.Errorf("no id extractors were given")
		}
		reasons := make([]string, 0, len(extractors))
		for _, extract := range extractors {
			v, err := extract(r)
			if err == nil {
				return v, nil
			}
			reasons = append(reasons, err.Error())
		}
		return "", fmt.Errorf("no id matched: %s", strings.Join(reasons, "; "))
	}
}
