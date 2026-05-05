/**
 * SRT Parser Utility
 * Converts SRT text to structured data and back.
 *
 * SRT format from backend (bilingual):
 * 1
 * 00:00:00,000 --> 00:00:01,139
 * Translated sentence (Vietnamese)
 * Original sentence (Chinese)
 *
 * (empty line)
 * 2
 * ...
 */

/**
 * Parse an SRT string into an array of subtitle objects.
 * @param {string} srtString - Raw SRT content from backend
 * @returns {Array<{id: number, startTime: string, endTime: string, translated: string, origin: string}>}
 */
export function parseSrt(srtString) {
  if (!srtString || typeof srtString !== 'string') return [];

  const subtitles = [];
  const blocks = srtString.trim().replace(/\r\n/g, '\n').split(/\n\n+/);

  for (const block of blocks) {
    const lines = block.trim().split('\n');
    if (lines.length < 3) continue;

    const id = parseInt(lines[0], 10);
    if (isNaN(id)) continue;

    const timeMatch = lines[1].match(
      /(\d{2}:\d{2}:\d{2}[,.:]\d{3})\s*-->\s*(\d{2}:\d{2}:\d{2}[,.:]\d{3})/
    );
    if (!timeMatch) continue;

    const startTime = timeMatch[1];
    const endTime = timeMatch[2];

    // All text lines after timestamp
    const textLines = lines.slice(2).filter((l) => l.trim() !== '');

    // If 2+ text lines: first = translated, rest joined = origin
    // If 1 text line: it's translated only
    let translated = '';
    let origin = '';

    if (textLines.length >= 2) {
      // Backend default (BilingualTranslationOnBottom) writes:
      // 1. Origin sentence
      // 2. Target sentence
      origin = textLines[0];
      translated = textLines.slice(1).join('\n');
    } else if (textLines.length === 1) {
      translated = textLines[0];
    }

    subtitles.push({ id, startTime, endTime, translated, origin });
  }

  return subtitles;
}

/**
 * Convert an array of subtitle objects back to an SRT string.
 * @param {Array<{id: number, startTime: string, endTime: string, translated: string, origin: string}>} subtitles
 * @returns {string} SRT formatted string
 */
export function stringifySrt(subtitles) {
  return subtitles
    .map((sub, index) => {
      const id = index + 1;
      const timeLine = `${sub.startTime} --> ${sub.endTime}`;
      const lines = [id, timeLine];
      if (sub.origin) {
        lines.push(sub.origin);
      }
      lines.push(sub.translated);
      return lines.join('\n');
    })
    .join('\n\n');
}

/**
 * Format a timestamp string like "00:00:01,139" into a short display format "0:01"
 */
export function formatTimestamp(timestamp) {
  if (!timestamp) return '0:00';
  const parts = timestamp.replace(',', '.').split(':');
  if (parts.length < 3) return '0:00';
  const hours = parseInt(parts[0], 10);
  const minutes = parseInt(parts[1], 10);
  const seconds = parseInt(parts[2], 10);
  const totalMinutes = hours * 60 + minutes;
  return `${totalMinutes}:${seconds.toString().padStart(2, '0')}`;
}

/**
 * Convert "00:00:01,139" to seconds (1.139)
 */
export function timestampToSeconds(timestamp) {
  if (!timestamp) return 0;
  const parts = timestamp.replace(',', '.').split(':');
  if (parts.length < 3) return 0;
  return (
    parseInt(parts[0], 10) * 3600 +
    parseInt(parts[1], 10) * 60 +
    parseFloat(parts[2])
  );
}

/**
 * Convert seconds (1.139) to "00:00:01,139"
 */
export function secondsToTimestamp(seconds) {
  const ms = Math.floor((seconds % 1) * 1000);
  const totalSeconds = Math.floor(seconds);
  const s = totalSeconds % 60;
  const totalMinutes = Math.floor(totalSeconds / 60);
  const m = totalMinutes % 60;
  const h = Math.floor(totalMinutes / 60);

  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')},${ms.toString().padStart(3, '0')}`;
}
