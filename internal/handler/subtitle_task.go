package handler

import (
	"krillin-ai/internal/dto"
	"krillin-ai/internal/response"
	"krillin-ai/internal/service"
	"krillin-ai/internal/deps"
	"krillin-ai/log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h Handler) StartSubtitleTask(c *gin.Context) {
	var req dto.StartVideoSubtitleTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.GetLogger().Error("StartSubtitleTask ShouldBindJSON err", zap.Error(err))
		response.R(c, response.Response{
			Error: -1,
			Msg:   "Parameter error",
			Data:  nil,
		})
		return
	}

	// Check if configuration needs re-initialization
	if configUpdated {
		log.GetLogger().Info("Detected config update, re-initializing service...")
		deps.CheckDependency()
		h.Service = service.NewService()
		configUpdated = false
	}

	svc := h.Service

	data, err := svc.StartSubtitleTask(req)
	if err != nil {
		response.R(c, response.Response{
			Error: -1,
			Msg:   err.Error(),
			Data:  nil,
		})
		return
	}
	response.R(c, response.Response{
		Error: 0,
		Msg:   "Success",
		Data:  data,
	})
}

func (h Handler) GetSubtitleTask(c *gin.Context) {
	var req dto.GetVideoSubtitleTaskReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.R(c, response.Response{
			Error: -1,
			Msg:   "Parameter error",
			Data:  nil,
		})
		return
	}

	// Check if configuration needs re-initialization
	if configUpdated {
		log.GetLogger().Info("Detected config update, re-initializing service...")
		h.Service = service.NewService()
		configUpdated = false
	}

	svc := h.Service
	data, err := svc.GetTaskStatus(req)
	if err != nil {
		response.R(c, response.Response{
			Error: -1,
			Msg:   err.Error(),
			Data:  nil,
		})
		return
	}
	response.R(c, response.Response{
		Error: 0,
		Msg:   "Success",
		Data:  data,
	})
}

func (h Handler) ApproveReview(c *gin.Context) {
	var req dto.ApproveReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.GetLogger().Error("ApproveReview ShouldBindJSON err", zap.Error(err))
		response.R(c, response.Response{
			Error: -1,
			Msg:   "Parameter error",
			Data:  nil,
		})
		return
	}

	svc := h.Service
	if err := svc.ApproveReview(req); err != nil {
		log.GetLogger().Error("ApproveReview err", zap.Error(err))
		response.R(c, response.Response{
			Error: -1,
			Msg:   err.Error(),
			Data:  nil,
		})
		return
	}
	response.R(c, response.Response{
		Error: 0,
		Msg:   "Review approved, pipeline resuming",
		Data:  nil,
	})
}

func (h Handler) UploadFile(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.R(c, response.Response{
			Error: -1,
			Msg:   "Failed to get file",
			Data:  nil,
		})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		response.R(c, response.Response{
			Error: -1,
			Msg:   "No files uploaded",
			Data:  nil,
		})
		return
	}

	// Save each file
	var savedFiles []string
	for _, file := range files {
		savePath := "./uploads/" + file.Filename
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			response.R(c, response.Response{
				Error: -1,
				Msg:   "Failed to save file: " + file.Filename,
				Data:  nil,
			})
			return
		}
		savedFiles = append(savedFiles, "local:"+savePath)
	}

	response.R(c, response.Response{
		Error: 0,
		Msg:   "File upload successful",
		Data:  gin.H{"file_path": savedFiles},
	})
}

func (h Handler) DownloadFile(c *gin.Context) {
	requestedFile := c.Param("filepath")
	if requestedFile == "" {
		response.R(c, response.Response{
			Error: -1,
			Msg:   "File path is empty",
			Data:  nil,
		})
		return
	}

	localFilePath := filepath.Join(".", requestedFile)
	if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
		response.R(c, response.Response{
			Error: -1,
			Msg:   "File does not exist",
			Data:  nil,
		})
		return
	}
	c.FileAttachment(localFilePath, filepath.Base(localFilePath))
}
