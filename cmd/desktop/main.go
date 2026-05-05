package main

import (
	"go.uber.org/zap"
	"markflow-studio/config"
	"markflow-studio/internal/desktop"
	"markflow-studio/internal/server"
	"markflow-studio/log"
	"os"
)

func main() {
	log.InitLogger()
	defer log.GetLogger().Sync()

	if !config.LoadConfig() {
		// Ensure basic configuration exists
		if err := config.CheckConfig(); err != nil {
			log.GetLogger().Error("Failed to load config", zap.Error(err))
			os.Exit(1)
		}
	}
	go func() {
		if err := deps.CheckDependency(); err != nil {
			log.GetLogger().Error("Failed to prepare dependencies", zap.Error(err))
			os.Exit(1)
		}
		if err := server.StartBackend(); err != nil {
			log.GetLogger().Error("Failed to start backend service", zap.Error(err))
			os.Exit(1)
		}
	}()
	config.ConfigBackup = config.Conf
	desktop.Show()
}
