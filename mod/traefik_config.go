package dockergatusgenerator

import (
	"github.com/traefik/paerser/parser"
)

// It is needed to maintain a copy of the models here, labels that are not mapped into these models
// will fail to decode
// SEE: https://github.com/traefik/traefik/blob/master/pkg/config/dynamic/http_config.go

type TraefikDynamicConfig struct {
	HTTP *TraefikDynamicHTTPConfiguration `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" export:"true"`
}

type TraefikDynamicHTTPConfiguration struct {
	Routers  map[string]*TraefikDynamicHttpRouter  `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty" export:"true"`
	Services map[string]*TraefikDynamicHttpService `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty" export:"true"`
}

type TraefikDynamicHttpRouter struct {
	EntryPoints                 []string                           `json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty" export:"true"`
	Middlewares                 []string                           `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	Service                     string                             `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	Rule                        string                             `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
	ParentRefs                  []string                           `json:"parentRefs,omitempty" toml:"parentRefs,omitempty" yaml:"parentRefs,omitempty" label:"-" export:"true"`
	Priority                    int                                `json:"priority,omitempty" toml:"priority,omitempty,omitzero" yaml:"priority,omitempty" export:"true"`
	TLS                         *RouterTLSConfig                   `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	Observability               *RouterObservabilityConfig         `json:"observability,omitempty" toml:"observability,omitempty" yaml:"observability,omitempty" export:"true"`
	DefaultRule                 bool                               `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`
	DeniedEncodedPathCharacters *RouterDeniedEncodedPathCharacters `json:"-" toml:"-" yaml:"-" label:"-" file:"-" kv:"-"`
}

type RouterTLSConfig struct {
	Options      string   `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty" export:"true"`
	CertResolver string   `json:"certResolver,omitempty" toml:"certResolver,omitempty" yaml:"certResolver,omitempty" export:"true"`
	Domains      []Domain `json:"domains,omitempty" toml:"domains,omitempty" yaml:"domains,omitempty" export:"true"`
}

type Domain struct {
	Main string   `description:"Default subject name." json:"main,omitempty" toml:"main,omitempty" yaml:"main,omitempty"`
	SANs []string `description:"Subject alternative names." json:"sans,omitempty" toml:"sans,omitempty" yaml:"sans,omitempty"`
}

type RouterObservabilityConfig struct {
	AccessLogs     *bool  `json:"accessLogs,omitempty" toml:"accessLogs,omitempty" yaml:"accessLogs,omitempty" export:"true"`
	Metrics        *bool  `json:"metrics,omitempty" toml:"metrics,omitempty" yaml:"metrics,omitempty" export:"true"`
	Tracing        *bool  `json:"tracing,omitempty" toml:"tracing,omitempty" yaml:"tracing,omitempty" export:"true"`
	TraceVerbosity string `json:"traceVerbosity,omitempty" toml:"traceVerbosity,omitempty" yaml:"traceVerbosity,omitempty" export:"true"`
}

type RouterDeniedEncodedPathCharacters struct {
	AllowEncodedSlash         bool `description:"Defines whether requests with encoded slash characters in the path are allowed." json:"allowEncodedSlash,omitempty" toml:"allowEncodedSlash,omitempty" yaml:"allowEncodedSlash,omitempty" export:"true"`
	AllowEncodedBackSlash     bool `description:"Defines whether requests with encoded back slash characters in the path are allowed." json:"allowEncodedBackSlash,omitempty" toml:"allowEncodedBackSlash,omitempty" yaml:"allowEncodedBackSlash,omitempty" export:"true"`
	AllowEncodedNullCharacter bool `description:"Defines whether requests with encoded null characters in the path are allowed." json:"allowEncodedNullCharacter,omitempty" toml:"allowEncodedNullCharacter,omitempty" yaml:"allowEncodedNullCharacter,omitempty" export:"true"`
	AllowEncodedSemicolon     bool `description:"Defines whether requests with encoded semicolon characters in the path are allowed." json:"allowEncodedSemicolon,omitempty" toml:"allowEncodedSemicolon,omitempty" yaml:"allowEncodedSemicolon,omitempty" export:"true"`
	AllowEncodedPercent       bool `description:"Defines whether requests with encoded percent characters in the path are allowed." json:"allowEncodedPercent,omitempty" toml:"allowEncodedPercent,omitempty" yaml:"allowEncodedPercent,omitempty" export:"true"`
	AllowEncodedQuestionMark  bool `description:"Defines whether requests with encoded question mark characters in the path are allowed." json:"allowEncodedQuestionMark,omitempty" toml:"allowEncodedQuestionMark,omitempty" yaml:"allowEncodedQuestionMark,omitempty" export:"true"`
	AllowEncodedHash          bool `description:"Defines whether requests with encoded hash characters in the path are allowed." json:"allowEncodedHash,omitempty" toml:"allowEncodedHash,omitempty" yaml:"allowEncodedHash,omitempty" export:"true"`
}

type TraefikDynamicHttpService struct {
	LoadBalancer *TraefikDynamicHttpServersLoadBalancer `json:"loadBalancer,omitempty" toml:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty" export:"true"`
}

type TraefikDynamicHttpServersLoadBalancer struct {
	Servers []TraefikDynamicHttpServer `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty" label-slice-as-struct:"server" export:"true"`
}

type TraefikDynamicHttpServer struct {
	URL    string `json:"url,omitempty" toml:"url,omitempty" yaml:"url,omitempty"`
	Scheme string `json:"-" toml:"-" yaml:"-" file:"-" kv:"-"`
	Port   string `json:"-" toml:"-" yaml:"-" file:"-" kv:"-"`
}

func DecodeTraefikLabels(labels map[string]string) (*TraefikDynamicConfig, error) {
	conf := &TraefikDynamicConfig{
		HTTP: &TraefikDynamicHTTPConfiguration{},
	}

	err := parser.Decode(labels, conf, parser.DefaultRootName, "traefik.http.routers", "traefik.http.services")
	if err != nil {
		return nil, err
	}

	return conf, nil
}
