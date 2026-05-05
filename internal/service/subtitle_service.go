package service

import (
	"context"
	"errors"
	"fmt"
	"markflow-studio/internal/dto"
	"markflow-studio/internal/storage"
	"markflow-studio/internal/types"
	"markflow-studio/log"
	"markflow-studio/pkg/util"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/zap"
)

func (s Service) StartSubtitleTask(req dto.StartVideoSubtitleTaskReq) (*dto.StartVideoSubtitleTaskResData, error) {
	// Validate link
	if strings.Contains(req.Url, "youtube.com") {
		videoId, _ := util.GetYouTubeID(req.Url)
		if videoId == "" {
			return nil, fmt.Errorf("Invalid link")
		}
	}
	if strings.Contains(req.Url, "bilibili.com") {
		videoId := util.GetBilibiliVideoId(req.Url)
		if videoId == "" {
			return nil, fmt.Errorf("Invalid link")
		}
	}
	// Generate task ID
	seperates := strings.Split(req.Url, "/")
	taskId := fmt.Sprintf("%s_%s", util.SanitizePathName(string([]rune(strings.ReplaceAll(seperates[len(seperates)-1], " ", ""))[:16])), util.GenerateRandStringWithUpperLowerNum(4))
	taskId = strings.ReplaceAll(taskId, "=", "") // Equals sign affects ffmpeg processing
	taskId = strings.ReplaceAll(taskId, "?", "") // Question mark affects ffmpeg processing

	// Determine subtitle type based on input options
	var resultType types.SubtitleResultType
	if req.TargetLang == "none" {
		resultType = types.SubtitleResultTypeOriginOnly
	} else {
		if req.Bilingual == types.SubtitleTaskBilingualYes {
			if req.TranslationSubtitlePos == types.SubtitleTaskTranslationSubtitlePosTop {
				resultType = types.SubtitleResultTypeBilingualTranslationOnTop
			} else {
				resultType = types.SubtitleResultTypeBilingualTranslationOnBottom
			}
		} else {
			resultType = types.SubtitleResultTypeTargetOnly
		}
	}

	// Text replacement map
	replaceWordsMap := make(map[string]string)
	if len(req.Replace) > 0 {
		for _, replace := range req.Replace {
			beforeAfter := strings.Split(replace, "|")
			if len(beforeAfter) == 2 {
				replaceWordsMap[beforeAfter[0]] = beforeAfter[1]
			} else {
				log.GetLogger().Info("generateAudioSubtitles replace param length err", zap.Any("replace", replace), zap.Any("taskId", taskId))
			}
		}
	}

	var err error
	ctx := context.Background()

	// Create subtitle task folder
	taskBasePath := filepath.Join("./tasks", taskId)
	if _, err = os.Stat(taskBasePath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Join(taskBasePath, "output"), os.ModePerm)
		if err != nil {
			log.GetLogger().Error("StartVideoSubtitleTask MkdirAll err", zap.Any("req", req), zap.Error(err))
		}
	}

	// Create task; initialize review channel if review mode is enabled
	taskPtr := &types.SubtitleTask{
		TaskId:       taskId,
		VideoSrc:     req.Url,
		TtsVoiceCode: req.TtsVoiceCode,
		Status:       types.SubtitleTaskStatusProcessing,
	}
	if req.EnableReview {
		taskPtr.ReviewDoneCh = make(chan struct{})
	}
	storage.SubtitleTasks.Store(taskId, taskPtr)

	stepParam := types.SubtitleTaskStepParam{
		TaskId:                  taskId,
		TaskPtr:                 taskPtr,
		TaskBasePath:            taskBasePath,
		Link:                    req.Url,
		SubtitleResultType:      resultType,
		EnableModalFilter:       req.ModalFilter == types.SubtitleTaskModalFilterYes,
		EnableTts:               req.Tts == types.SubtitleTaskTtsYes,
		TtsVoiceCode:            req.TtsVoiceCode,
		ReplaceWordsMap:         replaceWordsMap,
		OriginLanguage:          types.StandardLanguageCode(req.OriginLanguage),
		TargetLanguage:          types.StandardLanguageCode(req.TargetLang),
		UserUILanguage:          types.StandardLanguageCode(req.Language),
		EmbedSubtitleVideoType:  req.EmbedSubtitleVideoType,
		VerticalVideoMajorTitle: req.VerticalMajorTitle,
		VerticalVideoMinorTitle: req.VerticalMinorTitle,
		MaxWordOneLine:          12,
	}
	if req.OriginLanguageWordOneLine != 0 {
		stepParam.MaxWordOneLine = req.OriginLanguageWordOneLine
	}

	log.GetLogger().Info("current task info", zap.String("taskId", taskId), zap.Any("param", stepParam))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				log.GetLogger().Error("autoVideoSubtitle panic", zap.Any("panic:", r), zap.Any("stack:", buf))
				stepParam.TaskPtr.Status = types.SubtitleTaskStatusFailed
			}
		}()

		log.GetLogger().Info("video subtitle start task", zap.String("taskId", taskId))

		err = s.linkToFile(ctx, &stepParam)
		if err != nil {
			log.GetLogger().Error("StartVideoSubtitleTask linkToFile err", zap.Any("req", req), zap.Error(err))
			stepParam.TaskPtr.Status = types.SubtitleTaskStatusFailed
			stepParam.TaskPtr.FailReason = err.Error()
			return
		}

		err = s.audioToSubtitle(ctx, &stepParam)
		if err != nil {
			log.GetLogger().Error("StartVideoSubtitleTask audioToSubtitle err", zap.Any("req", req), zap.Error(err))
			stepParam.TaskPtr.Status = types.SubtitleTaskStatusFailed
			stepParam.TaskPtr.FailReason = err.Error()
			return
		}

		// ── REVIEW PAUSE ──────────────────────────────────────────────────────
		if req.EnableReview && taskPtr.ReviewDoneCh != nil {
			// Choose which SRT file to expose for review
			reviewSrtPath := filepath.Join(taskBasePath, types.SubtitleTaskTargetLanguageSrtFileName)
			switch resultType {
			case types.SubtitleResultTypeBilingualTranslationOnTop,
				types.SubtitleResultTypeBilingualTranslationOnBottom:
				reviewSrtPath = filepath.Join(taskBasePath, types.SubtitleTaskBilingualSrtFileName)
			case types.SubtitleResultTypeOriginOnly:
				reviewSrtPath = filepath.Join(taskBasePath, types.SubtitleTaskOriginLanguageSrtFileName)
			}

			taskPtr.ReviewSrtPath = reviewSrtPath
			taskPtr.ProcessPct = 95
			taskPtr.Status = types.SubtitleTaskStatusWaitingReview
			log.GetLogger().Info("Task paused for review",
				zap.String("taskId", taskId),
				zap.String("srtPath", reviewSrtPath))

			// Block until ApproveReview closes the channel
			<-taskPtr.ReviewDoneCh
			log.GetLogger().Info("Review approved, resuming pipeline", zap.String("taskId", taskId))
			taskPtr.Status = types.SubtitleTaskStatusProcessing

			// Sync settings from taskPtr (updated by ApproveReview API) to stepParam
			stepParam.RenderSettings = taskPtr.RenderSettings
			stepParam.VoiceSettings = taskPtr.VoiceSettings
			if taskPtr.VoiceSettings != nil && taskPtr.VoiceSettings.VoiceId != "" {
				stepParam.TtsVoiceCode = taskPtr.VoiceSettings.VoiceId
			}

			// Ensure the TTS engine reads from the user-reviewed file
			stepParam.TtsSourceFilePath = taskPtr.ReviewSrtPath
			// Ensure we have a valid subtitle file path for embedding
			stepParam.BilingualSrtFilePath = taskPtr.ReviewSrtPath

			// If no embed type was selected, default to horizontal so the user gets a video with subs
			if stepParam.EmbedSubtitleVideoType == "none" || stepParam.EmbedSubtitleVideoType == "" {
				stepParam.EmbedSubtitleVideoType = "horizontal"
			}
		}
		// ─────────────────────────────────────────────────────────────────────

		err = s.srtFileToSpeech(ctx, &stepParam)
		if err != nil {
			log.GetLogger().Error("StartVideoSubtitleTask srtFileToSpeech err", zap.Any("req", req), zap.Error(err))
			stepParam.TaskPtr.Status = types.SubtitleTaskStatusFailed
			stepParam.TaskPtr.FailReason = err.Error()
			return
		}
		err = s.embedSubtitles(ctx, &stepParam)
		if err != nil {
			log.GetLogger().Error("StartVideoSubtitleTask embedSubtitles err", zap.Any("req", req), zap.Error(err))
			stepParam.TaskPtr.Status = types.SubtitleTaskStatusFailed
			stepParam.TaskPtr.FailReason = err.Error()
			return
		}
		err = s.uploadSubtitles(ctx, &stepParam)
		if err != nil {
			log.GetLogger().Error("StartVideoSubtitleTask uploadSubtitles err", zap.Any("req", req), zap.Error(err))
			stepParam.TaskPtr.Status = types.SubtitleTaskStatusFailed
			stepParam.TaskPtr.FailReason = err.Error()
			return
		}

		log.GetLogger().Info("video subtitle task end", zap.String("taskId", taskId))
	}()

	return &dto.StartVideoSubtitleTaskResData{
		TaskId: taskId,
	}, nil
}

func (s Service) GetTaskStatus(req dto.GetVideoSubtitleTaskReq) (*dto.GetVideoSubtitleTaskResData, error) {
	task, ok := storage.SubtitleTasks.Load(req.TaskId)
	if !ok || task == nil {
		return nil, errors.New("Task does not exist")
	}
	taskPtr := task.(*types.SubtitleTask)
	if taskPtr.Status == types.SubtitleTaskStatusFailed {
		return nil, fmt.Errorf("Task failed, reason: %s", taskPtr.FailReason)
	}

	// Map numeric status to string for the frontend
	statusStr := "processing"
	switch taskPtr.Status {
	case types.SubtitleTaskStatusWaitingReview:
		statusStr = "waiting_review"
	case types.SubtitleTaskStatusSuccess:
		statusStr = "success"
	case types.SubtitleTaskStatusFailed:
		statusStr = "failed"
	}

	res := &dto.GetVideoSubtitleTaskResData{
		TaskId:         taskPtr.TaskId,
		ProcessPercent: taskPtr.ProcessPct,
		Status:         statusStr,
		VideoInfo: &dto.VideoInfo{
			Title:                 taskPtr.Title,
			Description:           taskPtr.Description,
			TranslatedTitle:       taskPtr.TranslatedTitle,
			TranslatedDescription: taskPtr.TranslatedDescription,
		},
		SubtitleInfo: lo.Map(taskPtr.SubtitleInfos, func(item types.SubtitleInfo, _ int) *dto.SubtitleInfo {
			return &dto.SubtitleInfo{
				Name:        item.Name,
				DownloadUrl: item.DownloadUrl,
			}
		}),
		TargetLanguage:    taskPtr.TargetLanguage,
		TtsVoiceCode:      taskPtr.TtsVoiceCode,
		SpeechDownloadUrl: taskPtr.SpeechDownloadUrl,
	}

	// Set VideoUrl for frontend
	if strings.HasPrefix(taskPtr.VideoSrc, "local:") {
		res.VideoUrl = "/api/file/" + strings.TrimPrefix(taskPtr.VideoSrc, "local:")
	} else {
		res.VideoUrl = taskPtr.VideoSrc
	}

	// Provide SRT content when the pipeline is waiting for review
	if taskPtr.Status == types.SubtitleTaskStatusWaitingReview && taskPtr.ReviewSrtPath != "" {
		srtBytes, err := os.ReadFile(taskPtr.ReviewSrtPath)
		if err == nil {
			res.ReviewSrtContent = string(srtBytes)
		} else {
			log.GetLogger().Warn("GetTaskStatus: failed to read review SRT", zap.Error(err))
		}
	}

	return res, nil
}

// ApproveReview is called when the user finishes reviewing/editing the subtitles.
// It writes the (possibly edited) SRT back to disk and unblocks the pipeline goroutine.
func (s Service) ApproveReview(req dto.ApproveReviewReq) error {
	task, ok := storage.SubtitleTasks.Load(req.TaskId)
	if !ok || task == nil {
		return errors.New("Task does not exist")
	}
	taskPtr := task.(*types.SubtitleTask)
	if taskPtr.Status != types.SubtitleTaskStatusWaitingReview {
		return errors.New("Task is not in review state")
	}
	if taskPtr.ReviewDoneCh == nil {
		return errors.New("Review channel not initialized")
	}

	// Persist the user's edits before resuming
	if req.EditedSrtContent != "" && taskPtr.ReviewSrtPath != "" {
		if err := os.WriteFile(taskPtr.ReviewSrtPath, []byte(req.EditedSrtContent), 0644); err != nil {
			log.GetLogger().Error("ApproveReview WriteFile err", zap.Error(err))
			return fmt.Errorf("Failed to save edited subtitles: %w", err)
		}
		log.GetLogger().Info("ApproveReview: edited SRT saved", zap.String("taskId", req.TaskId))
	}

	// Persist RenderSettings
	if req.RenderSettings != nil {
		taskPtr.RenderSettings = &types.RenderSettings{
			OriginalVolume:  req.RenderSettings.OriginalVolume,
			SubtitleStyle:   req.RenderSettings.SubtitleStyle,
			FontFamily:      req.RenderSettings.FontFamily,
			FontSize:        req.RenderSettings.FontSize,
			FontColor:       req.RenderSettings.FontColor,
			BorderColor:     req.RenderSettings.BorderColor,
			BorderWidth:     req.RenderSettings.BorderWidth,
			BgPadding:       req.RenderSettings.BgPadding,
			BottomDistance:  req.RenderSettings.BottomDistance,
			LineSpacing:     req.RenderSettings.LineSpacing,
			BgColor:         req.RenderSettings.BgColor,
			IsBold:          req.RenderSettings.IsBold,
			DisplayMode:     req.RenderSettings.DisplayMode,
			HighlightColor:  req.RenderSettings.HighlightColor,
			MaxWordsPerLine: req.RenderSettings.MaxWordsPerLine,
			VideoRatio:      req.RenderSettings.VideoRatio,
			FitMode:         req.RenderSettings.FitMode,
		}
	}

	// Persist VoiceSettings
	if req.VoiceSettings != nil {
		taskPtr.VoiceSettings = &types.VoiceSettings{
			VoiceId: req.VoiceSettings.VoiceId,
			Speed:   req.VoiceSettings.Speed,
			Emotion: req.VoiceSettings.Emotion,
		}
	}

	// Unblock the pipeline
	close(taskPtr.ReviewDoneCh)
	taskPtr.ReviewDoneCh = nil // Prevent double-close
	return nil
}
