package dockergatusgenerator

import (
	"bytes"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
)

type Templates struct {
	config Config
}

func containerName(container container.Summary) string {
	if len(container.Names) == 0 {
		return container.ID
	}

	s, _ := strings.CutPrefix(container.Names[0], "/")
	return s
}

func getUrlFromTraefik(container Container) string {
	if container.Traefik == nil {
		return container.Gatus.Url
	}

	// TODO extract entrypoint port from config
	conf := container.Traefik
	if conf.HTTP != nil && len(conf.HTTP.Routers) > 0 {
		routerConfig := slices.Collect(maps.Values(conf.HTTP.Routers))[0]
		traefikRule := routerConfig.Rule
		extractHostRegexp := regexp.MustCompile("Host\\(`(.+?)`\\)")
		extractPathPrefixRegexp := regexp.MustCompile("PathPrefix\\(`(.+?)`\\)")
		hostMatch := extractHostRegexp.FindStringSubmatch(traefikRule)
		if len(hostMatch) < 2 {
			return container.Gatus.Url
		}
		host := hostMatch[1]
		pathPrefixMatch := extractPathPrefixRegexp.FindStringSubmatch(traefikRule)
		pathPrefix := ""
		if len(pathPrefixMatch) >= 2 {
			pathPrefix = strings.TrimRight(pathPrefixMatch[1], "/")
		}
		return fmt.Sprintf("https://%s%s%s", host, pathPrefix, container.Gatus.Path)
	}

	return container.Gatus.Url
}

func getUrlFromDocker(container Container) string {
	traefikPort := container.Gatus.Port
	scheme := container.Gatus.Scheme

	if traefikPort == "" && container.Traefik != nil && container.Traefik.HTTP != nil && len(container.Traefik.HTTP.Services) > 0 {
		serviceConfig := slices.Collect(maps.Values(container.Traefik.HTTP.Services))[0]
		if len(serviceConfig.LoadBalancer.Servers) > 0 {
			server := serviceConfig.LoadBalancer.Servers[0]
			traefikPort = server.Port
			if server.Scheme != "" {
				scheme = server.Scheme
			}
		}
	}

	if scheme == "" {
		scheme = "http"
	}

	port, _ := strconv.Atoi(traefikPort)
	if container.Inspect.NetworkSettings != nil {
		if container.Inspect.NetworkSettings.Ports != nil {
			ports := slices.Collect(maps.Keys(container.Inspect.NetworkSettings.Ports))
			selectedPortIndex := slices.IndexFunc(ports, func(p network.Port) bool { return int(p.Num()) == port })
			if selectedPortIndex == -1 {
				for _, pk := range ports {
					pv := container.Inspect.NetworkSettings.Ports[pk]
					selectedPortIndex = slices.IndexFunc(pv, func(p network.PortBinding) bool { return p.HostPort == strconv.FormatInt(int64(port), 10) })
					if selectedPortIndex != -1 {
						port = int(pk.Num())
						break
					}
				}
				if selectedPortIndex == -1 {
					if len(ports) > 0 {
						port = int(ports[0].Num())
					} else if port <= 0 {
						port = 80
					}
				}
			}
		}

		if container.Inspect.NetworkSettings.Networks != nil {
			for _, network := range container.Inspect.NetworkSettings.Networks {
				host := network.IPAddress.String()
				if len(network.Aliases) > 0 {
					host = network.Aliases[0]
				}
				return fmt.Sprintf("%s://%s:%d%s", scheme, host, port, container.Gatus.Path)
			}
		}
	}

	return container.Gatus.Url
}

func getUrlFromExposed(config Config, container Container) string {
	ip := config.GetIp()
	targetPort, _ := strconv.Atoi(container.Gatus.Port)
	scheme := container.Gatus.Scheme
	port := 80

	if targetPort == 0 && container.Traefik != nil && container.Traefik.HTTP != nil && len(container.Traefik.HTTP.Services) > 0 {
		serviceConfig := slices.Collect(maps.Values(container.Traefik.HTTP.Services))[0]
		if len(serviceConfig.LoadBalancer.Servers) > 0 {
			server := serviceConfig.LoadBalancer.Servers[0]
			if server.URL != "" {
				url, _ := url.Parse(server.URL)
				if url != nil {
					scheme = url.Scheme
					if url.Port() != "" {
						targetPort, _ = strconv.Atoi(url.Port())
					}
				}
			} else {
				targetPort, _ = strconv.Atoi(server.Port)
				if server.Scheme != "" {
					scheme = server.Scheme
				}
			}
		}
		port = targetPort
	}

	if container.Inspect.HostConfig != nil && container.Inspect.HostConfig.NetworkMode == "host" {
		if targetPort != 0 {
			port = targetPort
		}
	} else if container.Inspect.NetworkSettings != nil && container.Inspect.NetworkSettings.Ports != nil {
		ports := slices.SortedFunc(
			maps.Keys(container.Inspect.NetworkSettings.Ports),
			func(a network.Port, b network.Port) int {
				return int(a.Num()) - int(b.Num())
			},
		)
		daPort := slices.IndexFunc(ports, func(p network.Port) bool { return int(p.Num()) == targetPort })
		if daPort != -1 {
			portb := container.Inspect.NetworkSettings.Ports[ports[daPort]]
			if portb[0].HostIP.Is4() && portb[0].HostIP.IsPrivate() {
				ip = portb[0].HostIP.String()
			}
			port, _ = strconv.Atoi(portb[0].HostPort)
		} else {
			for _, portKey := range ports {
				portb := container.Inspect.NetworkSettings.Ports[portKey]
				if len(portb) > 0 {
					parsedPort, _ := strconv.Atoi(portb[0].HostPort)
					// specific scenario when using something like traefik-kop
					if targetPort != 0 && targetPort != parsedPort {
						continue
					}
					if portb[0].HostIP.Is4() && portb[0].HostIP.IsPrivate() {
						ip = portb[0].HostIP.String()
					}
					port = parsedPort
					break
				}
			}
		}
	}

	if ip == "" {
		return container.Gatus.Url
	}

	if scheme == "" {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s:%d%s", scheme, ip, port, container.Gatus.Path)
}

func funcMap(tpl *template.Template, config Config) template.FuncMap {
	return template.FuncMap{
		"containerName":     containerName,
		"getUrlFromDocker":  getUrlFromDocker,
		"getUrlFromExposed": func(container Container) string { return getUrlFromExposed(config, container) },
		"getUrlFromTraefik": getUrlFromTraefik,
		"include": func(name string, data any) (string, error) {
			buffer := bytes.NewBufferString("")
			err := tpl.ExecuteTemplate(buffer, name, data)
			if err != nil {
				return "", err
			}
			return buffer.String(), nil
		},
	}
}

func NewTemplates(config Config) *Templates {
	return &Templates{config: config}
}

func (t *Templates) Get(name string) (*template.Template, error) {
	var src string
	path := t.config.GetTemplateFile(name)
	if path != "" {
		bytes, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		src = string(bytes[:])
	} else {
		src = t.config.GetTemplate(name)
		if src == "" {
			return nil, fmt.Errorf("Template '%s' does not exist", name)
		}
	}

	tpl := template.New(name).Funcs(sprig.FuncMap())
	tpl, err := tpl.Funcs(funcMap(tpl, t.config)).Parse(src)
	if err != nil {
		return nil, err
	}

	moreTemplatesPath := t.config.GetTemplatesPath()
	if moreTemplatesPath != "" {
		tpl, err = tpl.ParseGlob(filepath.Join(moreTemplatesPath, "*.tpl"))
		if err != nil {
			return nil, err
		}
	}

	return tpl, nil
}

func (t *Templates) Render(templateName string, input TemplateInput) error {
	filePath := filepath.Join(t.config.GetFolderPath(), fmt.Sprintf("%s.yml", templateName))
	logger := slog.With(slog.String("template", templateName), slog.String("outputFile", filePath))
	logger.Debug("Preparing template")
	tpl, err := t.Get(templateName)
	if err != nil {
		logger.Debug("Failed preparing template", "err", err)
		return err
	}

	logger.Debug("Opening output file")
	stream, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		logger.Debug("Could not open file", "err", err)
		return err
	}

	logger.Info("Rendering template")
	defer stream.Close()
	err = tpl.Execute(stream, input)
	if err != nil {
		logger.Debug("Failed rendering file", "err", err)
		return err
	}

	logger.Info("Template rendered")
	return nil
}

func (t *Templates) RenderAll(containers TemplateInputs) error {
	errors := make([]string, 0)
	for templateName, input := range containers {
		err := t.Render(templateName, input)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed rendering template %s: %s", templateName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "\n"))
	}
	return nil
}
