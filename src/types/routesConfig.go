package types

import (
	"filters"
	"gopkg.in/yaml.v2"
)

type UpstreamHost struct {
	Url  string `yaml:"url"`
	Port int64  `yaml:"port"`
}

type Upstream struct {
	Name  string         `yaml:"name"`
	Hosts []UpstreamHost `yaml:"hosts"`
}

type RouteConfig struct {
	Name            string           `yaml:"name"`
	Location        string           `yaml:"location"`
	BeforeFilters   []filters.Filter `yaml:"beforeFilters"`
	AfterFilters    []filters.Filter `yaml:"afterFilters"`
	ForwardUpstream string           `yaml:"forwardUpstream"`
}

type ServerConfig struct {
	Routes          []RouteConfig
	Upstreams       []Upstream
	Port            string      `yaml:"port"`
	nginxDirectives []NginxFlag `yaml:"nginxDirectives"`
}

type NginxFlag struct {
	Name  string
	Value string
}

func (c *ServerConfig) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}
