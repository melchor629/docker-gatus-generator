package main

import (
	"context"
	"log/slog"
	"os"

	mod "github.com/melchor629/docker-gatus-generator/mod"
)

func main() {
	config := &mod.EnvConfig{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: config.GetLogLevel()}))
	slog.SetDefault(logger)

	logger.Info("Checking gatus folder")
	gatusFolder := config.GetFolderPath()
	if gatusFolder == "" {
		logger.Error("Gatus folder setting is empty, please fill", "setting", mod.ENV_GATUS_FOLDER_PATH)
		os.Exit(1)
	}
	if _, err := os.Stat(gatusFolder); err != nil {
		logger.Error("Gatus folder does not exist or is not accessible", "err", err)
		os.Exit(1)
	}

	logger.Info("Starting docker API client")
	ctx := context.Background()
	apiClient, err := mod.NewDockerProvider(ctx, config)
	if err != nil {
		logger.ErrorContext(ctx, "Cannot create connection to docker", "err", err)
		os.Exit(1)
	}
	defer apiClient.Close()

	logger.Info("Starting container watcher")
	templates := mod.NewTemplates(config)
	for result, err := range apiClient.Iter {
		if err != nil {
			logger.ErrorContext(ctx, "Error in container watcher", "err", err)
		}

		err = templates.RenderAll(result)
		if err != nil {
			logger.ErrorContext(ctx, "Error rendering templates", "err", err)
		}
	}
}
