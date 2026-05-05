import React, { useState } from 'react';
import { Play, Pencil, Check, X } from 'lucide-react';
import { formatTimestamp } from '../utils/srtParser';

/**
 * A single subtitle card showing both origin + translated text,
 * with inline editing for the translated line.
 */
export default function SubtitleCard({ subtitle, index, onUpdate }) {
  const [isEditing, setIsEditing] = useState(false);
  const [editedTranslated, setEditedTranslated] = useState(subtitle.translated);

  const handleSave = () => {
    onUpdate(index, { ...subtitle, translated: editedTranslated });
    setIsEditing(false);
  };

  const handleCancel = () => {
    setEditedTranslated(subtitle.translated);
    setIsEditing(false);
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSave();
    }
    if (e.key === 'Escape') {
      handleCancel();
    }
  };

  const startDisplay = formatTimestamp(subtitle.startTime);
  const endDisplay = formatTimestamp(subtitle.endTime);

  return (
    <div className="p-3.5 bg-white border border-slate-200 rounded-lg shadow-sm hover:border-blue-300 transition-all duration-200 group">
      {/* Header: timestamp + actions */}
      <div className="flex items-center justify-between mb-2.5">
        <span className="text-[11px] font-mono text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded">
          {startDisplay} → {endDisplay}
        </span>
        <div className="flex items-center gap-2">
          <button
            className="text-slate-400 hover:text-blue-500 transition-colors cursor-pointer"
            title="Phát đoạn này"
          >
            <Play size={14} />
          </button>
          {!isEditing ? (
            <button
              onClick={() => setIsEditing(true)}
              className="flex items-center gap-1 text-xs text-blue-600 font-medium opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
            >
              <Pencil size={12} />
              Sửa
            </button>
          ) : (
            <div className="flex items-center gap-1">
              <button onClick={handleSave} className="text-green-600 hover:text-green-700 cursor-pointer" title="Lưu">
                <Check size={16} />
              </button>
              <button onClick={handleCancel} className="text-red-500 hover:text-red-600 cursor-pointer" title="Hủy">
                <X size={16} />
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Origin text (Chinese / source language) */}
      {subtitle.origin && (
        <div className="mb-2 pb-2 border-b border-slate-100">
          <span className="text-[10px] font-semibold text-slate-400 uppercase tracking-wide">Gốc</span>
          <p className="text-sm text-slate-700 mt-0.5 leading-relaxed">{subtitle.origin}</p>
        </div>
      )}

      {/* Translated text (Vietnamese / target language) */}
      <div>
        <span className="text-[10px] font-semibold text-blue-500 uppercase tracking-wide">Dịch</span>
        {isEditing ? (
          <textarea
            value={editedTranslated}
            onChange={(e) => setEditedTranslated(e.target.value)}
            onKeyDown={handleKeyDown}
            autoFocus
            rows={2}
            className="w-full mt-1 text-sm font-medium text-slate-800 border border-blue-300 rounded-md px-2 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-400 resize-none bg-blue-50/50"
          />
        ) : (
          <p className="text-sm font-medium text-slate-800 mt-0.5 leading-relaxed">
            {subtitle.translated || <span className="italic text-slate-400">(Trống)</span>}
          </p>
        )}
      </div>
    </div>
  );
}
