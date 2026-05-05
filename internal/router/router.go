package router

import (
	"io/fs"
	"markflow-studio/internal/handler"
	"markflow-studio/static"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	api := r.Group("/api")

	hdl := handler.NewHandler()
	{
		api.POST("/capability/subtitleTask", hdl.StartSubtitleTask)
		api.GET("/capability/subtitleTask", hdl.GetSubtitleTask)
		api.POST("/capability/subtitleTask/review", hdl.ApproveReview)
		api.POST("/file", hdl.UploadFile)
		api.GET("/file/*filepath", hdl.DownloadFile)
		api.HEAD("/file/*filepath", hdl.DownloadFile)
		api.GET("/config", hdl.GetConfig)
		api.POST("/config", hdl.UpdateConfig)
	}

	// Redirect old static paths
	r.GET("/static", func(c *gin.Context) { c.Redirect(301, "/") })
	r.GET("/static/*any", func(c *gin.Context) { c.Redirect(301, "/") })

	// Serve static assets from dist/assets
	assetsFS, err := fs.Sub(static.EmbeddedFiles, "dist/assets")
	if err == nil {
		r.StaticFS("/assets", http.FS(assetsFS))
	}

	// Serve the main page at root and handle SPA routing
	r.NoRoute(func(c *gin.Context) {
		// Try to serve static file from dist first
		// (though assets are already handled above)
		
		// Fallback to index.html for SPA
		indexFile, err := static.EmbeddedFiles.ReadFile("dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Index not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexFile)
	})
}
