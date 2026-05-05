package main

import (
	"go.uber.org/zap"
	"markflow-studio/config"
	"markflow-studio/internal/deps"
	"markflow-studio/internal/server"
	"markflow-studio/log"
	"os"
)

func main() {
	log.InitLogger()
	defer log.GetLogger().Sync()

	var err error
	if !config.LoadConfig() {
		return
	}

	if err = config.CheckConfig(); err != nil {
		log.GetLogger().Error("Failed to load config", zap.Error(err))
		return
	}

	if err = deps.CheckDependency(); err != nil {
		log.GetLogger().Error("Failed to prepare dependencies", zap.Error(err))
		return
	}
	if err = server.StartBackend(); err != nil {
		log.GetLogger().Error("Failed to start backend service", zap.Error(err))
		os.Exit(1)
	}
}
