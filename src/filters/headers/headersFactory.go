package headers

import (
	"net/http"
)

func HeadersFactory(name string) func(w http.ResponseWriter, r *http.Request) {
	switch name {
	case "tugboat":
		return tugboatReturnHeaders()
	default:
		return defaultReturnHeaders()
	}
}
