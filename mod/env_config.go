package dockergatusgenerator

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const (
	ENV_GATUS_LOG_LEVEL         = "GATUS_LOG_LEVEL"
	ENV_GATUS_FOLDER_PATH       = "GATUS_FOLDER"
	ENV_GATUS_TEMPLATE_FMT      = "GATUS_TEMPLATE_%s"
	ENV_GATUS_TEMPLATE_FILE_FMT = "GATUS_TEMPLATE_FILE_%s"
	ENV_GATUS_IP                = "GATUS_IP"
	ENV_GATUS_TEMPLATES_PATH    = "GATUS_TEMPLATES_PATH"
	ENV_GATUS_HOSTNAME          = "GATUS_HOSTNAME"
)

type EnvConfig struct{}

func (*EnvConfig) GetLogLevel() slog.Leveler {
	level := strings.ToLower(os.Getenv(ENV_GATUS_LOG_LEVEL))
	if level == "debug" {
		return slog.LevelDebug
	}

	if level == "warn" {
		return slog.LevelWarn
	}

	if level == "error" {
		return slog.LevelError
	}

	return slog.LevelInfo
}

func (*EnvConfig) GetFolderPath() string {
	return os.Getenv(ENV_GATUS_FOLDER_PATH)
}

func (*EnvConfig) GetTemplate(name string) string {
	return os.Getenv(fmt.Sprintf(ENV_GATUS_TEMPLATE_FMT, strings.ToUpper(name)))
}

func (*EnvConfig) GetTemplateFile(name string) string {
	return os.Getenv(fmt.Sprintf(ENV_GATUS_TEMPLATE_FILE_FMT, strings.ToUpper(name)))
}

func (*EnvConfig) GetIp() string {
	return os.Getenv(ENV_GATUS_IP)
}

func (*EnvConfig) GetTemplatesPath() string {
	return os.Getenv(ENV_GATUS_TEMPLATES_PATH)
}

func (*EnvConfig) GetHostname() string {
	return os.Getenv(ENV_GATUS_HOSTNAME)
}
