package dockergatusgenerator

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type logpair[K any] struct {
	s string
	l K
}

func TestGetLogLevel(t *testing.T) {
	config := EnvConfig{}

	defer os.Unsetenv(ENV_GATUS_LOG_LEVEL)

	values := []logpair[slog.Leveler]{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"Info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"Warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROr", slog.LevelError},
		{"", slog.LevelInfo},
		{"nani", slog.LevelInfo},
	}
	for _, value := range values {
		os.Setenv(ENV_GATUS_LOG_LEVEL, value.s)
		result := config.GetLogLevel()
		assert.Equal(t, value.l, result, "Input %s has resulted in %s instead of %s", value.s, result, value.l)
	}
}

func TestGetFolderPath(t *testing.T) {
	config := EnvConfig{}

	defer os.Unsetenv(ENV_GATUS_FOLDER_PATH)

	result := config.GetFolderPath()
	assert.Empty(t, result)

	os.Setenv(ENV_GATUS_FOLDER_PATH, "/test")
	result = config.GetFolderPath()
	assert.Equal(t, "/test", result)
}

func TestGetTemplate(t *testing.T) {
	config := EnvConfig{}

	key := fmt.Sprintf(ENV_GATUS_TEMPLATE_FMT, "TEST")
	defer os.Unsetenv(key)

	result := config.GetTemplate("test")
	assert.Empty(t, result)

	os.Setenv(key, "Hello {{ . }}")
	result = config.GetTemplate("test")
	assert.Equal(t, "Hello {{ . }}", result)
}

func TestGetTemplateFile(t *testing.T) {
	config := EnvConfig{}

	key := fmt.Sprintf(ENV_GATUS_TEMPLATE_FILE_FMT, "TEST")
	defer os.Unsetenv(key)

	result := config.GetTemplateFile("test")
	assert.Empty(t, result)

	os.Setenv(key, "/test.tpl")
	result = config.GetTemplateFile("test")
	assert.Equal(t, "/test.tpl", result)
}

func TestGetIp(t *testing.T) {
	config := EnvConfig{}

	defer os.Unsetenv(ENV_GATUS_IP)

	result := config.GetIp()
	assert.Empty(t, result)

	os.Setenv(ENV_GATUS_IP, "192.168.1.1")
	result = config.GetIp()
	assert.Equal(t, "192.168.1.1", result)
}

func TestGetTemplatesPath(t *testing.T) {
	config := EnvConfig{}

	defer os.Unsetenv(ENV_GATUS_TEMPLATES_PATH)

	result := config.GetTemplatesPath()
	assert.Empty(t, result)

	os.Setenv(ENV_GATUS_TEMPLATES_PATH, "/templates")
	result = config.GetTemplatesPath()
	assert.Equal(t, "/templates", result)
}

func TestGetHostname(t *testing.T) {
	config := EnvConfig{}

	defer os.Unsetenv(ENV_GATUS_HOSTNAME)

	result := config.GetHostname()
	assert.Empty(t, result)

	os.Setenv(ENV_GATUS_HOSTNAME, "templates.local")
	result = config.GetHostname()
	assert.Equal(t, "templates.local", result)
}
