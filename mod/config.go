package dockergatusgenerator

import (
	"log/slog"

	"github.com/moby/moby/api/types/container"
	traefikDynamic "github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type GatusConfig struct {
	Path       string
	Scheme     string
	Port       string
	Conditions string
	Url        string
}

type Container struct {
	ID      string
	Name    string
	Summary container.Summary
	Inspect container.InspectResponse
	Traefik *traefikDynamic.Configuration
	Gatus   *GatusConfig
}

type TemplateInput struct {
	Containers []Container
	Ip         string
	Hostname   string
}

type Config interface {
	GetLogLevel() slog.Leveler
	GetFolderPath() string
	GetTemplate(name string) string
	GetTemplateFile(name string) string
	GetIp() string
	GetTemplatesPath() string
	GetHostname() string
}
