const API_BASE = '/api';

/**
 * Fetch the current status and data for a subtitle task.
 * @param {string} taskId
 * @returns {Promise<Object>} - The task data from the backend
 */
export async function getSubtitleTask(taskId) {
  const res = await fetch(`${API_BASE}/capability/subtitleTask?taskId=${encodeURIComponent(taskId)}`);
  const json = await res.json();
  if (json.error !== 0) {
    throw new Error(json.msg || 'Failed to fetch task');
  }
  return json.data;
}

/**
 * Submit reviewed/edited SRT content back to the backend.
 * @param {string} taskId
 * @param {string} editedSrtContent - The full SRT string after user edits
 * @returns {Promise<Object>}
 */
export async function approveReview(taskId, editedSrtContent, renderSettings, voiceSettings) {
  const res = await fetch(`${API_BASE}/capability/subtitleTask/review`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      task_id: taskId,
      edited_srt_content: editedSrtContent,
      render_settings: renderSettings ? {
        original_volume: renderSettings.originalVolume,
        subtitle_style: renderSettings.subtitleStyle,
        font_family: renderSettings.fontFamily,
        font_size: renderSettings.fontSize,
        font_color: renderSettings.fontColor,
        border_color: renderSettings.borderColor,
        border_width: renderSettings.borderWidth,
        bg_padding: renderSettings.bgPadding,
        bottom_distance: renderSettings.bottomDistance,
        line_spacing: renderSettings.lineSpacing,
        bg_color: renderSettings.bgColor,
        is_bold: renderSettings.isBold,
        display_mode: renderSettings.displayMode,
        highlight_color: renderSettings.highlightColor,
        max_words_per_line: renderSettings.maxWordsPerLine,
        video_ratio: renderSettings.videoRatio,
        fit_mode: renderSettings.fitMode,
      } : undefined,
      voice_settings: voiceSettings ? {
        voice_id: voiceSettings.voiceId,
        speed: voiceSettings.speed,
        emotion: voiceSettings.emotion,
      } : undefined,
    }),
  });
  const json = await res.json();
  if (json.error !== 0) {
    throw new Error(json.msg || 'Failed to approve review');
  }
  return json.data;
}

/**
 * Fetch the application config.
 * @returns {Promise<Object>}
 */
export async function getConfig() {
  const res = await fetch(`${API_BASE}/config`);
  const json = await res.json();
  if (json.error !== 0) {
    throw new Error(json.msg || 'Failed to fetch config');
  }
  return json.data;
}
