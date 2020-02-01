package types

import (
	"net/http"
)

type RoutesInterface interface {
	RouteNext(w http.ResponseWriter, r *http.Request)
	HandlerMethod() func(w http.ResponseWriter, r *http.Request)
}
