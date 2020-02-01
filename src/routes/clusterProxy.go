package routes

import (
	"filters"
	"fmt"
	"log"
	"net/http"
)

type ClusterProxy struct {
	beforeFilters []filters.Filter
	afterFilters  []filters.Filter
}

func (route ClusterProxy) Print() string {
	return "ClusterProxy"
}

func (route ClusterProxy) RouteNext(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Reached next route of ClusterProxy")
}

func (route ClusterProxy) HandlerMethod() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Req: %s %s\n", r.Host, r.URL.Path)
		filters.PerformFilters(route.beforeFilters, w, r)
		route.RouteNext(w, r)
		filters.PerformFilters(route.afterFilters, w, r)
	}
}
