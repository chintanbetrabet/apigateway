package headers

import (
	"fmt"
	"net/http"
)

func defaultReturnHeaders() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "added default return headers")
	}
}
