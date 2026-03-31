package dockergatusgenerator

import (
	"fmt"
	"log/slog"
	"strings"
)

type TestConfig map[string]string

func (t TestConfig) GetLogLevel() slog.Leveler {
	level := strings.ToLower(t[ENV_GATUS_LOG_LEVEL])
	if level == "debug" {
		return slog.LevelDebug
	}

	if level == "warn" {
		return slog.LevelWarn
	}

	if level == "error" {
		return slog.LevelWarn
	}

	return slog.LevelInfo
}

func (t TestConfig) GetFolderPath() string {
	return t[ENV_GATUS_FOLDER_PATH]
}

func (t TestConfig) GetTemplate(name string) string {
	return t[fmt.Sprintf(ENV_GATUS_TEMPLATE_FMT, strings.ToUpper(name))]
}

func (t TestConfig) GetTemplateFile(name string) string {
	return t[fmt.Sprintf(ENV_GATUS_TEMPLATE_FILE_FMT, strings.ToUpper(name))]
}

func (t TestConfig) GetIp() string {
	return t[ENV_GATUS_IP]
}

func (t TestConfig) GetTemplatesPath() string {
	return t[ENV_GATUS_TEMPLATES_PATH]
}

func (t TestConfig) GetHostname() string {
	return t[ENV_GATUS_HOSTNAME]
}
