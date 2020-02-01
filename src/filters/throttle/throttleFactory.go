package throttle

import (
	"net/http"
)

func ThrottleFactory(name string) func(w http.ResponseWriter, r *http.Request) {
  return defaultThrottleMethod()
}
