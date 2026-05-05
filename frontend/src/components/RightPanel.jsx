import React, { useState, useEffect } from 'react';
import { CheckCircle, Loader2, AlertCircle, Info, Download, ExternalLink, Video, FileText, Music } from 'lucide-react';
import SubtitleCard from './SubtitleCard';
import { parseSrt, stringifySrt } from '../utils/srtParser';
import { getSubtitleTask, approveReview } from '../utils/api';

export default function RightPanel({ taskId, initialTaskData, subtitles, onUpdateSubtitle, renderSettings, onUpdateSetting, voiceSettings }) {
  const [activeTab, setActiveTab] = useState('subtitles');

  return (
    <div className="w-[380px] bg-white flex flex-col h-full border-l border-slate-200 shadow-sm z-10">
      {/* Tab headers */}
      <div className="flex p-2 gap-1 border-b border-slate-100 bg-slate-50/50">
        <button
          onClick={() => setActiveTab('config')}
          className={`flex-1 py-2 text-xs font-medium rounded-md transition-colors cursor-pointer ${
            activeTab === 'config'
              ? 'bg-white text-slate-800 shadow-sm border border-slate-200'
              : 'text-slate-500 hover:text-slate-700'
          }`}
        >
          Cấu hình xuất
        </button>
        <button
          onClick={() => setActiveTab('subtitles')}
          className={`flex-1 py-2 text-xs font-medium rounded-md transition-colors cursor-pointer ${
            activeTab === 'subtitles'
              ? 'bg-blue-600 text-white shadow-sm'
              : 'text-slate-500 hover:text-slate-700'
          }`}
        >
          Phụ đề
        </button>
      </div>

      {/* Tab content */}
      <div className="flex-1 overflow-y-auto">
        {activeTab === 'config' ? (
          <ConfigTab settings={renderSettings} onUpdate={onUpdateSetting} />
        ) : (
          <SubtitlesTab 
            taskId={taskId} 
            initialData={initialTaskData} 
            externalSubtitles={subtitles} 
            onUpdateExternal={onUpdateSubtitle} 
            renderSettings={renderSettings}
            voiceSettings={voiceSettings}
          />
        )}
      </div>
    </div>
  );
}

function SubtitlesTab({ taskId, initialData, externalSubtitles, onUpdateExternal, renderSettings, voiceSettings }) {
  const [taskData, setTaskData] = useState(initialData);
  const [taskStatus, setTaskStatus] = useState(initialData?.status || 'processing');
  const [loading, setLoading] = useState(!initialData && !externalSubtitles);
  const [error, setError] = useState(null);
  const [approving, setApproving] = useState(false);
  const [approveSuccess, setApproveSuccess] = useState(false);

  // Sync with props
  useEffect(() => {
    if (initialData) {
      setTaskData(initialData);
      setTaskStatus(initialData.status);
    }
  }, [initialData]);

  // If we have external subtitles, use them
  const subtitles = externalSubtitles || [];

  const handleUpdateSubtitle = (index, updatedSub) => {
    if (onUpdateExternal) {
      onUpdateExternal(index, updatedSub);
    }
  };

  const handleApprove = async () => {
    if (!taskId || subtitles.length === 0) return;

    setApproving(true);
    setApproveSuccess(false);
    try {
      const srtContent = stringifySrt(subtitles);
      await approveReview(taskId, srtContent, renderSettings, voiceSettings);
      setApproveSuccess(true);
      setTaskStatus('processing');
      setTimeout(() => setApproveSuccess(false), 3000);
    } catch (err) {
      setError('Duyệt thất bại: ' + err.message);
    } finally {
      setApproving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-3 text-slate-400 p-8">
        <Loader2 size={32} className="animate-spin" />
        <p className="text-sm">Đang tải dữ liệu...</p>
      </div>
    );
  }

  if (error && subtitles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-3 text-red-400 p-8">
        <AlertCircle size={32} />
        <p className="text-sm text-center">{error}</p>
      </div>
    );
  }

  if (taskStatus === 'processing') {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-3 text-blue-500 p-8">
        <Loader2 size={32} className="animate-spin" />
        <p className="text-sm text-center font-medium">
          {approveSuccess ? 'Đã duyệt! Đang xử lý TTS...' : 'Đang xử lý video...'}
        </p>
        <p className="text-xs text-slate-400 text-center">Vui lòng đợi trong giây lát</p>
      </div>
    );
  }

  if (taskStatus === 'success') {
    const speechUrl = taskData?.speech_download_url;
    const resultFiles = taskData?.subtitle_info || [];

    return (
      <div className="p-6 h-full flex flex-col">
        <div className="flex flex-col items-center justify-center gap-4 py-8 mb-6 bg-green-50 border border-green-100 rounded-2xl text-green-600">
          <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center shadow-inner">
            <CheckCircle size={32} className="text-green-600" />
          </div>
          <div className="text-center">
            <h3 className="font-bold text-lg">Hoàn thành!</h3>
            <p className="text-sm text-green-700/70">Dự án của bạn đã sẵn sàng</p>
          </div>
        </div>

        <div className="space-y-4">
          <h4 className="text-xs font-bold text-slate-400 uppercase tracking-widest px-1">Tải kết quả</h4>
          
          <div className="grid gap-2">
             {/* Final Video link if available */}
             <a 
              href={taskData?.video_url} 
              target="_blank"
              className="flex items-center justify-between p-3 bg-white border border-slate-200 rounded-xl hover:border-blue-300 hover:shadow-md transition-all group"
            >
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-blue-50 rounded-lg flex items-center justify-center text-blue-600 group-hover:bg-blue-600 group-hover:text-white transition-colors">
                  <Video size={18} />
                </div>
                <div>
                  <p className="text-sm font-semibold text-slate-700">Video kết quả</p>
                  <p className="text-[10px] text-slate-400">MP4 Format</p>
                </div>
              </div>
              <Download size={16} className="text-slate-300 group-hover:text-blue-500" />
            </a>

            {/* SRT Files */}
            {resultFiles.map((file, i) => (
              <a 
                key={i}
                href={file.download_url} 
                target="_blank"
                className="flex items-center justify-between p-3 bg-white border border-slate-200 rounded-xl hover:border-blue-300 hover:shadow-md transition-all group"
              >
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 bg-amber-50 rounded-lg flex items-center justify-center text-amber-600 group-hover:bg-amber-600 group-hover:text-white transition-colors">
                    <FileText size={18} />
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-slate-700">{file.name}</p>
                    <p className="text-[10px] text-slate-400">Subtitle File</p>
                  </div>
                </div>
                <Download size={16} className="text-slate-300 group-hover:text-amber-500" />
              </a>
            ))}

            {/* Speech File */}
            {speechUrl && (
              <a 
                href={speechUrl} 
                target="_blank"
                className="flex items-center justify-between p-3 bg-white border border-slate-200 rounded-xl hover:border-blue-300 hover:shadow-md transition-all group"
              >
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 bg-purple-50 rounded-lg flex items-center justify-center text-purple-600 group-hover:bg-purple-600 group-hover:text-white transition-colors">
                    <Music size={18} />
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-slate-700">Audio lồng tiếng</p>
                    <p className="text-[10px] text-slate-400">MP3 Format</p>
                  </div>
                </div>
                <Download size={16} className="text-slate-300 group-hover:text-purple-500" />
              </a>
            )}
          </div>
        </div>

        <div className="mt-auto pt-6">
          <button className="w-full py-3 bg-slate-900 text-white rounded-xl font-semibold text-sm hover:bg-slate-800 transition-colors shadow-lg shadow-slate-200 flex items-center justify-center gap-2">
            <ExternalLink size={16} /> Quay lại dự án
          </button>
        </div>
      </div>
    );
  }

  // Waiting review - show subtitle list
  return (
    <div className="p-4 flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="font-semibold text-sm">Phiên âm</h3>
          <p className="text-xs text-slate-500">{subtitles.length} đoạn</p>
        </div>
        <button
          onClick={() => {
            const srtContent = stringifySrt(subtitles);
            const blob = new Blob([srtContent], { type: 'text/srt;charset=utf-8' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `subtitles_${taskId || 'export'}.srt`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
          }}
          className="px-3 py-1.5 text-xs font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 rounded-md transition-colors cursor-pointer border border-blue-200"
        >
          ⬇ Tải SRT
        </button>
      </div>

      {/* Info banner */}
      <div className="mb-3 p-2.5 bg-blue-50 border border-blue-200 rounded-lg flex items-start gap-2">
        <Info size={14} className="text-blue-500 mt-0.5 shrink-0" />
        <p className="text-[11px] text-blue-700 leading-relaxed">
          Di chuột vào từng đoạn để hiện nút <strong>Sửa</strong>. Ấn <strong>Enter</strong> để lưu, <strong>Esc</strong> để hủy.
        </p>
      </div>

      {/* Subtitle list */}
      <div className="flex-1 overflow-y-auto space-y-3 pr-1">
        {subtitles.map((sub, index) => (
          <SubtitleCard
            key={sub.id}
            subtitle={sub}
            index={index}
            onUpdate={handleUpdateSubtitle}
          />
        ))}
      </div>

      {/* Approve button */}
      <div className="pt-4 border-t border-slate-100 mt-3">
        {error && (
          <p className="text-xs text-red-500 mb-2">{error}</p>
        )}
        <button
          onClick={handleApprove}
          disabled={approving}
          className="w-full py-2.5 bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white text-sm font-semibold rounded-lg transition-colors flex items-center justify-center gap-2 cursor-pointer shadow-sm shadow-green-200"
        >
          {approving ? (
            <>
              <Loader2 size={16} className="animate-spin" />
              Đang gửi...
            </>
          ) : (
            <>
              <CheckCircle size={16} />
              Duyệt Phụ Đề & Chạy TTS
            </>
          )}
        </button>
      </div>
    </div>
  );
}

function ConfigTab({ settings, onUpdate }) {
  const PRESETS = [
    { id: 'default', label: 'Mặc định', styles: { fontColor: '#FFFFFF', bgColor: '#00000000', isBold: false } },
    { id: 'karaoke-yellow', label: 'Karaoke Vàng', styles: { fontColor: '#FFFF00', borderColor: '#000000', borderWidth: 10, isBold: true } },
    { id: 'highlight-blue', label: 'Highlight Xanh', styles: { fontColor: '#FFFFFF', bgColor: '#0066FFCC', isBold: true } },
    { id: 'minimal', label: 'Tối giản', styles: { fontColor: '#FFFFFF', fontSize: 5, isBold: false } },
    { id: 'tiktok', label: 'TikTok', styles: { fontColor: '#FFFFFF', bgColor: '#FF0050', isBold: true, fontSize: 8 } },
    { id: 'tiktok-white', label: 'TikTok Trắng', styles: { fontColor: '#000000', bgColor: '#FFFFFF', isBold: true, fontSize: 8 } },
    { id: 'reels-yellow', label: 'Reels Vàng', styles: { fontColor: '#000000', bgColor: '#FFD700', isBold: true, fontSize: 8 } },
    { id: 'modern-bold', label: 'Đậm Hiện Đại', styles: { fontColor: '#FFFFFF', isBold: true, fontSize: 9, borderWidth: 15 } },
  ];

  const RATIOS = [
    { id: '16:9', label: 'YouTube (16:9)', desc: 'Chuẩn cho video YouTube ngang (1920x1080)' },
    { id: '9:16', label: 'TikTok (9:16)', desc: 'Chuẩn cho video TikTok dọc (1080x1920)' },
    { id: 'reels', label: 'Instagram Reels (9:16)', desc: 'Chuẩn cho Instagram Reels dọc (1080x1920)', val: '9:16' },
    { id: 'story', label: 'Story (9:16)', desc: 'Chuẩn cho Instagram/Facebook Story dọc (1080x1920)', val: '9:16' },
    { id: '1:1', label: 'Vuông (1:1)', desc: 'Video vuông cho mạng xã hội (1080x1080)' },
    { id: 'original', label: 'Gốc', desc: 'Giữ nguyên tỷ lệ gốc của video' },
    { id: 'custom', label: 'Tùy chỉnh', desc: 'Tự thiết lập kích thước' },
  ];

  const FIT_MODES = [
    { id: 'cover', label: 'Cover (Phủ kín)', desc: 'Phóng to video để phủ kín khung, cắt phần thừa' },
    { id: 'contain', label: 'Contain (Vừa khung)', desc: 'Thu nhỏ video vừa khung, thêm viền đen nếu cần' },
    { id: 'fill', label: 'Fill (Lấp đầy)', desc: 'Phóng to từ trung tâm để lấp đầy khung' },
    { id: 'stretch', label: 'Stretch (Kéo giãn)', desc: 'Kéo giãn video để vừa khung (có thể méo)' },
  ];

  const applyPreset = (preset) => {
    onUpdate('subtitleStyle', preset.id);
    Object.entries(preset.styles).forEach(([k, v]) => onUpdate(k, v));
  };

  return (
    <div className="p-5 space-y-8 bg-white overflow-y-auto h-full pb-20">
      {/* Video Ratio Section */}
      <section className="space-y-4">
        <div className="px-1">
          <h4 className="text-[14px] font-bold text-slate-800">Tỷ lệ khung hình (Video Ratio)</h4>
          <p className="text-[11px] text-slate-400 mt-1">Chuyển đổi video sang các định dạng khác nhau (YouTube, TikTok, Reels, v.v.)</p>
        </div>
        
        <div className="grid grid-cols-2 gap-2">
          {RATIOS.map(r => (
            <button
              key={r.id}
              onClick={() => onUpdate('videoRatio', r.val || r.id)}
              className={`p-3 text-left rounded-xl border transition-all ${
                (r.val || r.id) === settings.videoRatio 
                  ? 'bg-blue-50 border-blue-200 shadow-sm' 
                  : 'bg-white border-slate-100 hover:border-slate-200'
              }`}
            >
              <p className={`text-[11px] font-bold ${ (r.val || r.id) === settings.videoRatio ? 'text-blue-600' : 'text-slate-700'}`}>{r.label}</p>
              <p className="text-[9px] text-slate-400 mt-1 leading-tight">{r.desc}</p>
            </button>
          ))}
        </div>
      </section>

      {/* Fit Mode Section */}
      <section className="space-y-4">
        <h4 className="text-[14px] font-bold text-slate-800 px-1">Chế độ điều chỉnh</h4>
        <div className="space-y-2">
          {FIT_MODES.map(mode => (
            <label 
              key={mode.id} 
              className={`flex items-start gap-3 p-3 rounded-xl border transition-all cursor-pointer ${
                settings.fitMode === mode.id ? 'bg-blue-50 border-blue-200' : 'bg-white border-slate-100 hover:border-slate-200'
              }`}
            >
              <input 
                type="radio" 
                name="fitMode" 
                checked={settings.fitMode === mode.id} 
                onChange={() => onUpdate('fitMode', mode.id)} 
                className="mt-1 accent-blue-600" 
              />
              <div>
                <p className={`text-xs font-bold ${settings.fitMode === mode.id ? 'text-blue-600' : 'text-slate-700'}`}>{mode.label}</p>
                <p className="text-[10px] text-slate-400 mt-0.5">{mode.desc}</p>
              </div>
            </label>
          ))}
        </div>
      </section>

      {/* Volume Section */}
      <section className="space-y-3 pt-4 border-t border-slate-100">
        <div className="flex justify-between items-center px-1">
          <h4 className="text-[13px] font-bold text-slate-800">Âm lượng giọng gốc ({settings.originalVolume}%)</h4>
        </div>
        <div className="px-2">
          <input 
            type="range" min="0" max="100" step="5" 
            value={settings.originalVolume}
            onChange={(e) => onUpdate('originalVolume', parseInt(e.target.value))}
            className="w-full h-1.5 bg-slate-100 rounded-full appearance-none accent-slate-800 cursor-pointer"
          />
          <div className="flex justify-between mt-2 text-[10px] text-slate-400 font-medium">
            <span>Tắt (0%)</span>
            <span>Thấp (25%)</span>
            <span>Vừa (50%)</span>
            <span>Cao (100%)</span>
          </div>
        </div>
      </section>

      {/* Subtitle Style Presets */}
      <section className="space-y-4">
        <div className="flex items-center gap-2 px-1">
          <h4 className="text-[13px] font-bold text-slate-800">Kiểu phụ đề</h4>
          <span className="text-[10px] text-slate-400">(Xem trước có thể không sync)</span>
        </div>
        
        <div className="grid grid-cols-2 gap-2">
          {PRESETS.map(p => (
            <button
              key={p.id}
              onClick={() => applyPreset(p)}
              className={`py-3.5 text-xs font-medium rounded-xl border transition-all ${
                settings.subtitleStyle === p.id 
                  ? 'bg-blue-50 border-blue-200 text-blue-600 shadow-sm' 
                  : 'bg-white border-slate-100 text-slate-600 hover:border-slate-200'
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </section>

      {/* Language Visibility & Order */}
      <section className="space-y-4 pt-2 border-t border-slate-50">
        <h4 className="text-[13px] font-bold text-slate-800 px-1">Hiển thị ngôn ngữ</h4>
        <div className="flex bg-slate-50 p-1 rounded-xl border border-slate-100">
          {[
            { id: 'both', label: 'Cả hai' },
            { id: 'origin', label: 'Chỉ Gốc' },
            { id: 'translated', label: 'Chỉ Dịch' },
          ].map(opt => (
            <button
              key={opt.id}
              onClick={() => onUpdate('subtitleVisibility', opt.id)}
              className={`flex-1 py-2 text-xs font-bold rounded-lg transition-all ${
                settings.subtitleVisibility === opt.id ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-400 hover:text-slate-600'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>

        {settings.subtitleVisibility === 'both' && (
          <div className="space-y-2 px-1 animate-in fade-in slide-in-from-top-1 duration-200">
            <label className="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Thứ tự ưu tiên (Dòng trên)</label>
            <div className="grid grid-cols-2 gap-2">
              {[
                { id: 'translated-top', label: 'Dịch trên' },
                { id: 'origin-top', label: 'Gốc trên' },
              ].map(opt => (
                <button
                  key={opt.id}
                  onClick={() => onUpdate('subtitleOrder', opt.id)}
                  className={`py-2 text-[11px] font-bold rounded-lg border transition-all ${
                    settings.subtitleOrder === opt.id ? 'bg-blue-600 text-white border-blue-600' : 'bg-white text-slate-500 border-slate-200 hover:border-slate-300'
                  }`}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>
        )}
      </section>

      {/* Advanced Toggle */}
      <button 
        onClick={() => onUpdate('advanced', !settings.advanced)}
        className="flex items-center gap-1.5 text-xs font-bold text-blue-600 px-1 hover:text-blue-700 transition-colors"
      >
        {settings.advanced ? 'Ẩn tùy chỉnh nâng cao ↑' : 'Hiện tùy chỉnh nâng cao ↓'}
      </button>

      {/* Advanced Options */}
      {settings.advanced && (
        <div className="space-y-5 pt-2 animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-[11px] font-bold text-slate-500 uppercase">Font chữ</label>
              <select 
                value={settings.fontFamily}
                onChange={(e) => onUpdate('fontFamily', e.target.value)}
                className="w-full p-2 text-xs bg-slate-50 border border-slate-200 rounded-lg outline-none focus:border-blue-300"
              >
                <option>Helvetica</option>
                <option>Arial</option>
                <option>Roboto</option>
                <option>Inter</option>
                <option>Outfit</option>
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-[11px] font-bold text-slate-500 uppercase">Kích thước (%)</label>
              <input 
                type="number" step="0.1"
                value={settings.fontSize}
                onChange={(e) => {
                  const val = parseFloat(e.target.value);
                  onUpdate('fontSize', isNaN(val) ? 0 : val);
                }}
                className="w-full p-2 text-xs bg-slate-50 border border-slate-200 rounded-lg outline-none"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-[11px] font-bold text-slate-500 uppercase">Màu chữ</label>
              <div className="flex gap-2">
                <input type="color" value={settings.fontColor} onChange={(e) => onUpdate('fontColor', e.target.value)} className="w-8 h-8 rounded cursor-pointer" />
                <input type="text" value={settings.fontColor.toUpperCase()} className="flex-1 p-1 text-[10px] font-mono border border-slate-200 rounded" />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-[11px] font-bold text-slate-500 uppercase">Màu viền</label>
              <div className="flex gap-2">
                <input type="color" value={settings.borderColor} onChange={(e) => onUpdate('borderColor', e.target.value)} className="w-8 h-8 rounded cursor-pointer" />
                <input type="text" value={settings.borderColor.toUpperCase()} className="flex-1 p-1 text-[10px] font-mono border border-slate-200 rounded" />
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-[11px] font-bold text-slate-500 uppercase">Độ dày viền</label>
              <input 
                type="number" 
                value={settings.borderWidth} 
                onChange={(e) => {
                  const val = parseInt(e.target.value);
                  onUpdate('borderWidth', isNaN(val) ? 0 : val);
                }} 
                className="w-full p-2 text-xs bg-slate-50 border border-slate-200 rounded-lg" 
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-[11px] font-bold text-slate-500 uppercase">Padding nền</label>
              <input 
                type="number" 
                value={settings.bgPadding} 
                onChange={(e) => {
                  const val = parseInt(e.target.value);
                  onUpdate('bgPadding', isNaN(val) ? 0 : val);
                }} 
                className="w-full p-2 text-xs bg-slate-50 border border-slate-200 rounded-lg" 
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-[11px] font-bold text-slate-500 uppercase">Khoảng cách dưới</label>
              <input 
                type="number" 
                value={settings.bottomDistance} 
                onChange={(e) => {
                  const val = parseInt(e.target.value);
                  onUpdate('bottomDistance', isNaN(val) ? 0 : val);
                }} 
                className="w-full p-2 text-xs bg-slate-50 border border-slate-200 rounded-lg" 
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-[11px] font-bold text-slate-500 uppercase">Khoảng cách dòng</label>
              <input 
                type="number" step="0.1" 
                value={settings.lineSpacing} 
                onChange={(e) => {
                  const val = parseFloat(e.target.value);
                  onUpdate('lineSpacing', isNaN(val) ? 0 : val);
                }} 
                className="w-full p-2 text-xs bg-slate-50 border border-slate-200 rounded-lg" 
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-[11px] font-bold text-slate-500 uppercase">Chế độ hiển thị</label>
            <div className="space-y-2">
              {[
                { id: 'full', label: 'Đầy đủ', desc: 'Hiển thị toàn bộ câu cùng lúc' },
                { id: 'highlight', label: 'Đánh dấu', desc: 'Highlight từ đang nói' },
                { id: 'karaoke', label: 'Karaoke', desc: 'Tích lũy từng từ đến hết câu' },
              ].map(mode => (
                <label key={mode.id} className={`flex items-start gap-3 p-3 rounded-xl border transition-all cursor-pointer ${settings.displayMode === mode.id ? 'bg-blue-50 border-blue-200' : 'bg-white border-slate-100 hover:border-slate-200'}`}>
                  <input type="radio" name="displayMode" checked={settings.displayMode === mode.id} onChange={() => onUpdate('displayMode', mode.id)} className="mt-1 accent-blue-600" />
                  <div>
                    <p className="text-xs font-bold text-slate-700">{mode.label}</p>
                    <p className="text-[10px] text-slate-400">{mode.desc}</p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          <div className="flex items-center gap-2 py-2">
            <input type="checkbox" id="bold" checked={settings.isBold} onChange={(e) => onUpdate('isBold', e.target.checked)} className="w-4 h-4 rounded accent-blue-600" />
            <label htmlFor="bold" className="text-xs font-bold text-slate-600">Chữ đậm</label>
          </div>
        </div>
      )}
    </div>
  );
}
