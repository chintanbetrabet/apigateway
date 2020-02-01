package auth

import (
	"net/http"
)

func tugboatAuthMethod() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Tugboat auth called", http.StatusForbidden)
	}
}
