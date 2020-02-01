package routes

import (
	"filters"
	"log"
	"types"
)

func RoutesFactory(name string, beforeFilters []filters.Filter, afterFilters []filters.Filter, upstream types.Upstream) types.RoutesInterface {
	log.Println("Name is", name)
	switch name {
	case "/cluster-proxy":
		log.Println("return CP object with", beforeFilters, afterFilters)
		return &ClusterProxy{beforeFilters, afterFilters}
	case "/":
		log.Println("return tugboat", beforeFilters, afterFilters)
		return &Tugboat{beforeFilters, afterFilters}
	default:
		log.Println("return Custom object with: ", name, upstream)
		return &CustomRoute{name, beforeFilters, afterFilters, upstream}
	}
}
