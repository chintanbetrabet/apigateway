package throttle

import (
	"fmt"
	"net/http"
)

func defaultThrottleMethod() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Called Basic throttle")
	}
}
