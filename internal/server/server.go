package server

import (
	"context"
	"fmt"
	"markflow-studio/config"
	"markflow-studio/internal/router"
	"markflow-studio/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var BackEnd *http.Server

func StartBackend() error {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()
	router.SetupRouter(engine)
	BackEnd = &http.Server{
		Addr: fmt.Sprintf("%s:%d", config.Conf.Server.Host, config.Conf.Server.Port),
		Handler: engine,
	}
	log.GetLogger().Info("Service started", zap.String("host", config.Conf.Server.Host), zap.Int("port", config.Conf.Server.Port))
	// return engine.Run(fmt.Sprintf("%s:%d", config.Conf.Server.Host, config.Conf.Server.Port))
	err := BackEnd.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.GetLogger().Error("Failed to start service", zap.Error(err))
		return err
	}
	log.GetLogger().Info("Service closed")
	return nil
}

func StopBackend() error {
	if BackEnd == nil {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := BackEnd.Shutdown(ctx); err != nil {
		log.GetLogger().Error("Failed to close service", zap.Error(err))
		return err
	}
	BackEnd = nil
	log.GetLogger().Info("Service closed successfully")
	return nil
}
