package routes

import (
	"filters"
	"fmt"
	"log"
	"net/http"
	"types"
)

type CustomRoute struct {
	route         string
	beforeFilters []filters.Filter
	afterFilters  []filters.Filter
	upstream      types.Upstream
}

func (route CustomRoute) Print() string {
	return "CustomRoute"
}

func (route CustomRoute) RouteNext(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Reached next route of Custom route:", route.route)
	fmt.Fprintln(w, "Fowarding to one of", len(route.upstream.Hosts), " upstreams among:")
	for i, host := range route.upstream.Hosts {
		host_output := fmt.Sprintf("%d.  %s:%d", i+1, host.Url, host.Port)
		fmt.Fprintln(w, host_output)
	}

}

func (route CustomRoute) HandlerMethod() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Req: %s %s\n", r.Host, r.URL.Path)
		filters.PerformFilters(route.beforeFilters, w, r)
		route.RouteNext(w, r)
		filters.PerformFilters(route.afterFilters, w, r)
	}
}
