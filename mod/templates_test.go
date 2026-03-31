package dockergatusgenerator

import (
	"fmt"
	"net/netip"
	"os"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func assertEmptyOutput(t *testing.T, tpls *Templates) {
	l, _ := os.ReadDir(tpls.config.GetFolderPath())
	assert.Empty(t, l)
}

func assertFileOutput(t *testing.T, tpls *Templates, name, contents string) {
	l, _ := os.ReadDir(tpls.config.GetFolderPath())
	assert.NotEmpty(t, l)
	assert.Equal(t, l[0].Name(), fmt.Sprintf("%s.yml", name))
	if contents != "" {
		c, _ := os.ReadFile(fmt.Sprintf("%s.yml", name))
		assert.Equal(t, contents, string(c[:]))
	}
}

func Test_Get_NonExistingTemplate(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH: t.TempDir(),
	})

	_, err := tpls.Get("test")
	assert.ErrorContains(t, err, "Template 'test' does not exist")
}

func Test_Get_IncorrectTemplate(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH: t.TempDir(),
		"GATUS_TEMPLATE_TEST": "{{ if . }}...{{ ",
	})

	_, err := tpls.Get("test")
	assert.ErrorContains(t, err, "test:1: unclosed action")
}

func Test_Get_NonExistingTemplateFile(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH:      t.TempDir(),
		"GATUS_TEMPLATE_FILE_TEST": "./aaaaaaaaaaaaaaaaaaaaaaaaaa.tpl",
	})

	_, err := tpls.Get("test")
	assert.ErrorContains(t, err, "no such file or directory")
}

func Test_Get_ExistingTemplateFile(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH:      t.TempDir(),
		"GATUS_TEMPLATE_FILE_TEST": "../templates/test.tpl",
	})

	tpl, err := tpls.Get("test")
	assert.Nil(t, err)
	assert.NotNil(t, tpl)
}

func Test_Render_NonExisting(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH: t.TempDir(),
	})

	err := tpls.Render("test", TemplateInput{})
	assert.ErrorContains(t, err, "Template 'test' does not exist")
	assertEmptyOutput(t, tpls)
}

func Test_Render_CannotOpenOutputFile(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH:      "./it/does/not/exist",
		"GATUS_TEMPLATE_FILE_TEST": "../templates/test.tpl",
	})

	err := tpls.Render("test", TemplateInput{})
	assert.ErrorContains(t, err, "no such file or directory")
	assertEmptyOutput(t, tpls)
}

func Test_Render_ErrorsExecutingTemplate(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH: t.TempDir(),
		"GATUS_TEMPLATE_TEST": "{{ range .}}.{{ end }}",
	})

	err := tpls.Render("test", TemplateInput{})
	assert.ErrorContains(t, err, "range can't iterate over {[]  }")
	assertFileOutput(t, tpls, "test", "")
}

func Test_Render_GeneratesOutput(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH:      t.TempDir(),
		"GATUS_TEMPLATE_FILE_TEST": "../templates/test.tpl",
	})

	err := tpls.Render("test", TemplateInput{})
	assert.Nil(t, err)
	assertFileOutput(t, tpls, "test", "")
}

func Test_RenderAll_Nothing(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH: t.TempDir(),
	})

	err := tpls.RenderAll(TemplateInputs{})
	assert.Nil(t, err)
	assertEmptyOutput(t, tpls)
}

func Test_RenderAll_SomethingFailing(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH: t.TempDir(),
	})

	err := tpls.RenderAll(TemplateInputs{"test": TemplateInput{}})
	assert.ErrorContains(t, err, "Template 'test' does not exist")
	assertEmptyOutput(t, tpls)
}

func Test_RenderAll_Works(t *testing.T) {
	tpls := NewTemplates(TestConfig{
		ENV_GATUS_FOLDER_PATH:      t.TempDir(),
		"GATUS_TEMPLATE_FILE_TEST": "../templates/test.tpl",
	})

	err := tpls.RenderAll(TemplateInputs{"test": TemplateInput{}})
	assert.Nil(t, err)
	assertFileOutput(t, tpls, "test", "")
}

func Test_containerName_WithoutNames(t *testing.T) {
	result := containerName(container.Summary{ID: "test"})
	assert.Equal(t, "test", result)
}

func Test_containerName_WithName(t *testing.T) {
	result := containerName(container.Summary{ID: "test", Names: []string{"/container"}})
	assert.Equal(t, "container", result)
}

func Test_getUrlFromTraefik_noTraefik(t *testing.T) {
	result := getUrlFromTraefik(Container{
		Gatus: &GatusConfig{Url: "https://example.com"},
	})
	assert.Equal(t, "https://example.com", result)
}

func Test_getUrlFromTraefik_noRouters(t *testing.T) {
	result := getUrlFromTraefik(Container{
		Gatus: &GatusConfig{Url: "https://example.com"},
		Traefik: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{},
		},
	})
	assert.Equal(t, "https://example.com", result)
}

func Test_getUrlFromTraefik_routerWithoutHost(t *testing.T) {
	result := getUrlFromTraefik(Container{
		Gatus: &GatusConfig{Url: "https://example.com"},
		Traefik: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Routers: map[string]*dynamic.Router{
					"test": {Rule: ""},
				},
			},
		},
	})
	assert.Equal(t, "https://example.com", result)
}

func Test_getUrlFromTraefik_routerWithHostButNoPathPrefix(t *testing.T) {
	result := getUrlFromTraefik(Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
		Traefik: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Routers: map[string]*dynamic.Router{
					"test": {Rule: "Host(`test.com`)"},
				},
			},
		},
	})
	assert.Equal(t, "https://test.com/health", result)
}

func Test_getUrlFromTraefik_routerWithHostAndPathPrefix(t *testing.T) {
	result := getUrlFromTraefik(Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
		Traefik: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Routers: map[string]*dynamic.Router{
					"test": {Rule: "Host(`test.com`) && PathPrefix(`/test`)"},
				},
			},
		},
	})
	assert.Equal(t, "https://test.com/test/health", result)
}

func Test_getUrlFromDocker_NoTraefikNoNetworks(t *testing.T) {
	result := getUrlFromDocker(Container{
		Gatus:   &GatusConfig{Url: "https://example.com", Path: "/health"},
		Inspect: container.InspectResponse{},
	})
	assert.Equal(t, "https://example.com", result)
}

func Test_getUrlFromDocker_NoTraefikButOneNetworkWithoutPortsNorAliases(t *testing.T) {
	result := getUrlFromDocker(Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health", Port: "1234", Scheme: "https"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"test": {IPAddress: netip.AddrFrom4([4]byte{192, 168, 1, 1})},
				},
			},
		},
	})
	assert.Equal(t, "https://192.168.1.1:1234/health", result)
}

func Test_getUrlFromDocker_NoTraefikButOneNetworkWithAliasWithoutPorts(t *testing.T) {
	result := getUrlFromDocker(Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health", Port: "1234", Scheme: "https"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"test": {
						Aliases:   []string{"test"},
						IPAddress: netip.AddrFrom4([4]byte{192, 168, 1, 1}),
					},
				},
			},
		},
	})
	assert.Equal(t, "https://test:1234/health", result)
}

func Test_getUrlFromDocker_NoTraefikButOneNetworkWithAliasAndSelectFromExposedPort(t *testing.T) {
	result := getUrlFromDocker(Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health", Scheme: "https"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"test": {
						Aliases:   []string{"test"},
						IPAddress: netip.AddrFrom4([4]byte{192, 168, 1, 1}),
					},
				},
				Ports: network.PortMap{
					network.MustParsePort("1234"): {},
				},
			},
		},
	})
	assert.Equal(t, "https://test:1234/health", result)
}

func Test_getUrlFromDocker_NoTraefikButOneNetworkWithAliasAndDeduceFromExposedPort(t *testing.T) {
	result := getUrlFromDocker(Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health", Port: "4321", Scheme: "https"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"test": {
						Aliases:   []string{"test"},
						IPAddress: netip.AddrFrom4([4]byte{192, 168, 1, 1}),
					},
				},
				Ports: network.PortMap{
					network.MustParsePort("1234"): {
						network.PortBinding{HostPort: "4321"},
					},
				},
			},
		},
	})
	assert.Equal(t, "https://test:1234/health", result)
}

func Test_getUrlFromDocker_WithTraefikAndOneNetworkWithAliasAndDeduceFromExposedPort(t *testing.T) {
	result := getUrlFromDocker(Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"test": {
						Aliases:   []string{"test"},
						IPAddress: netip.AddrFrom4([4]byte{192, 168, 1, 1}),
					},
				},
				Ports: network.PortMap{
					network.MustParsePort("1234"): {
						network.PortBinding{HostPort: "4321"},
					},
				},
			},
		},
		Traefik: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Services: map[string]*dynamic.Service{
					"test": {
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Servers: []dynamic.Server{{Port: "4321", Scheme: "https"}},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, "https://test:1234/health", result)
}

func Test_getUrlFromExposed_WithoutTraefkNorExposedPort(t *testing.T) {
	config := TestConfig{}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
	})
	assert.Equal(t, "https://example.com", result)
}

func Test_getUrlFromExposed_WithoutTraefkUsingHostMode(t *testing.T) {
	config := TestConfig{ENV_GATUS_IP: "10.10.10.1"}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health", Port: "1234"},
		Inspect: container.InspectResponse{
			HostConfig: &container.HostConfig{
				NetworkMode: "host",
			},
		},
	})
	assert.Equal(t, "http://10.10.10.1:1234/health", result)
}

func Test_getUrlFromExposed_WithoutTraefkUsingExposedPortWithoutHostIP(t *testing.T) {
	config := TestConfig{ENV_GATUS_IP: "10.10.10.1"}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Ports: network.PortMap{
					network.MustParsePort("1234"): []network.PortBinding{
						{HostPort: "1234"},
					},
				},
			},
		},
	})
	assert.Equal(t, "http://10.10.10.1:1234/health", result)
}

func Test_getUrlFromExposed_WithoutTraefkUsingExposedPortWithHostIP(t *testing.T) {
	config := TestConfig{ENV_GATUS_IP: "10.10.10.1"}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Ports: network.PortMap{
					network.MustParsePort("1234"): []network.PortBinding{
						{HostPort: "1234", HostIP: netip.AddrFrom4([4]byte{192, 168, 10, 1})},
					},
				},
			},
		},
	})
	assert.Equal(t, "http://192.168.10.1:1234/health", result)
}

func Test_getUrlFromExposed_WithoutTraefkUsingSelectedExposedPortWithoutHostIP(t *testing.T) {
	config := TestConfig{ENV_GATUS_IP: "10.10.10.1"}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health", Port: "4321"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Ports: network.PortMap{
					network.MustParsePort("1234"): []network.PortBinding{
						{HostPort: "4444", HostIP: netip.AddrFrom4([4]byte{192, 168, 10, 1})},
					},
					network.MustParsePort("4321"): []network.PortBinding{
						{HostPort: "1234"},
					},
				},
			},
		},
	})
	assert.Equal(t, "http://10.10.10.1:1234/health", result)
}

func Test_getUrlFromExposed_WithoutTraefkUsingSelectedExposedPortWithHostIP(t *testing.T) {
	config := TestConfig{ENV_GATUS_IP: "10.10.10.1"}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health", Port: "4321"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Ports: network.PortMap{
					network.MustParsePort("1234"): []network.PortBinding{
						{HostPort: "4444", HostIP: netip.AddrFrom4([4]byte{192, 168, 10, 1})},
					},
					network.MustParsePort("4321"): []network.PortBinding{
						{HostPort: "1234", HostIP: netip.AddrFrom4([4]byte{192, 168, 10, 1})},
					},
				},
			},
		},
	})
	assert.Equal(t, "http://192.168.10.1:1234/health", result)
}

func Test_getUrlFromExposed_WithTraefikButNoExposedPorts(t *testing.T) {
	config := TestConfig{ENV_GATUS_IP: "10.10.10.1"}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
		Traefik: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Services: map[string]*dynamic.Service{
					"test": {
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Servers: []dynamic.Server{{Port: "1234", Scheme: "https"}},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, "https://10.10.10.1:1234/health", result)
}

func Test_getUrlFromExposed_WithTraefikAndExposedPortButNoHostIP(t *testing.T) {
	config := TestConfig{ENV_GATUS_IP: "10.10.10.1"}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Ports: network.PortMap{
					network.MustParsePort("80"): []network.PortBinding{
						{HostPort: "8080"},
					},
					network.MustParsePort("4321"): []network.PortBinding{
						{HostPort: "1234"},
					},
				},
			},
		},
		Traefik: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Services: map[string]*dynamic.Service{
					"test": {
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Servers: []dynamic.Server{{Port: "1234"}},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, "http://10.10.10.1:1234/health", result)
}

func Test_getUrlFromExposed_WithTraefikAndExposedPortAndHostIP(t *testing.T) {
	config := TestConfig{ENV_GATUS_IP: "10.10.10.1"}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Ports: network.PortMap{
					network.MustParsePort("1234"): []network.PortBinding{
						{HostPort: "4444", HostIP: netip.AddrFrom4([4]byte{192, 168, 10, 1})},
					},
					network.MustParsePort("4321"): []network.PortBinding{
						{HostPort: "1234", HostIP: netip.AddrFrom4([4]byte{192, 168, 10, 1})},
					},
				},
			},
		},
		Traefik: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Services: map[string]*dynamic.Service{
					"test": {
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Servers: []dynamic.Server{{Port: "1234"}},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, "http://192.168.10.1:4444/health", result)
}

func Test_getUrlFromExposed_WithTraefikUrlAndExposedPortAndHostIP(t *testing.T) {
	config := TestConfig{ENV_GATUS_IP: "10.10.10.1"}
	result := getUrlFromExposed(config, Container{
		Gatus: &GatusConfig{Url: "https://example.com", Path: "/health"},
		Inspect: container.InspectResponse{
			NetworkSettings: &container.NetworkSettings{
				Ports: network.PortMap{
					network.MustParsePort("1234"): []network.PortBinding{
						{HostPort: "4444", HostIP: netip.AddrFrom4([4]byte{192, 168, 10, 1})},
					},
					network.MustParsePort("4321"): []network.PortBinding{
						{HostPort: "1234", HostIP: netip.AddrFrom4([4]byte{192, 168, 10, 1})},
					},
				},
			},
		},
		Traefik: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Services: map[string]*dynamic.Service{
					"test": {
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Servers: []dynamic.Server{{URL: "https://a:1234"}},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, "https://192.168.10.1:4444/health", result)
}
