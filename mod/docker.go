package dockergatusgenerator

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/moby/moby/api/types/events"
	"github.com/moby/moby/client"
	traefikLabels "github.com/traefik/traefik/v3/pkg/config/label"
)

type DockerProvider struct {
	client  *client.Client
	context context.Context
	config  Config
}

func coalesce(args ...string) string {
	for _, a := range args {
		if a != "" {
			return a
		}
	}
	return ""
}

func NewDockerProvider(context context.Context, config Config) (*DockerProvider, error) {
	client, err := client.New(client.FromEnv, client.WithUserAgent("@melchor629/gatus-docker-generator"))
	if err != nil {
		return nil, err
	}

	return &DockerProvider{client, context, config}, nil
}

func (dp *DockerProvider) handleResult(containers client.ContainerListResult) (TemplateInputs, error) {
	containersPerTemplate := make(TemplateInputs)
	for _, container := range containers.Items {
		cname := containerName(container)
		logger := slog.With(slog.String("containerId", container.ID), slog.String("containerName", cname))
		templateListJoined, ok := container.Labels["gatus.template"]
		if !ok {
			logger.Debug("Container %s has no templates, ignoring")
			continue
		}

		inspect, err := dp.client.ContainerInspect(dp.context, container.ID, client.ContainerInspectOptions{})
		if err != nil {
			return nil, fmt.Errorf("Cannot get more information about the container: %s", err)
		}

		templateList := strings.Split(templateListJoined, ",")
		traefikConfig, err := traefikLabels.DecodeConfiguration(container.Labels)
		path, _ := container.Labels["gatus.path"]
		port, _ := container.Labels["gatus.port"]
		scheme, _ := container.Labels["gatus.scheme"]
		conditions, _ := container.Labels["gatus.conditions"]
		name, _ := container.Labels["gatus.name"]
		url, _ := container.Labels["gatus.url"]

		logger.Debug("Container selected for templates", "templates", strings.Join(templateList, ", "))
		for _, templateName := range templateList {
			tpath, _ := container.Labels[fmt.Sprintf("gatus.%s.path", templateName)]
			tport, _ := container.Labels[fmt.Sprintf("gatus.%s.port", templateName)]
			tscheme, _ := container.Labels[fmt.Sprintf("gatus.%s.scheme", templateName)]
			tconditions, _ := container.Labels[fmt.Sprintf("gatus.%s.conditions", templateName)]
			tname, _ := container.Labels[fmt.Sprintf("gatus.%s.name", templateName)]
			turl, _ := container.Labels[fmt.Sprintf("gatus.%s.url", templateName)]
			containersPerTemplate.Append(templateName, Container{
				ID:      container.ID,
				Name:    coalesce(tname, name, cname),
				Inspect: inspect.Container,
				Summary: container,
				Traefik: traefikConfig,
				Gatus: &GatusConfig{
					Path:       coalesce(tpath, path),
					Scheme:     coalesce(tscheme, scheme),
					Port:       coalesce(tport, port),
					Conditions: coalesce(tconditions, conditions),
					Url:        coalesce(turl, url),
				},
			})
		}
	}
	if len(containers.Items) == 0 {
		slog.Debug("No containers detected")
	}
	return containersPerTemplate.Finish(dp.config), nil
}

func (dp *DockerProvider) handle() (TemplateInputs, error) {
	time.Sleep(time.Duration(500) * time.Millisecond)
	slog.Debug("Retrieving containers from docker")
	containers, err := dp.client.ContainerList(dp.context, client.ContainerListOptions{
		All:     true,
		Filters: client.Filters{}.Add("label", "gatus.enable=true"),
	})
	if err != nil {
		return nil, err
	}

	return dp.handleResult(containers)
}

func (dp *DockerProvider) Iter(yield func(TemplateInputs, error) bool) {
	containersPerTemplate, err := dp.handle()
	if !yield(containersPerTemplate, err) {
		return
	}

	slog.Info("Starting docker watcher loop")
	watchResult := dp.client.Events(dp.context, client.EventsListOptions{
		Filters: client.Filters{}.Add("label", "gatus.enable=true").Add("event", "create", "update", "start", "destroy"),
	})
	go func() {
		err = <-watchResult.Err
	}()
	for message := range watchResult.Messages {
		if message.Type != events.ContainerEventType {
			continue
		}

		slog.Info("Received a change from docker events", "action", message.Action, "containerId", message.Actor.ID)
		containersPerTemplate, err := dp.handle()
		if !yield(containersPerTemplate, err) {
			return
		}
	}

	if err != nil {
		slog.ErrorContext(dp.context, "Watcher failed and has stopped", "err", err)
	} else {
		slog.Info("Docker watcher loop finished")
	}
}

func (dp *DockerProvider) Close() error {
	return dp.client.Close()
}
