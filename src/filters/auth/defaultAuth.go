package auth

import (
	"fmt"
	"net/http"
)

func defaultAuthMethod() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Default AUTH")
	}
}
