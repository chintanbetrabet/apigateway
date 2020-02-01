package auth

import (
	"net/http"
)

func sessionAuthMethod() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized for session", http.StatusForbidden)
	}
}
