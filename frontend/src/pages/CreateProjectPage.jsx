import React, { useState } from 'react';
import { Upload, Loader2, CheckCircle, ArrowLeft, Play } from 'lucide-react';

const LANGUAGES = [
  { value: 'zh_cn', label: 'Tiếng Trung (Giản thể)' }, { value: 'zh_tw', label: 'Tiếng Trung (Phồn thể)' },
  { value: 'en', label: 'Tiếng Anh' }, { value: 'vi', label: 'Tiếng Việt' },
  { value: 'ja', label: 'Tiếng Nhật' }, { value: 'ko', label: 'Tiếng Hàn' },
  { value: 'fr', label: 'Tiếng Pháp' }, { value: 'de', label: 'Tiếng Đức' },
  { value: 'es', label: 'Tiếng Tây Ban Nha' }, { value: 'pt', label: 'Tiếng Bồ Đào Nha' },
  { value: 'ru', label: 'Tiếng Nga' }, { value: 'it', label: 'Tiếng Ý' },
  { value: 'th', label: 'Tiếng Thái' }, { value: 'id', label: 'Tiếng Indonesia' },
  { value: 'ms', label: 'Tiếng Mã Lai' }, { value: 'ar', label: 'Tiếng Ả Rập' },
  { value: 'hi', label: 'Tiếng Hindi' }, { value: 'bn', label: 'Tiếng Bengal' },
  { value: 'fil', label: 'Tiếng Philippines' }, { value: 'my', label: 'Tiếng Miến Điện' },
  { value: 'fa', label: 'Tiếng Ba Tư' }, { value: 'tr', label: 'Tiếng Thổ Nhĩ Kỳ' },
  { value: 'pl', label: 'Tiếng Ba Lan' }, { value: 'nl', label: 'Tiếng Hà Lan' },
  { value: 'uk', label: 'Tiếng Ukraine' }, { value: 'ro', label: 'Tiếng Romania' },
  { value: 'sv', label: 'Tiếng Thụy Điển' }, { value: 'da', label: 'Tiếng Đan Mạch' },
  { value: 'no', label: 'Tiếng Na Uy' }, { value: 'fi', label: 'Tiếng Phần Lan' },
];

export default function CreateProjectPage({ onNavigate }) {
  const [step, setStep] = useState(1);
  const [videoSource, setVideoSource] = useState('url'); // 'url' or 'file'
  const [videoUrl, setVideoUrl] = useState('');
  const [uploadedVideoUrl, setUploadedVideoUrl] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState('');

  const [sourceLang, setSourceLang] = useState('zh_cn');
  const [targetLang, setTargetLang] = useState('vi');
  const [bilingual, setBilingual] = useState(true);
  const [bilingualPos, setBilingualPos] = useState('bottom');
  const [fillerFilter, setFillerFilter] = useState(true);

  const [enableTTS, setEnableTTS] = useState(false);
  const [ttsVoiceCode, setTtsVoiceCode] = useState('');

  const [enableComposite, setEnableComposite] = useState(false);
  const [compositeType, setCompositeType] = useState('horizontal');

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const [availableVoices, setAvailableVoices] = useState([]);

  React.useEffect(() => {
    fetch('/api/config')
      .then(res => res.json())
      .then(json => {
        if (json.error === 0 && json.data?.tts?.voices) {
          const voices = json.data.tts.voices;
          setAvailableVoices(voices);
          // Auto-set a default voice from gallery if we have one
          if (voices.length > 0) {
            setTtsVoiceCode(voices[0].id);
          }
        }
      });
  }, []);

  // Upload file
  const handleFileUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    setUploading(true);
    setUploadProgress(`Đang tải lên: ${file.name}`);
    setError(null);

    const formData = new FormData();
    formData.append('file', file);

    try {
      const res = await fetch('/api/file', { method: 'POST', body: formData });
      const json = await res.json();
      if (json.error !== 0) throw new Error(json.msg || 'Upload failed');
      setUploadedVideoUrl(json.data.file_path);
      setUploadProgress(`✅ Đã tải lên: ${file.name}`);
    } catch (err) {
      setError('Upload thất bại: ' + err.message);
      setUploadProgress('');
    } finally {
      setUploading(false);
    }
  };

  // Submit task
  const handleSubmit = async () => {
    const videoPath = videoSource === 'url' ? videoUrl : uploadedVideoUrl;
    if (!videoPath) {
      setError('Vui lòng nhập URL hoặc upload video trước.');
      return;
    }

    if (enableTTS && !ttsVoiceCode) {
      setError('Vui lòng chọn hoặc nhập mã giọng nói khi kích hoạt lồng tiếng.');
      setStep(3);
      return;
    }

    setSubmitting(true);
    setError(null);

    const params = {
      url: String(videoPath),
      language: 'zh_cn',
      origin_lang: sourceLang,
      target_lang: targetLang,
      bilingual: bilingual ? 1 : 2,
      translation_subtitle_pos: bilingualPos === 'top' ? 1 : 2,
      tts: enableTTS ? 1 : 2,
      modal_filter: fillerFilter ? 1 : 2,
      embed_subtitle_video_type: enableComposite ? compositeType : 'none',
      enable_review: true,
    };

    if (enableTTS && ttsVoiceCode) {
      params.tts_voice_code = ttsVoiceCode;
    }

    try {
      const res = await fetch('/api/capability/subtitleTask', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(params),
      });
      const json = await res.json();
      if (json.error !== 0) throw new Error(json.msg || 'Gửi yêu cầu thất bại');

      const taskId = json.data?.task_id;
      if (!taskId) throw new Error('Không nhận được Task ID từ server');

      // Save to localStorage
      const projectName = videoSource === 'url'
        ? videoUrl
        : (typeof uploadedVideoUrl === 'string' ? uploadedVideoUrl.split('/').pop() : 'Uploaded Video');

      const projects = JSON.parse(localStorage.getItem('mk_projects') || '[]');
      projects.unshift({
        taskId,
        name: projectName,
        status: 'processing',
        percent: 0,
        createdAt: new Date().toISOString(),
      });
      localStorage.setItem('mk_projects', JSON.stringify(projects));

      onNavigate('dashboard');
    } catch (err) {
      setError(err.message || String(err));
    } finally {
      setSubmitting(false);
    }
  };

  const canProceed = (step === 1 && (videoUrl || uploadedVideoUrl))
    || step === 2 || step === 3;

  return (
    <div className="flex-1 overflow-y-auto bg-slate-50">
      <div className="max-w-2xl mx-auto py-8 px-6">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          <button onClick={() => onNavigate('dashboard')} className="p-2 hover:bg-slate-200 rounded-lg transition-colors cursor-pointer">
            <ArrowLeft size={20} className="text-slate-500" />
          </button>
          <div>
            <h1 className="text-xl font-bold text-slate-800">Tạo dự án mới</h1>
            <p className="text-sm text-slate-500 mt-0.5">Thiết lập video dịch và lồng tiếng</p>
          </div>
        </div>

        {/* Stepper */}
        <div className="flex items-center gap-3 mb-8">
          {[1, 2, 3].map((s) => (
            <React.Fragment key={s}>
              <button
                onClick={() => setStep(s)}
                className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold transition-colors cursor-pointer ${
                  step === s
                    ? 'bg-blue-600 text-white shadow-sm'
                    : step > s
                    ? 'bg-green-100 text-green-700'
                    : 'bg-slate-100 text-slate-400'
                }`}
              >
                {step > s ? <CheckCircle size={16} /> : s}
              </button>
              {s < 3 && <div className={`flex-1 h-0.5 rounded ${step > s ? 'bg-green-300' : 'bg-slate-200'}`} />}
            </React.Fragment>
          ))}
        </div>

        {/* Step 1: Video source */}
        {step === 1 && (
          <div className="bg-white border border-slate-200 rounded-xl p-6 shadow-sm">
            <h2 className="font-semibold text-base text-slate-800 mb-1">Bước 1: Nguồn video</h2>
            <p className="text-sm text-slate-500 mb-6">Nhập URL video hoặc tải lên từ máy tính</p>

            {/* Toggle URL / File */}
            <div className="flex gap-2 mb-5">
              <button
                onClick={() => setVideoSource('url')}
                className={`px-4 py-2 text-sm font-medium rounded-lg cursor-pointer transition-colors ${videoSource === 'url' ? 'bg-blue-600 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'}`}
              >
                Nhập URL
              </button>
              <button
                onClick={() => setVideoSource('file')}
                className={`px-4 py-2 text-sm font-medium rounded-lg cursor-pointer transition-colors ${videoSource === 'file' ? 'bg-blue-600 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'}`}
              >
                Upload file
              </button>
            </div>

            {videoSource === 'url' ? (
              <input
                type="url"
                value={videoUrl}
                onChange={(e) => setVideoUrl(e.target.value)}
                placeholder="https://www.youtube.com/watch?v=... hoặc TikTok URL"
                className="w-full px-4 py-3 border-2 border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 transition-colors"
              />
            ) : (
              <div>
                <label className="flex flex-col items-center justify-center w-full py-10 border-2 border-dashed border-slate-300 rounded-lg cursor-pointer hover:border-blue-400 hover:bg-blue-50/50 transition-colors">
                  {uploading ? (
                    <Loader2 size={32} className="animate-spin text-blue-500 mb-2" />
                  ) : (
                    <Upload size={32} className="text-slate-300 mb-2" />
                  )}
                  <span className="text-sm text-slate-500">
                    {uploadProgress || 'Kéo thả hoặc click để chọn file video'}
                  </span>
                  <input type="file" accept="video/*" onChange={handleFileUpload} className="hidden" disabled={uploading} />
                </label>
              </div>
            )}
          </div>
        )}

        {/* Step 2: Language & Subtitles */}
        {step === 2 && (
          <div className="bg-white border border-slate-200 rounded-xl p-6 shadow-sm space-y-5">
            <div>
              <h2 className="font-semibold text-base text-slate-800 mb-1">Bước 2: Ngôn ngữ & Phụ đề</h2>
              <p className="text-sm text-slate-500 mb-6">Chọn ngôn ngữ nguồn và ngôn ngữ đích dịch</p>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1.5">Ngôn ngữ nguồn</label>
                <select value={sourceLang} onChange={(e) => setSourceLang(e.target.value)}
                  className="w-full px-3 py-2.5 border-2 border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 cursor-pointer">
                  {LANGUAGES.map((l) => <option key={l.value} value={l.value}>{l.label}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1.5">Ngôn ngữ đích</label>
                <select value={targetLang} onChange={(e) => setTargetLang(e.target.value)}
                  className="w-full px-3 py-2.5 border-2 border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 cursor-pointer">
                  {LANGUAGES.map((l) => <option key={l.value} value={l.value}>{l.label}</option>)}
                </select>
              </div>
            </div>

            <div className="flex items-center justify-between py-3 border-t border-slate-100">
              <div>
                <p className="text-sm font-medium text-slate-700">Phụ đề song ngữ</p>
                <p className="text-xs text-slate-400">Hiển thị cả ngôn ngữ gốc và bản dịch</p>
              </div>
              <ToggleSwitch checked={bilingual} onChange={setBilingual} />
            </div>

            {bilingual && (
              <div className="pl-4 border-l-2 border-blue-200">
                <label className="block text-sm font-medium text-slate-600 mb-1.5">Vị trí bản dịch</label>
                <select value={bilingualPos} onChange={(e) => setBilingualPos(e.target.value)}
                  className="px-3 py-2 border border-slate-200 rounded-lg text-sm cursor-pointer">
                  <option value="top">Bản dịch ở trên</option>
                  <option value="bottom">Bản dịch ở dưới</option>
                </select>
              </div>
            )}

            <div className="flex items-center justify-between py-3 border-t border-slate-100">
              <div>
                <p className="text-sm font-medium text-slate-700">Lọc từ đệm</p>
                <p className="text-xs text-slate-400">Tự động lọc bỏ các từ thừa (ờ, ừm...)</p>
              </div>
              <ToggleSwitch checked={fillerFilter} onChange={setFillerFilter} />
            </div>
          </div>
        )}

        {/* Step 3: TTS & Composite */}
        {step === 3 && (
          <div className="bg-white border border-slate-200 rounded-xl p-6 shadow-sm space-y-5">
            <div>
              <h2 className="font-semibold text-base text-slate-800 mb-1">Bước 3: Lồng tiếng & Tổng hợp</h2>
              <p className="text-sm text-slate-500 mb-6">Cài đặt giọng nói AI và định dạng video xuất</p>
            </div>

            <div className="flex items-center justify-between py-3">
              <div>
                <p className="text-sm font-medium text-slate-700">Kích hoạt lồng tiếng</p>
                <p className="text-xs text-slate-400">Sử dụng AI để lồng tiếng bản dịch</p>
              </div>
              <ToggleSwitch checked={enableTTS} onChange={setEnableTTS} />
            </div>

            {enableTTS && (
              <div className="pl-4 border-l-2 border-blue-200">
                <label className="block text-sm font-medium text-slate-600 mb-1.5">Mã giọng nói (Voice Code)</label>
                <div className="flex gap-2">
                  <select 
                    value={ttsVoiceCode} 
                    onChange={(e) => setTtsVoiceCode(e.target.value)}
                    className="flex-1 px-3 py-2.5 border-2 border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 cursor-pointer bg-white"
                  >
                    {availableVoices.map(v => (
                      <option key={v.id} value={v.id}>{v.name} ({v.id})</option>
                    ))}
                    <option value="custom">-- Nhập mã tùy chỉnh --</option>
                  </select>
                </div>
                {ttsVoiceCode === 'custom' && (
                  <input
                    type="text"
                    onChange={(e) => {
                      // Note: this is a bit tricky since value is shared
                      // but for simplicity we'll just allow typing if custom is selected
                    }}
                    onBlur={(e) => setTtsVoiceCode(e.target.value)}
                    placeholder="Nhập UserVoiceId từ vclip.io..."
                    className="mt-2 w-full px-3 py-2.5 border-2 border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400"
                  />
                )}
              </div>
            )}

            <div className="flex items-center justify-between py-3 border-t border-slate-100">
              <div>
                <p className="text-sm font-medium text-slate-700">Tổng hợp video</p>
                <p className="text-xs text-slate-400">Xuất video kèm phụ đề nhúng sẵn</p>
              </div>
              <ToggleSwitch checked={enableComposite} onChange={setEnableComposite} />
            </div>

            {enableComposite && (
              <div className="pl-4 border-l-2 border-blue-200">
                <label className="block text-sm font-medium text-slate-600 mb-1.5">Định dạng xuất</label>
                <select value={compositeType} onChange={(e) => setCompositeType(e.target.value)}
                  className="px-3 py-2 border border-slate-200 rounded-lg text-sm cursor-pointer">
                  <option value="horizontal">Ngang (16:9)</option>
                  <option value="vertical">Dọc (9:16)</option>
                  <option value="all">Ngang + Dọc</option>
                </select>
              </div>
            )}
          </div>
        )}

        {/* Error */}
        {error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg">
            {error}
          </div>
        )}

        {/* Navigation buttons */}
        <div className="flex items-center justify-between mt-6">
          <button
            onClick={() => step > 1 ? setStep(step - 1) : onNavigate('dashboard')}
            className="px-5 py-2.5 text-sm font-medium text-slate-600 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 cursor-pointer transition-colors"
          >
            {step > 1 ? 'Quay lại' : 'Hủy'}
          </button>

          {step < 3 ? (
            <button
              onClick={() => setStep(step + 1)}
              className="px-5 py-2.5 text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg cursor-pointer transition-colors shadow-sm"
            >
              Tiếp tục
            </button>
          ) : (
            <button
              onClick={handleSubmit}
              disabled={submitting}
              className="flex items-center gap-2 px-6 py-2.5 text-sm font-semibold text-white bg-green-600 hover:bg-green-700 disabled:bg-green-400 rounded-lg cursor-pointer transition-colors shadow-sm"
            >
              {submitting ? <Loader2 size={16} className="animate-spin" /> : <Play size={16} />}
              {submitting ? 'Đang gửi...' : 'Bắt đầu dịch'}
            </button>
          )}
        </div>
      </div>
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
