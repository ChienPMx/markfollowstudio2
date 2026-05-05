import React, { useState, useEffect } from 'react';
import VideoPlayer from '../components/VideoPlayer';
import RightPanel from '../components/RightPanel';
import BottomTimeline from '../components/BottomTimeline';
import VoiceConfigModal from '../components/VoiceConfigModal';
import { getSubtitleTask } from '../utils/api';
import { parseSrt } from '../utils/srtParser';
import { Loader2, X, AlertCircle } from 'lucide-react';

export default function EditorPage({ taskId, onNavigate }) {
  const [taskData, setTaskData] = useState(null);
  const [subtitles, setSubtitles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Comprehensive render settings matching the screenshots
  const [renderSettings, setRenderSettings] = useState({
    originalVolume: 0,
    subtitleStyle: 'karaoke-yellow',
    advanced: false,
    fontFamily: 'Helvetica',
    fontSize: 6.5, // % of height
    fontColor: '#FFFFFF',
    borderColor: '#000000',
    borderWidth: 9, // % of font size
    bgPadding: 40, // % of font size
    bottomDistance: 10, // % of height
    lineSpacing: 1.2,
    bgColor: '#00000000', // RRGGBBAA
    isBold: true,
    displayMode: 'full', // full, highlight, karaoke, word, fade
    highlightColor: '#FFFF00',
    maxWordsPerLine: 5,
    blurRegions: [],
    subtitleVisibility: 'both', // both, origin, translated
    subtitleOrder: 'translated-top', // translated-top, origin-top
    videoRatio: 'original', // 16:9, 9:16, 1:1, original, custom
    fitMode: 'fill' // cover, contain, fill, stretch
  });

  const [voiceSettings, setVoiceSettings] = useState({
    voiceId: 'alloy', // Better default
    speed: 1.0,
    emotion: 0.6
  });

  const [isVoiceModalOpen, setIsVoiceModalOpen] = useState(false);

  const updateRenderSetting = (key, val) => setRenderSettings(prev => ({ ...prev, [key]: val }));
  const updateVoiceSetting = (key, val) => setVoiceSettings(prev => ({ ...prev, [key]: val }));

  useEffect(() => {
    if (!taskId) return;

    const fetchTask = async () => {
      try {
        const data = await getSubtitleTask(taskId);
        console.log("Task data received:", data);
        setTaskData(data);
        
        // Handle subtitles
        if (data.review_srt_content) {
          const parsed = parseSrt(data.review_srt_content);
          setSubtitles(parsed);
          
          // Sync voiceId from task if it exists
          if (data.tts_voice_code) {
            setVoiceSettings(prev => ({ ...prev, voiceId: data.tts_voice_code }));
          }
        } else if (data.status === 'success' && data.subtitle_info?.length > 0) {
          console.log("Task success, fetching SRT file...");
          // If success, try to fetch the first SRT file for preview
          const srtFile = data.subtitle_info.find(f => f.name && f.name.toLowerCase().includes('.srt'));
          if (srtFile) {
            const srtRes = await fetch(srtFile.download_url);
            const srtText = await srtRes.text();
            setSubtitles(parseSrt(srtText) || []);
          }
        } else if (data.status === 'failed') {
          console.error("Task failed:", data.error_msg);
          setError(data.error_msg || "Task failed on server");
        }

        console.log("Setting loading to false");
        setLoading(false);
      } catch (err) {
        console.error("Fetch error:", err);
        setError(err.message || "Failed to fetch task data");
        setLoading(false);
      }
    };

    fetchTask();
    const interval = setInterval(() => {
      if (taskData?.status === 'processing') {
        fetchTask();
      }
    }, 5000);

    return () => clearInterval(interval);
  }, [taskId, taskData?.status]);

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <Loader2 className="animate-spin text-blue-500" size={48} />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center text-red-500 p-10 text-center">
        <div>
          <p className="text-xl font-bold mb-2">Lỗi tải dữ liệu</p>
          <p>{error}</p>
        </div>
      </div>
    );
  }

  const handleUpdateSubtitle = (index, updatedSub) => {
    setSubtitles((prev) => {
      const newList = [...prev];
      newList[index] = updatedSub;
      return newList;
    });
  };

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      <div className="flex flex-1 overflow-hidden">
        <VideoPlayer 
          taskData={taskData} 
          subtitles={subtitles} 
          renderSettings={renderSettings} 
          onUpdateSetting={updateRenderSetting}
          onOpenVoiceModal={() => setIsVoiceModalOpen(true)}
          onBack={() => onNavigate('dashboard')}
        />
        <RightPanel 
          taskId={taskId} 
          initialTaskData={taskData} 
          subtitles={subtitles} 
          onUpdateSubtitle={handleUpdateSubtitle}
          renderSettings={renderSettings}
          onUpdateSetting={updateRenderSetting}
          voiceSettings={voiceSettings}
        />
      </div>
      <BottomTimeline 
        taskData={taskData} 
        subtitles={subtitles} 
        onUpdateSubtitle={handleUpdateSubtitle}
      />

      <VoiceConfigModal 
        isOpen={isVoiceModalOpen}
        onClose={() => setIsVoiceModalOpen(false)}
        settings={voiceSettings}
        onSave={(newSettings) => setVoiceSettings(newSettings)}
      />
    </div>
  );
}
