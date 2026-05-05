import React, { useRef, useState, useEffect, useMemo } from 'react';
import { Play, Pause, ChevronLeft, Trash2, Mic, FileText, Video, Volume2 } from 'lucide-react';
import { timestampToSeconds } from '../utils/srtParser';

export default function VideoPlayer({ taskData, subtitles = [], renderSettings, onUpdateSetting, onOpenVoiceModal, onBack }) {
  const videoRef = useRef(null);
  const containerRef = useRef(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [isDragging, setIsDragging] = useState(false);

  const videoUrl = taskData?.video_url;
  const title = taskData?.video_info?.title || "Dự án lồng tiếng";
  const taskId = taskData?.task_id;

  const currentSubtitle = useMemo(() => {
    return subtitles.find(sub => {
      const start = timestampToSeconds(sub.startTime);
      const end = timestampToSeconds(sub.endTime);
      return currentTime >= start && currentTime <= end;
    });
  }, [subtitles, currentTime]);

  const togglePlay = () => {
    if (videoRef.current) {
      if (isPlaying) {
        videoRef.current.pause();
      } else {
        videoRef.current.play();
      }
      setIsPlaying(!isPlaying);
    }
  };

  const handleTimeUpdate = () => {
    if (videoRef.current) {
      setCurrentTime(videoRef.current.currentTime);
    }
  };

  const handleLoadedMetadata = () => {
    if (videoRef.current) {
      setDuration(videoRef.current.duration);
    }
  };

  const formatTime = (time) => {
    const min = Math.floor(time / 60);
    const sec = Math.floor(time % 60);
    return `${min}:${sec.toString().padStart(2, '0')}`;
  };

  const handleSeek = (e) => {
    const percent = parseFloat(e.target.value);
    const newTime = (percent / 100) * duration;
    if (videoRef.current) {
      videoRef.current.currentTime = newTime;
      setCurrentTime(newTime);
    }
  };

  // Dragging logic for subtitles
  const handleMouseDown = (e) => {
    setIsDragging(true);
  };

  useEffect(() => {
    const handleMouseMove = (e) => {
      if (!isDragging || !containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const y = e.clientY - rect.top;
      const percentY = (1 - y / rect.height) * 100;
      const constrainedY = Math.max(5, Math.min(95, percentY));
      onUpdateSetting('bottomDistance', Math.round(constrainedY));
    };

    const handleMouseUp = () => {
      setIsDragging(false);
    };

    if (isDragging) {
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
    }
    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging]);

  // Style objects based on renderSettings
  const subtitleStyle = {
    bottom: `${renderSettings.bottomDistance}%`,
    fontFamily: renderSettings.fontFamily,
    fontSize: `${renderSettings.fontSize}%`, // This is relative to container height usually, but CSS needs a unit.
    // For preview, we'll use a hack to make % height work reasonably
    fontSize: `calc(${renderSettings.fontSize} * 0.4vh)`, 
    color: renderSettings.fontColor,
    fontWeight: renderSettings.isBold ? 'bold' : 'normal',
    WebkitTextStroke: `${(renderSettings.borderWidth || 0) * 0.1}px ${renderSettings.borderColor || '#000000'}`,
    lineHeight: renderSettings.lineSpacing || 1.2,
    backgroundColor: renderSettings.bgColor || 'transparent', // Hex RRGGBBAA works in most browsers
    padding: `${renderSettings.bgPadding * 0.05}em`,
    cursor: 'ns-resize'
  };

  const aspectClass = useMemo(() => {
    switch (renderSettings.videoRatio) {
      case '16:9': return 'aspect-video';
      case '9:16': return 'aspect-[9/16]';
      case '1:1': return 'aspect-square';
      default: return 'aspect-video';
    }
  }, [renderSettings.videoRatio]);

  return (
    <div className="flex-1 flex flex-col bg-slate-50 border-r border-slate-200">
      {/* ... header remains same ... */}
      <div className="h-14 border-b border-slate-200 bg-white flex items-center justify-between px-4">
        {/* ... */}
        <div className="flex items-center gap-3">
          <button 
            onClick={onBack}
            className="p-1 hover:bg-slate-100 rounded cursor-pointer"
          >
            <ChevronLeft size={20} />
          </button>
          <div>
            <h2 className="font-semibold text-sm">{title}</h2>
            <p className="text-xs text-slate-500 font-mono">ID: {taskId}</p>
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          <button className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-red-600 bg-red-50 hover:bg-red-100 rounded-md transition-colors cursor-pointer">
            <Trash2 size={14} /> Xóa
          </button>
          <button 
            onClick={onOpenVoiceModal}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-700 bg-white border border-slate-200 hover:bg-slate-50 rounded-md transition-colors cursor-pointer"
          >
            <Mic size={14} /> Giọng nói
          </button>
          <button className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-700 bg-white border border-slate-200 hover:bg-slate-50 rounded-md transition-colors cursor-pointer">
            <FileText size={14} /> SRT
          </button>
          <button className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors cursor-pointer shadow-sm">
            <Video size={14} /> Tạo video
          </button>
        </div>
      </div>

      <div className="flex-1 flex flex-col items-center justify-center p-6 bg-slate-100 overflow-hidden relative">
        <div 
          ref={containerRef}
          className={`w-full max-w-[600px] ${aspectClass} bg-black rounded-lg shadow-2xl relative overflow-hidden flex flex-col items-center justify-center group border border-slate-300 transition-all duration-300`}
        >
          {videoUrl ? (
            <video
              ref={videoRef}
              src={videoUrl}
              className="w-full h-full object-contain"
              onTimeUpdate={handleTimeUpdate}
              onLoadedMetadata={handleLoadedMetadata}
              onClick={togglePlay}
            />
          ) : (
            <div className="text-white opacity-50 text-sm">Chưa có video</div>
          )}
          
          {/* Subtitle overlay */}
          {currentSubtitle && (
            <div 
              className="absolute left-0 right-0 text-center px-8 pointer-events-auto select-none transition-all duration-75"
              style={{ bottom: subtitleStyle.bottom }}
              onMouseDown={handleMouseDown}
            >
              <div 
                className="inline-block rounded-lg backdrop-blur-sm shadow-xl"
                style={{
                  ...subtitleStyle,
                  bottom: 'auto' // Use parent's bottom
                }}
              >
                {renderSettings.subtitleVisibility === 'both' ? (
                  <>
                    {renderSettings.subtitleOrder === 'translated-top' ? (
                      <>
                        <p className="mb-0.5">{currentSubtitle.translated}</p>
                        {currentSubtitle.origin && (
                          <p className="opacity-80 border-t border-white/20 pt-1 mt-1 font-normal italic" style={{ fontSize: '0.8em' }}>
                            {currentSubtitle.origin}
                          </p>
                        )}
                      </>
                    ) : (
                      <>
                        {currentSubtitle.origin && (
                          <p className="mb-0.5">{currentSubtitle.origin}</p>
                        )}
                        <p className="opacity-80 border-t border-white/20 pt-1 mt-1 font-normal italic" style={{ fontSize: '0.8em' }}>
                          {currentSubtitle.translated}
                        </p>
                      </>
                    )}
                  </>
                ) : renderSettings.subtitleVisibility === 'translated' ? (
                  <p>{currentSubtitle.translated}</p>
                ) : (
                  <p>{currentSubtitle.origin}</p>
                )}
              </div>
              
              {/* Drag handle hint */}
              {isDragging && (
                <div className="absolute -top-6 left-1/2 -translate-x-1/2 bg-blue-600 text-white text-[10px] px-2 py-1 rounded shadow-lg whitespace-nowrap">
                  Vị trí: {renderSettings.bottomDistance}%
                </div>
              )}
            </div>
          )}
          
          {/* Overlay play button when paused */}
          {!isPlaying && videoUrl && (
            <button 
              onClick={togglePlay}
              className="absolute inset-0 m-auto w-16 h-16 rounded-full bg-black/40 text-white flex items-center justify-center hover:bg-black/60 transition-all backdrop-blur-sm border border-white/20"
            >
              <Play size={32} fill="currentColor" className="ml-1" />
            </button>
          )}
        </div>

        {/* Player controls */}
        <div className="w-full max-w-[600px] mt-6 flex flex-col gap-3 bg-white p-4 rounded-xl shadow-sm border border-slate-200">
          <div className="flex items-center gap-3 text-xs text-slate-500 font-medium">
            <span className="w-10">{formatTime(currentTime)}</span>
            <input 
              type="range"
              min="0"
              max="100"
              value={duration ? (currentTime / duration) * 100 : 0}
              onChange={handleSeek}
              className="flex-1 h-1.5 bg-slate-100 rounded-full appearance-none cursor-pointer accent-blue-600 hover:bg-slate-200 transition-colors"
            />
            <span className="w-10 text-right">{formatTime(duration)}</span>
          </div>
          
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <button 
                onClick={togglePlay}
                className="w-10 h-10 rounded-full bg-blue-600 text-white flex items-center justify-center hover:bg-blue-700 transition-colors shadow-md shadow-blue-100 cursor-pointer"
              >
                {isPlaying ? <Pause size={20} fill="currentColor" /> : <Play size={20} fill="currentColor" className="ml-1" />}
              </button>
              
              <div className="flex items-center gap-2 text-slate-400 hover:text-slate-600 transition-colors">
                <Volume2 size={18} />
                <div className="w-20 h-1 bg-slate-100 rounded-full relative">
                  <div className="absolute left-0 top-0 bottom-0 w-3/4 bg-slate-300 rounded-full"></div>
                </div>
              </div>
            </div>
            
            <div className="text-[11px] font-bold text-slate-400 tracking-widest uppercase">
              PREVIEW MODE
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
