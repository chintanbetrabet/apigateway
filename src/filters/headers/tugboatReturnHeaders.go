package headers

import (
	"fmt"
	"net/http"
)

func tugboatReturnHeaders() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "added tugboat headers")
	}
}
