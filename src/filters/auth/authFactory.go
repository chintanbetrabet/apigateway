package auth

import (
	"net/http"
)

func AuthFactory(name string) func(w http.ResponseWriter, r *http.Request) {
	switch name {
	case "session":
		return sessionAuthMethod()
	case "token":
		return tokenAuthMethod()
	case "tugboat":
		return tugboatAuthMethod()
	default:
		return defaultAuthMethod()
	}
}
