package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"routes"
	"types"
)

func start_server(port string, configuredRoutes []types.RouteConfig, upstreamsMap map[string]types.Upstream) {
	mux := http.NewServeMux()
	for _, route := range configuredRoutes {
		upstream := upstreamsMap[route.ForwardUpstream]
		mux.HandleFunc(route.Location, routes.HandlersFactory(route.Location, route.BeforeFilters, route.AfterFilters, upstream))
	}
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func main() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	config := &types.ServerConfig{}
	if err := config.Parse(data); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v", config)
	upstreamsMap := make(map[string]types.Upstream)
	for _, upstream := range config.Upstreams {
		upstreamsMap[upstream.Name] = upstream
	}
	start_server(config.Port, config.Routes, upstreamsMap)
}
