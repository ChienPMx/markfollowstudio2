import React, { useState, useRef, useEffect, useMemo } from 'react';
import { timestampToSeconds, secondsToTimestamp } from '../utils/srtParser';

const PIXELS_PER_SECOND = 100; // Base zoom (1x)

export default function BottomTimeline({ taskData, subtitles = [], onUpdateSubtitle }) {
  const [zoom, setZoom] = useState(1);
  const containerRef = useRef(null);
  const scrollRef = useRef(null);
  
  const duration = taskData?.video_info?.duration || 30; // Default if not available
  const pxPerSec = PIXELS_PER_SECOND * zoom;
  const totalWidth = duration * pxPerSec;

  const handleZoom = (z) => setZoom(z);

  return (
    <div className="h-48 border-t border-slate-200 bg-white p-4 flex flex-col shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)] z-20">
      <div className="flex items-center justify-between mb-2">
        <h3 className="font-semibold text-sm">Timeline chỉnh sửa</h3>
        <div className="flex items-center gap-2 text-xs">
          <span className="text-slate-500 mr-2">Zoom:</span>
          {[0.25, 0.5, 1, 1.5, 2].map(z => (
            <button 
              key={z}
              onClick={() => handleZoom(z)}
              className={`px-2 py-1 rounded font-medium transition-colors ${zoom === z ? 'bg-blue-600 text-white shadow-sm' : 'bg-slate-100 hover:bg-slate-200 text-slate-600'}`}
            >
              {z}x
            </button>
          ))}
          <span className="ml-4 font-mono text-slate-500">0:00 / {Math.floor(duration / 60)}:{(duration % 60).toString().padStart(2, '0')}</span>
        </div>
      </div>

      <div 
        ref={scrollRef}
        className="flex-1 bg-slate-50 border border-slate-200 rounded-lg relative overflow-x-auto scrollbar-hide"
      >
        <div 
          className="relative h-full"
          style={{ width: `${totalWidth}px`, minWidth: '100%' }}
        >
          {/* Time markers */}
          <div className="absolute top-0 left-0 h-full w-full pointer-events-none opacity-20" style={{
            backgroundImage: `linear-gradient(90deg, #cbd5e1 1px, transparent 1px)`,
            backgroundSize: `${pxPerSec}px 100%`
          }}></div>

          {/* Subtitle segments */}
          <div className="absolute inset-0 pt-8 pb-4">
            {subtitles.map((sub, index) => (
              <TimelineSegment 
                key={sub.id || index}
                subtitle={sub}
                index={index}
                pxPerSec={pxPerSec}
                onUpdate={(newSub) => onUpdateSubtitle(index, newSub)}
              />
            ))}
          </div>
        </div>
      </div>
      
      <div className="flex items-center justify-between mt-2 text-[11px] text-slate-400">
        <div className="flex items-center gap-1">
          <span className="text-yellow-500">💡</span>
          <span>Kéo các cạnh của đoạn để điều chỉnh thời gian bắt đầu/kết thúc</span>
        </div>
        <span>{subtitles.length} đoạn</span>
      </div>
    </div>
  );
}

function TimelineSegment({ subtitle, index, pxPerSec, onUpdate }) {
  const start = timestampToSeconds(subtitle.startTime);
  const end = timestampToSeconds(subtitle.endTime);
  const width = (end - start) * pxPerSec;
  const left = start * pxPerSec;

  const handleDrag = (type, e) => {
    e.preventDefault();
    const startX = e.clientX;
    const initialStart = start;
    const initialEnd = end;

    const onMouseMove = (moveE) => {
      const deltaX = moveE.clientX - startX;
      const deltaTime = deltaX / pxPerSec;

      if (type === 'start') {
        const newStart = Math.max(0, Math.min(initialEnd - 0.1, initialStart + deltaTime));
        onUpdate({ 
          ...subtitle, 
          startTime: secondsToTimestamp(newStart) 
        });
      } else if (type === 'end') {
        const newEnd = Math.max(initialStart + 0.1, initialEnd + deltaTime);
        onUpdate({ 
          ...subtitle, 
          endTime: secondsToTimestamp(newEnd) 
        });
      } else if (type === 'move') {
        const newStart = Math.max(0, initialStart + deltaTime);
        const newEnd = newStart + (initialEnd - initialStart);
        onUpdate({
          ...subtitle,
          startTime: secondsToTimestamp(newStart),
          endTime: secondsToTimestamp(newEnd)
        });
      }
    };

    const onMouseUp = () => {
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
    };

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  };

  return (
    <div 
      className="absolute h-16 bg-blue-100/80 border border-blue-300 rounded-md shadow-sm group cursor-move select-none"
      style={{ 
        left: `${left}px`, 
        width: `${width}px`,
        transition: 'none' // Disable transition during drag
      }}
      onMouseDown={(e) => handleDrag('move', e)}
    >
      {/* Waveform placeholder */}
      <div className="absolute inset-0 flex items-center justify-center gap-0.5 opacity-30 p-2 pointer-events-none">
        {Array.from({ length: Math.floor(width / 5) }).map((_, i) => (
          <div key={i} className="w-0.5 bg-blue-500 rounded-full" style={{ height: `${20 + (i * 13) % 60}%` }}></div>
        ))}
      </div>

      {/* Label */}
      <div className="absolute inset-0 flex items-center justify-center px-2">
        <span className="text-[10px] font-bold text-blue-800 truncate">
          Đoạn {index + 1}: {subtitle.translated}
        </span>
      </div>

      {/* Resize handles */}
      <div 
        className="absolute left-0 top-0 bottom-0 w-2 hover:bg-blue-400 cursor-ew-resize rounded-l-md transition-colors"
        onMouseDown={(e) => { e.stopPropagation(); handleDrag('start', e); }}
      />
      <div 
        className="absolute right-0 top-0 bottom-0 w-2 hover:bg-blue-400 cursor-ew-resize rounded-r-md transition-colors"
        onMouseDown={(e) => { e.stopPropagation(); handleDrag('end', e); }}
      />

      {/* Tooltip on hover */}
      <div className="absolute -top-6 left-1/2 -translate-x-1/2 bg-slate-800 text-white text-[9px] px-1.5 py-0.5 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
        {subtitle.startTime.split(',')[0]} - {subtitle.endTime.split(',')[0]}
      </div>
    </div>
  );
}
