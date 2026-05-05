import React, { useState, useEffect } from 'react';
import { Save, Loader2, CheckCircle, RefreshCw, Trash2 } from 'lucide-react';

export default function SettingsPage() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState(null);

  // === App Config ===
  const [segmentDuration, setSegmentDuration] = useState(5);
  const [transcribeParallelNum, setTranscribeParallelNum] = useState(1);
  const [translateParallelNum, setTranslateParallelNum] = useState(3);
  const [transcribeMaxAttempts, setTranscribeMaxAttempts] = useState(3);
  const [translateMaxAttempts, setTranslateMaxAttempts] = useState(5);
  const [maxSentenceLength, setMaxSentenceLength] = useState(70);
  const [proxy, setProxy] = useState('');

  // === Server Config ===
  const [serverHost, setServerHost] = useState('127.0.0.1');
  const [serverPort, setServerPort] = useState(8888);

  // === LLM Config ===
  const [llmBaseUrl, setLlmBaseUrl] = useState('');
  const [llmApiKey, setLlmApiKey] = useState('');
  const [llmModel, setLlmModel] = useState('gpt-4o-mini');

  // === Transcription Config ===
  const [transcribeProvider, setTranscribeProvider] = useState('openai');
  const [enableGpu, setEnableGpu] = useState(false);
  const [transcribeOpenaiApiKey, setTranscribeOpenaiApiKey] = useState('');
  const [transcribeOpenaiBaseUrl, setTranscribeOpenaiBaseUrl] = useState('');
  const [transcribeOpenaiModel, setTranscribeOpenaiModel] = useState('whisper-1');
  const [fasterwhisperModel, setFasterwhisperModel] = useState('large-v2');
  const [whisperkitModel, setWhisperkitModel] = useState('large-v2');
  const [whispercppModel, setWhispercppModel] = useState('large-v2');

  // === TTS Config ===
  const [ttsProvider, setTtsProvider] = useState('openai');
  const [ttsOpenaiBaseUrl, setTtsOpenaiBaseUrl] = useState('');
  const [ttsOpenaiApiKey, setTtsOpenaiApiKey] = useState('');
  const [ttsOpenaiModel, setTtsOpenaiModel] = useState('tts-1');
  const [ttsOpenaiVoice, setTtsOpenaiVoice] = useState('alloy');
  const [vclipApiKey, setVclipApiKey] = useState('');
  const [vclipVoiceId, setVclipVoiceId] = useState('');
  const [vclipSpeed, setVclipSpeed] = useState(1.0);
  const [voices, setVoices] = useState([]); // [{name, id}]
  const [newVoiceName, setNewVoiceName] = useState('');
  const [newVoiceId, setNewVoiceId] = useState('');

  // Raw config ref for merging
  const [rawConfig, setRawConfig] = useState(null);

  useEffect(() => { loadConfig(); }, []);

  const loadConfig = async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/config');
      const json = await res.json();
      if (json.error !== 0) throw new Error(json.msg);
      const c = json.data;
      setRawConfig(c);

      // App
      setSegmentDuration(c?.app?.segment_duration ?? 5);
      setTranscribeParallelNum(c?.app?.transcribe_parallel_num ?? 1);
      setTranslateParallelNum(c?.app?.translate_parallel_num ?? 3);
      setTranscribeMaxAttempts(c?.app?.transcribe_max_attempts ?? 3);
      setTranslateMaxAttempts(c?.app?.translate_max_attempts ?? 5);
      setMaxSentenceLength(c?.app?.max_sentence_length ?? 70);
      setProxy(c?.app?.proxy ?? '');

      // Server
      setServerHost(c?.server?.host ?? '127.0.0.1');
      setServerPort(c?.server?.port ?? 8888);

      // LLM
      setLlmBaseUrl(c?.llm?.base_url ?? '');
      setLlmApiKey(c?.llm?.api_key ?? '');
      setLlmModel(c?.llm?.model ?? 'gpt-4o-mini');

      // Transcription
      setTranscribeProvider(c?.transcribe?.provider ?? 'openai');
      setEnableGpu(c?.transcribe?.enable_gpu_acceleration ?? false);
      setTranscribeOpenaiBaseUrl(c?.transcribe?.openai?.base_url ?? '');
      setTranscribeOpenaiApiKey(c?.transcribe?.openai?.api_key ?? '');
      setTranscribeOpenaiModel(c?.transcribe?.openai?.model ?? 'whisper-1');
      setFasterwhisperModel(c?.transcribe?.fasterwhisper?.model ?? 'large-v2');
      setWhisperkitModel(c?.transcribe?.whisperkit?.model ?? 'large-v2');
      setWhispercppModel(c?.transcribe?.whispercpp?.model ?? 'large-v2');

      // TTS
      setTtsProvider(c?.tts?.provider === 'edge-tts' ? 'openai' : (c?.tts?.provider ?? 'openai'));
      setTtsOpenaiBaseUrl(c?.tts?.openai?.base_url ?? '');
      setTtsOpenaiApiKey(c?.tts?.openai?.api_key ?? '');
      setTtsOpenaiModel(c?.tts?.openai?.model ?? 'tts-1');
      setTtsOpenaiVoice(c?.tts?.openai?.voice ?? 'alloy');
      setVclipApiKey(c?.tts?.vclip?.api_key ?? '');
      setVclipVoiceId(c?.tts?.vclip?.voice_id ?? '');
      setVclipSpeed(c?.tts?.vclip?.speed ?? 1.0);
      setVoices(c?.tts?.voices ?? []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setSaved(false);
    setError(null);

    const payload = {
      app: {
        segment_duration: parseInt(segmentDuration),
        transcribe_parallel_num: parseInt(transcribeParallelNum),
        translate_parallel_num: parseInt(translateParallelNum),
        transcribe_max_attempts: parseInt(transcribeMaxAttempts),
        translate_max_attempts: parseInt(translateMaxAttempts),
        max_sentence_length: parseInt(maxSentenceLength),
        proxy: proxy,
      },
      server: {
        host: serverHost,
        port: parseInt(serverPort),
      },
      llm: {
        base_url: llmBaseUrl,
        api_key: llmApiKey,
        model: llmModel,
      },
      transcribe: {
        provider: transcribeProvider,
        enable_gpu_acceleration: enableGpu,
        openai: { base_url: transcribeOpenaiBaseUrl, api_key: transcribeOpenaiApiKey, model: transcribeOpenaiModel },
        fasterwhisper: { model: fasterwhisperModel },
        whisperkit: { model: whisperkitModel },
        whispercpp: { model: whispercppModel },
      },
      tts: {
        provider: ttsProvider,
        openai: { base_url: ttsOpenaiBaseUrl, api_key: ttsOpenaiApiKey, model: ttsOpenaiModel, voice: ttsOpenaiVoice },
        vclip: { api_key: vclipApiKey, voice_id: vclipVoiceId, speed: parseFloat(vclipSpeed) },
        voices: voices,
      },
    };

    try {
      const res = await fetch('/api/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const json = await res.json();
      if (json.error !== 0) throw new Error(json.msg || 'Save failed');
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <Loader2 size={32} className="animate-spin text-blue-500" />
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto bg-slate-50">
      <div className="max-w-3xl mx-auto py-8 px-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-xl font-bold text-slate-800">Cài đặt</h1>
            <p className="text-sm text-slate-500 mt-0.5">Cấu hình hệ thống, LLM, Whisper và TTS</p>
          </div>
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white text-sm font-semibold rounded-lg transition-colors cursor-pointer shadow-sm"
          >
            {saving ? <Loader2 size={16} className="animate-spin" /> : saved ? <CheckCircle size={16} /> : <Save size={16} />}
            {saving ? 'Đang lưu...' : saved ? 'Đã lưu!' : 'Lưu cấu hình'}
          </button>
        </div>

        {error && (
          <div className="mb-6 p-3 bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg">{error}</div>
        )}

        {/* ========== LLM Config ========== */}
        <Section title="Cấu hình LLM" subtitle="Kết nối với dịch vụ AI để dịch thuật">
          <div className="grid grid-cols-1 gap-4">
            <Field label="API Base URL" value={llmBaseUrl} onChange={setLlmBaseUrl} placeholder="https://api.openai.com/v1" />
            <Field label="API Key" value={llmApiKey} onChange={setLlmApiKey} placeholder="sk-..." type="password" />
            <Field label="Model Name" value={llmModel} onChange={setLlmModel} placeholder="gpt-4o-mini" />
          </div>
        </Section>

        {/* ========== Transcription Config ========== */}
        <Section title="Cấu hình Transcription (Whisper)" subtitle="Nhận dạng giọng nói từ video">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1.5">Provider</label>
              <select value={transcribeProvider} onChange={(e) => setTranscribeProvider(e.target.value)}
                className="w-full px-3 py-2.5 border-2 border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 cursor-pointer bg-white">
                <option value="openai">OpenAI Whisper API</option>
                <option value="fasterwhisper">FasterWhisper (Local)</option>
                <option value="whispercpp">WhisperCpp (Windows)</option>
                <option value="whisperkit">WhisperKit (macOS)</option>
              </select>
            </div>

            <div className="flex items-center justify-between py-2">
              <div>
                <p className="text-sm font-medium text-slate-700">GPU Acceleration</p>
                <p className="text-xs text-slate-400">Tăng tốc xử lý bằng GPU (nếu có)</p>
              </div>
              <ToggleSwitch checked={enableGpu} onChange={setEnableGpu} />
            </div>

            {transcribeProvider === 'openai' && (
              <div className="pl-4 border-l-2 border-blue-200 space-y-3">
                <Field label="Whisper API Base URL" value={transcribeOpenaiBaseUrl} onChange={setTranscribeOpenaiBaseUrl} placeholder="https://api.openai.com/v1" />
                <Field label="Whisper API Key" value={transcribeOpenaiApiKey} onChange={setTranscribeOpenaiApiKey} placeholder="sk-..." type="password" />
                <Field label="Whisper Model" value={transcribeOpenaiModel} onChange={setTranscribeOpenaiModel} placeholder="whisper-1" />
              </div>
            )}
            {transcribeProvider === 'fasterwhisper' && (
              <div className="pl-4 border-l-2 border-blue-200">
                <SelectField label="FasterWhisper Model" value={fasterwhisperModel} onChange={setFasterwhisperModel}
                  options={[{ v: 'tiny', l: 'tiny' }, { v: 'medium', l: 'medium' }, { v: 'large-v2', l: 'large-v2' }, { v: 'large-v3', l: 'large-v3' }]} />
              </div>
            )}
            {transcribeProvider === 'whispercpp' && (
              <div className="pl-4 border-l-2 border-blue-200">
                <SelectField label="WhisperCpp Model" value={whispercppModel} onChange={setWhispercppModel}
                  options={[{ v: 'large-v2', l: 'large-v2' }, { v: 'large-v3', l: 'large-v3' }]} />
              </div>
            )}
            {transcribeProvider === 'whisperkit' && (
              <div className="pl-4 border-l-2 border-blue-200">
                <SelectField label="WhisperKit Model" value={whisperkitModel} onChange={setWhisperkitModel}
                  options={[{ v: 'large-v2', l: 'large-v2' }, { v: 'large-v3', l: 'large-v3' }]} />
              </div>
            )}
          </div>
        </Section>

        {/* ========== TTS Config ========== */}
        <Section title="Cấu hình TTS (Text-to-Speech)" subtitle="Dịch vụ tổng hợp giọng nói cho lồng tiếng">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1.5">TTS Provider</label>
              <select value={ttsProvider} onChange={(e) => setTtsProvider(e.target.value)}
                className="w-full px-3 py-2.5 border-2 border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 cursor-pointer bg-white">
                <option value="openai">OpenAI TTS</option>
                <option value="vclip">VClip TTS</option>
              </select>
            </div>


            {ttsProvider === 'openai' && (
              <div className="pl-4 border-l-2 border-green-200 space-y-3">
                <Field label="TTS API Base URL" value={ttsOpenaiBaseUrl} onChange={setTtsOpenaiBaseUrl} placeholder="https://api.openai.com/v1" />
                <Field label="TTS API Key" value={ttsOpenaiApiKey} onChange={setTtsOpenaiApiKey} placeholder="sk-..." type="password" />
                <Field label="TTS Model" value={ttsOpenaiModel} onChange={setTtsOpenaiModel} placeholder="tts-1" />
                <Field label="TTS Voice" value={ttsOpenaiVoice} onChange={setTtsOpenaiVoice} placeholder="alloy" />
              </div>
            )}
            {ttsProvider === 'vclip' && (
              <div className="pl-4 border-l-2 border-green-200 space-y-3">
                <Field label="VClip API Key" value={vclipApiKey} onChange={setVclipApiKey} placeholder="vclip-api-key" type="password" />
                <Field label="Mã giọng nói mặc định (UserVoiceId)" value={vclipVoiceId} onChange={setVclipVoiceId} placeholder="Nhập mã từ vclip.io" />
                <Field label="Speed" value={vclipSpeed} onChange={setVclipSpeed} placeholder="1.0" type="number" />
              </div>
            )}
          </div>
        </Section>
        {/* ========== Voice Gallery ========== */}
        <Section title="Thư viện giọng nói" subtitle="Lưu các ID giọng nói thường dùng để chọn nhanh">
          <div className="space-y-4">
            {/* List of voices */}
            <div className="border border-slate-100 rounded-lg overflow-hidden">
              <table className="w-full text-sm text-left">
                <thead className="bg-slate-50 text-slate-500 uppercase text-[10px] font-bold">
                  <tr>
                    <th className="px-4 py-3">Tên giọng</th>
                    <th className="px-4 py-3">ID (Voice ID / UserVoiceId)</th>
                    <th className="px-4 py-3 w-10"></th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {voices.map((v, idx) => (
                    <tr key={idx} className="hover:bg-slate-50 transition-colors">
                      <td className="px-4 py-3 font-medium text-slate-700">{v.name}</td>
                      <td className="px-4 py-3 text-slate-500 font-mono text-xs">{v.id}</td>
                      <td className="px-4 py-3 text-center">
                        <button 
                          onClick={() => setVoices(voices.filter((_, i) => i !== idx))}
                          className="text-slate-300 hover:text-red-500 transition-colors cursor-pointer"
                        >
                          <Trash2 size={14} />
                        </button>
                      </td>
                    </tr>
                  ))}
                  {voices.length === 0 && (
                    <tr>
                      <td colSpan="3" className="px-4 py-8 text-center text-slate-400 italic">Chưa có giọng nói nào được lưu</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Add new voice form */}
            <div className="p-4 bg-blue-50/50 border border-blue-100 rounded-xl flex gap-3 items-end">
              <div className="flex-1">
                <label className="block text-[10px] font-bold text-blue-600 uppercase mb-1">Tên hiển thị</label>
                <input 
                  type="text" 
                  value={newVoiceName}
                  onChange={(e) => setNewVoiceName(e.target.value)}
                  placeholder="Ví dụ: Quang Anh (VClip)"
                  className="w-full px-3 py-2 bg-white border border-blue-200 rounded-lg text-sm outline-none focus:border-blue-400"
                />
              </div>
              <div className="flex-1">
                <label className="block text-[10px] font-bold text-blue-600 uppercase mb-1">ID (Mã giọng)</label>
                <input 
                  type="text" 
                  value={newVoiceId}
                  onChange={(e) => setNewVoiceId(e.target.value)}
                  placeholder="ID từ OpenAI hoặc VClip"
                  className="w-full px-3 py-2 bg-white border border-blue-200 rounded-lg text-sm outline-none focus:border-blue-400"
                />
              </div>
              <button 
                onClick={() => {
                  if (newVoiceName && newVoiceId) {
                    setVoices([...voices, { name: newVoiceName, id: newVoiceId }]);
                    setNewVoiceName('');
                    setNewVoiceId('');
                  }
                }}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-bold rounded-lg transition-colors cursor-pointer"
              >
                Thêm
              </button>
            </div>
          </div>
        </Section>

        {/* ========== App Config ========== */}
        <Section title="Cấu hình ứng dụng" subtitle="Các thông số xử lý video">
          <div className="grid grid-cols-2 gap-4">
            <Field label="Thời lượng phân đoạn (phút)" value={segmentDuration} onChange={setSegmentDuration} type="number" />
            <Field label="Số luồng chuyển soạn song song" value={transcribeParallelNum} onChange={setTranscribeParallelNum} type="number" />
            <Field label="Số luồng dịch song song" value={translateParallelNum} onChange={setTranslateParallelNum} type="number" />
            <Field label="Số lần thử chuyển soạn tối đa" value={transcribeMaxAttempts} onChange={setTranscribeMaxAttempts} type="number" />
            <Field label="Số lần thử dịch tối đa" value={translateMaxAttempts} onChange={setTranslateMaxAttempts} type="number" />
            <Field label="Số ký tự tối đa mỗi câu" value={maxSentenceLength} onChange={setMaxSentenceLength} type="number" />
          </div>
          <div className="mt-4">
            <Field label="Proxy" value={proxy} onChange={setProxy} placeholder="http://127.0.0.1:7890" />
          </div>
        </Section>

        {/* ========== Server Config ========== */}
        <Section title="Cấu hình Server" subtitle="Địa chỉ và cổng máy chủ">
          <div className="grid grid-cols-2 gap-4">
            <Field label="Server Address" value={serverHost} onChange={setServerHost} placeholder="127.0.0.1" />
            <Field label="Server Port" value={serverPort} onChange={setServerPort} type="number" placeholder="8888" />
          </div>
        </Section>

        {/* Bottom save button */}
        <div className="flex justify-end mt-6 pb-8">
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white text-sm font-semibold rounded-lg transition-colors cursor-pointer shadow-sm"
          >
            {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
            {saving ? 'Đang lưu...' : 'Lưu tất cả cấu hình'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ========== Reusable Sub-Components ==========

function Section({ title, subtitle, children }) {
  return (
    <div className="bg-white border border-slate-200 rounded-xl p-6 shadow-sm mb-5">
      <h2 className="font-semibold text-base text-slate-800 mb-0.5">{title}</h2>
      {subtitle && <p className="text-xs text-slate-400 mb-5">{subtitle}</p>}
      {children}
    </div>
  );
}

function Field({ label, value, onChange, placeholder = '', type = 'text' }) {
  return (
    <div>
      <label className="block text-sm font-medium text-slate-700 mb-1.5">{label}</label>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full px-3 py-2.5 border-2 border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 transition-colors bg-white"
      />
    </div>
  );
}

function SelectField({ label, value, onChange, options }) {
  return (
    <div>
      <label className="block text-sm font-medium text-slate-700 mb-1.5">{label}</label>
      <select value={value} onChange={(e) => onChange(e.target.value)}
        className="w-full px-3 py-2.5 border-2 border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 cursor-pointer bg-white">
        {options.map((o) => <option key={o.v} value={o.v}>{o.l}</option>)}
      </select>
    </div>
  );
}

function ToggleSwitch({ checked, onChange }) {
  return (
    <button
      onClick={() => onChange(!checked)}
      className={`w-11 h-6 rounded-full relative transition-colors cursor-pointer ${checked ? 'bg-blue-600' : 'bg-slate-200'}`}
    >
      <div className={`w-5 h-5 bg-white rounded-full absolute top-0.5 transition-transform shadow-sm ${checked ? 'translate-x-[22px]' : 'translate-x-0.5'}`} />
    </button>
  );
}
