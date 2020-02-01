package routes

import (
	"filters"
	"log"
	"net/http"
)

type Tugboat struct {
	beforeFilters []filters.Filter
	afterFilters  []filters.Filter
}

func (route Tugboat) Print() string {
	return "Tugboat"
}

func (route Tugboat) RouteNext(w http.ResponseWriter, r *http.Request) {
	
}

func (route Tugboat) HandlerMethod() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Req: %s %s\n", r.Host, r.URL.Path)
		filters.PerformFilters(route.beforeFilters, w, r)
		route.RouteNext(w, r)
		filters.PerformFilters(route.afterFilters, w, r)
	}
}
