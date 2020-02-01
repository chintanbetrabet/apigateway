package filters

import (
	"filters/auth"
	"filters/headers"
	"filters/throttle"
	"fmt"
	"log"
	"net/http"
)

type Filter struct {
	Type     string `yaml:"type"`
	Strategy string `yaml:"strategy"`
}

func PerformFilters(routeFilters []Filter, w http.ResponseWriter, r *http.Request) {
	for _, filter := range routeFilters {
		log.Println("Called: ", filter.Type, filter.Strategy)
		perform_function := FiltersFactory(filter.Type, filter.Strategy)
		perform_function(w, r)
	}
}

type filterMethod func(w http.ResponseWriter, r *http.Request)

func FiltersFactory(filterType string, strategy string) filterMethod {
	switch filterType {
	case "auth":
		return auth.AuthFactory(strategy)
	case "throttle":
		return throttle.ThrottleFactory(strategy)
	case "headers":
		return headers.HeadersFactory(strategy)
	default:
		return func(w http.ResponseWriter, r *http.Request) {
			output := fmt.Sprintf("Filter: %s Not found!. %s", filterType, http.StatusText(http.StatusInternalServerError))
			http.Error(w, output,
				http.StatusInternalServerError)
		}
	}
}
