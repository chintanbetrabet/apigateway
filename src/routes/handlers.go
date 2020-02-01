package routes

import (
	"filters"
	"net/http"
	"types"
)

func HandlersFactory(name string, beforeFilters []filters.Filter, afterFilters []filters.Filter, upstream types.Upstream) func(w http.ResponseWriter, r *http.Request) {
	return RoutesFactory(name, beforeFilters, afterFilters, upstream).HandlerMethod()
}
