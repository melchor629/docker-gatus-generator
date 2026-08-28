package dockergatusgenerator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ReadsLabels(t *testing.T) {
	labels, err := DecodeTraefikLabels(map[string]string{
		"traefik.http.routers.TEST.rule":                        "Host(`example.com`)",
		"traefik.http.services.TEST.loadbalancer.server.port":   "123",
		"traefik.http.services.TEST.loadbalancer.server.scheme": "https",
		"traefik.http.services.TEST.loadbalancer.server.url":    "example.com",
	})

	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, labels.HTTP) {
		return
	}
	if !assert.NotNil(t, labels.HTTP.Routers) {
		return
	}
	if !assert.NotNil(t, labels.HTTP.Services) {
		return
	}
	if !assert.NotNil(t, labels.HTTP.Routers["TEST"]) {
		return
	}
	assert.Equal(t, "Host(`example.com`)", labels.HTTP.Routers["TEST"].Rule)
	if !assert.NotNil(t, labels.HTTP.Services["TEST"]) {
		return
	}
	if !assert.NotNil(t, labels.HTTP.Services["TEST"].LoadBalancer) {
		return
	}
	if !assert.NotNil(t, labels.HTTP.Services["TEST"].LoadBalancer.Servers) {
		return
	}
	assert.Equal(t, "example.com", labels.HTTP.Services["TEST"].LoadBalancer.Servers[0].URL)
	assert.Equal(t, "123", labels.HTTP.Services["TEST"].LoadBalancer.Servers[0].Port)
	assert.Equal(t, "https", labels.HTTP.Services["TEST"].LoadBalancer.Servers[0].Scheme)
}

func Test_ExtraLabels_Succeeds(t *testing.T) {
	labels, err := DecodeTraefikLabels(map[string]string{
		"gatus.enabled":     "!",
		"traefik.http.pato": "?",
	})

	assert.Nil(t, err)
	if assert.NotNil(t, labels) {
		assert.NotNil(t, labels.HTTP)
	}
}
