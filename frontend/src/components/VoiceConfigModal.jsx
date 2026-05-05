import React, { useState } from 'react';
import { X, Search, Play, Check } from 'lucide-react';

const VOICES = [
  { id: 'alloy', name: 'Alloy (OpenAI)', gender: 'Nữ/Nam', color: 'bg-blue-100 text-blue-600', provider: 'openai' },
  { id: 'echo', name: 'Echo (OpenAI)', gender: 'Nam', color: 'bg-blue-100 text-blue-600', provider: 'openai' },
  { id: 'fable', name: 'Fable (OpenAI)', gender: 'Nữ/Nam', color: 'bg-pink-100 text-pink-600', provider: 'openai' },
  { id: 'onyx', name: 'Onyx (OpenAI)', gender: 'Nam', color: 'bg-blue-100 text-blue-600', provider: 'openai' },
  { id: 'nova', name: 'Nova (OpenAI)', gender: 'Nữ', color: 'bg-pink-100 text-pink-600', provider: 'openai' },
  { id: 'shimmer', name: 'Shimmer (OpenAI)', gender: 'Nữ', color: 'bg-pink-100 text-pink-600', provider: 'openai' },
  { id: 'custom', name: 'Tùy chỉnh (VClip)', gender: 'ID', color: 'bg-purple-100 text-purple-600', provider: 'vclip' },
];

export default function VoiceConfigModal({ isOpen, onClose, settings, onSave }) {
  const [localSettings, setLocalSettings] = useState(settings);
  const [search, setSearch] = useState('');
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [voices, setVoices] = useState(VOICES);

  React.useEffect(() => {
    if (isOpen) {
      fetch('/api/config')
        .then(res => res.json())
        .then(json => {
          if (json.error === 0 && json.data?.tts?.voices) {
            setVoices([...VOICES, ...json.data.tts.voices.map(v => ({
              id: v.id,
              name: v.name,
              gender: 'Sưu tầm',
              color: 'bg-green-100 text-green-600',
              provider: 'custom'
            }))]);
          }
        });
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const filteredVoices = voices.filter(v => 
    v.name.toLowerCase().includes(search.toLowerCase())
  );

  const selectedVoice = voices.find(v => v.id === localSettings.voiceId) || voices[0];

  const handleSave = () => {
    let finalSettings = { ...localSettings };
    if (localSettings.voiceId === 'custom' && localSettings.customVoiceId) {
      finalSettings.voiceId = localSettings.customVoiceId;
    }
    onSave(finalSettings);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-white w-full max-w-[500px] rounded-3xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-slate-100">
          <h2 className="text-xl font-bold text-slate-800 tracking-tight">Cấu hình giọng nói</h2>
          <button onClick={onClose} className="p-2 hover:bg-slate-100 rounded-full text-slate-400 transition-colors">
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-8 overflow-y-auto max-h-[70vh]">
          {/* Voice Selection */}
          <div className="space-y-3">
            <label className="text-sm font-bold text-slate-700">Chọn giọng nói</label>
            <div className="relative">
              <button 
                onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                className="w-full flex items-center justify-between p-4 bg-slate-50 border border-slate-200 rounded-2xl text-left hover:bg-white hover:border-blue-300 transition-all shadow-sm"
              >
                <div className="flex items-center gap-3">
                  <span className="font-bold text-slate-700">{selectedVoice.name}</span>
                  <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold uppercase ${selectedVoice.color}`}>
                    {selectedVoice.gender}
                  </span>
                </div>
                <X size={18} className={`text-slate-400 transition-transform ${isDropdownOpen ? 'rotate-180' : ''}`} />
              </button>

              {isDropdownOpen && (
                <div className="absolute top-full left-0 right-0 mt-2 bg-white border border-slate-100 rounded-2xl shadow-xl z-10 overflow-hidden animate-in slide-in-from-top-2 duration-200">
                  <div className="p-3 border-b border-slate-50">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={16} />
                      <input 
                        type="text" 
                        placeholder="Tìm kiếm giọng nói..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-100 rounded-xl text-sm outline-none focus:border-blue-200"
                        onClick={(e) => e.stopPropagation()}
                      />
                    </div>
                  </div>
                  <div className="max-h-60 overflow-y-auto">
                    {filteredVoices.map(v => (
                      <button
                        key={v.id}
                        onClick={() => {
                          setLocalSettings({ ...localSettings, voiceId: v.id });
                          setIsDropdownOpen(false);
                        }}
                        className={`w-full flex items-center justify-between px-4 py-3 hover:bg-slate-50 transition-colors ${localSettings.voiceId === v.id ? 'bg-blue-50/50' : ''}`}
                      >
                        <div className="flex items-center gap-3">
                          <span className={`font-bold ${localSettings.voiceId === v.id ? 'text-blue-600' : 'text-slate-600'}`}>{v.name}</span>
                          <span className={`text-[9px] px-1.5 py-0.5 rounded-full font-bold uppercase ${v.color}`}>
                            {v.gender}
                          </span>
                        </div>
                        <Play size={14} className="text-slate-400 hover:text-blue-600" />
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Custom Voice ID for VClip */}
          {localSettings.voiceId === 'custom' && (
            <div className="space-y-3 p-4 bg-purple-50 border border-purple-100 rounded-2xl animate-in slide-in-from-top-2">
              <label className="text-sm font-bold text-purple-700">Mã giọng nói VClip (UserVoiceId)</label>
              <input 
                type="text"
                placeholder="Nhập mã giọng nói từ vclip.io..."
                value={localSettings.customVoiceId || ''}
                onChange={(e) => setLocalSettings({ ...localSettings, customVoiceId: e.target.value })}
                className="w-full px-4 py-3 bg-white border border-purple-200 rounded-xl text-sm outline-none focus:border-purple-400"
              />
              <p className="text-[10px] text-purple-400 font-medium italic">
                * Mã này bạn lấy từ phần quản lý giọng nói trên website VClip.io
              </p>
            </div>
          )}
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <label className="text-sm font-bold text-slate-700">Tốc độ nói: {localSettings.speed.toFixed(2)}x</label>
            </div>
            <div className="px-2">
              <input 
                type="range" min="0.5" max="3" step="0.1"
                value={localSettings.speed}
                onChange={(e) => setLocalSettings({ ...localSettings, speed: parseFloat(e.target.value) })}
                className="w-full h-2 bg-slate-100 rounded-full appearance-none accent-blue-600 cursor-pointer"
              />
              <div className="flex justify-between mt-2 text-[10px] text-slate-400 font-bold uppercase tracking-wider">
                <span>Chậm (0.5x)</span>
                <span>Nhanh (3.0x)</span>
              </div>
            </div>
          </div>

          {/* Emotion Slider */}
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <label className="text-sm font-bold text-slate-700">Cảm xúc: {localSettings.emotion.toFixed(2)}</label>
            </div>
            <div className="px-2">
              <input 
                type="range" min="0.1" max="1" step="0.05"
                value={localSettings.emotion}
                onChange={(e) => setLocalSettings({ ...localSettings, emotion: parseFloat(e.target.value) })}
                className="w-full h-2 bg-slate-100 rounded-full appearance-none accent-blue-600 cursor-pointer"
              />
              <div className="flex justify-between mt-2 text-[10px] text-slate-400 font-bold uppercase tracking-wider">
                <span>Đều đặn (0.1)</span>
                <span>Biểu cảm (1.0)</span>
              </div>
            </div>
          </div>

          {/* Summary Card */}
          <div className="p-5 bg-slate-50 border border-slate-100 rounded-2xl space-y-1.5">
            <h5 className="text-[11px] font-bold text-slate-400 uppercase tracking-widest mb-1">Cấu hình hiện tại:</h5>
            <p className="text-xs font-bold text-slate-600">Giọng: <span className="text-slate-800">{selectedVoice.name}</span></p>
            <p className="text-xs font-bold text-slate-600">Tốc độ: <span className="text-slate-800">{localSettings.speed.toFixed(2)}x</span></p>
            <p className="text-xs font-bold text-slate-600">Cảm xúc: <span className="text-slate-800">{localSettings.emotion.toFixed(2)}</span></p>
          </div>
        </div>

        {/* Footer */}
        <div className="p-6 border-t border-slate-50 bg-slate-50/30 flex gap-3">
          <button 
            onClick={onClose}
            className="flex-1 py-4 text-sm font-bold text-slate-500 bg-white border border-slate-200 rounded-2xl hover:bg-slate-50 transition-all active:scale-95"
          >
            Hủy
          </button>
          <button 
            onClick={handleSave}
            className="flex-1 py-4 text-sm font-bold text-white bg-blue-600 rounded-2xl hover:bg-blue-700 shadow-lg shadow-blue-200 transition-all active:scale-95 flex items-center justify-center gap-2"
          >
            <Check size={18} /> Lưu thay đổi
          </button>
        </div>
      </div>
    </div>
  );
}
