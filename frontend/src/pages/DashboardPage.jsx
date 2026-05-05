import React, { useState, useEffect } from 'react';
import { Plus, Loader2, CheckCircle, AlertCircle, Clock, Trash2, ArrowRight, FileText } from 'lucide-react';
import { getSubtitleTask } from '../utils/api';

export default function DashboardPage({ onNavigate }) {
  const [projects, setProjects] = useState([]);

  // Load projects from localStorage
  useEffect(() => {
    const saved = JSON.parse(localStorage.getItem('mk_projects') || '[]');
    setProjects(saved);
  }, []);

  // Poll status for active projects
  useEffect(() => {
    const activeProjects = projects.filter(
      (p) => p.status === 'processing' || p.status === 'waiting_review'
    );
    if (activeProjects.length === 0) return;

    const interval = setInterval(async () => {
      let updated = false;
      const newProjects = [...projects];

      for (const proj of activeProjects) {
        try {
          const data = await getSubtitleTask(proj.taskId);
          const idx = newProjects.findIndex((p) => p.taskId === proj.taskId);
          if (idx !== -1) {
            newProjects[idx] = {
              ...newProjects[idx],
              status: data.status,
              percent: data.process_percent,
              subtitleInfo: data.subtitle_info,
            };
            updated = true;
          }
        } catch (err) {
          // silently ignore polling errors
        }
      }

      if (updated) {
        setProjects(newProjects);
        localStorage.setItem('mk_projects', JSON.stringify(newProjects));
      }
    }, 3000);

    return () => clearInterval(interval);
  }, [projects]);

  const handleDelete = (taskId) => {
    if (!confirm('Bạn có chắc muốn xóa dự án này?')) return;
    const updated = projects.filter((p) => p.taskId !== taskId);
    setProjects(updated);
    localStorage.setItem('mk_projects', JSON.stringify(updated));
  };

  const statusConfig = {
    processing: { label: 'Đang xử lý', color: 'text-blue-600 bg-blue-50', icon: <Loader2 size={14} className="animate-spin" /> },
    waiting_review: { label: 'Chờ duyệt', color: 'text-amber-600 bg-amber-50', icon: <Clock size={14} /> },
    success: { label: 'Hoàn thành', color: 'text-green-600 bg-green-50', icon: <CheckCircle size={14} /> },
    failed: { label: 'Thất bại', color: 'text-red-600 bg-red-50', icon: <AlertCircle size={14} /> },
  };

  return (
    <div className="flex-1 overflow-y-auto bg-slate-50 p-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Dự án lồng tiếng</h1>
          <p className="text-sm text-slate-500 mt-1">Quản lý tất cả dự án dịch và lồng tiếng video của bạn</p>
        </div>
        <button
          onClick={() => onNavigate('create')}
          className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-lg transition-colors shadow-sm shadow-blue-200 cursor-pointer"
        >
          <Plus size={18} />
          Tạo dự án mới
        </button>
      </div>

      {/* Empty state */}
      {projects.length === 0 && (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-20 h-20 bg-slate-100 rounded-2xl flex items-center justify-center mb-6">
            <FileText size={36} className="text-slate-300" />
          </div>
          <h3 className="text-lg font-semibold text-slate-600 mb-2">Chưa có dự án nào</h3>
          <p className="text-sm text-slate-400 mb-6 max-w-sm">
            Bấm nút "Tạo dự án mới" để bắt đầu dịch và lồng tiếng video đầu tiên của bạn.
          </p>
          <button
            onClick={() => onNavigate('create')}
            className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-lg transition-colors cursor-pointer"
          >
            <Plus size={18} />
            Tạo dự án mới
          </button>
        </div>
      )}

      {/* Project grid */}
      {projects.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {projects.map((proj) => {
            const sc = statusConfig[proj.status] || statusConfig.processing;
            return (
              <div
                key={proj.taskId}
                className="bg-white border border-slate-200 rounded-xl p-5 hover:border-blue-300 hover:shadow-md transition-all duration-200 group"
              >
                {/* Title + delete */}
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1 min-w-0">
                    <h3 className="font-semibold text-sm text-slate-800 truncate">
                      {proj.name || proj.taskId}
                    </h3>
                    <p className="text-[11px] text-slate-400 font-mono mt-0.5 truncate">
                      ID: {proj.taskId}
                    </p>
                  </div>
                  <button
                    onClick={(e) => { e.stopPropagation(); handleDelete(proj.taskId); }}
                    className="text-slate-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-all cursor-pointer p-1"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>

                {/* Status badge */}
                <div className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-[11px] font-medium ${sc.color}`}>
                  {sc.icon}
                  {sc.label}
                  {proj.status === 'processing' && proj.percent > 0 && ` ${proj.percent}%`}
                </div>

                {/* Progress bar (processing only) */}
                {proj.status === 'processing' && (
                  <div className="mt-3 h-1.5 bg-slate-100 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-blue-500 rounded-full transition-all duration-500"
                      style={{ width: `${proj.percent || 0}%` }}
                    />
                  </div>
                )}

                {/* Action buttons */}
                <div className="mt-4 pt-3 border-t border-slate-100 flex items-center justify-end gap-2">
                  {proj.status === 'waiting_review' && (
                    <button
                      onClick={() => onNavigate('editor', proj.taskId)}
                      className="flex items-center gap-1.5 text-xs font-medium text-amber-600 hover:text-amber-700 cursor-pointer"
                    >
                      Review phụ đề
                      <ArrowRight size={14} />
                    </button>
                  )}
                  {proj.status === 'success' && (
                    <button
                      onClick={() => onNavigate('editor', proj.taskId)}
                      className="flex items-center gap-1.5 text-xs font-medium text-green-600 hover:text-green-700 cursor-pointer"
                    >
                      Xem kết quả
                      <ArrowRight size={14} />
                    </button>
                  )}
                  {proj.status === 'processing' && (
                    <span className="text-[11px] text-slate-400">Đang xử lý...</span>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
