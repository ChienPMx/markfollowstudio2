import React from 'react';
import { Home, Plus, Settings, MessageSquare, Phone, Sun, Moon } from 'lucide-react';

export default function Sidebar({ currentPage, onNavigate }) {
  return (
    <div className="w-60 min-w-60 border-r border-slate-200 bg-white flex flex-col justify-between h-full">
      {/* Logo */}
      <div>
        <div className="p-4 flex items-center gap-3 border-b border-slate-100">
          <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-sm shadow-sm">
            MK
          </div>
          <div>
            <h1 className="font-bold text-sm text-slate-800">MarkFlow Studio</h1>
            <p className="text-[11px] text-slate-400">Lồng tiếng AI</p>
          </div>
        </div>

        {/* Navigation */}
        <nav className="p-2 space-y-0.5">
          <NavItem
            icon={<Home size={18} />}
            label="Dự án lồng tiếng"
            active={currentPage === 'dashboard' || currentPage === 'editor'}
            onClick={() => onNavigate('dashboard')}
          />
          <NavItem
            icon={<Plus size={18} />}
            label="Tạo dự án mới"
            active={currentPage === 'create'}
            onClick={() => onNavigate('create')}
          />
          <NavItem
            icon={<Settings size={18} />}
            label="Cài đặt"
            active={currentPage === 'settings'}
            onClick={() => onNavigate('settings')}
          />
        </nav>
      </div>

      {/* Bottom section */}
      <div className="p-2 border-t border-slate-100">
        <NavItem
          icon={<MessageSquare size={18} />}
          label="Cộng đồng"
          onClick={() => window.open('https://discord.gg', '_blank')}
        />
        <NavItem
          icon={<Phone size={18} />}
          label="Liên hệ"
          onClick={() => window.open('mailto:support@mkstudio.ai', '_blank')}
        />

        {/* Theme toggle */}
        <div className="px-3 py-2.5 flex items-center justify-between text-sm text-slate-500 mt-1">
          <div className="flex items-center gap-3">
            <Sun size={18} />
            <span className="text-xs">Sáng</span>
          </div>
          <div className="w-9 h-5 bg-slate-200 rounded-full relative cursor-pointer transition-colors hover:bg-slate-300">
            <div className="w-4 h-4 bg-white rounded-full absolute left-0.5 top-0.5 shadow transition-transform"></div>
          </div>
        </div>

        {/* User */}
        <div className="mt-1 px-3 py-2.5 flex items-center gap-3 cursor-pointer hover:bg-slate-50 rounded-lg transition-colors">
          <div className="w-7 h-7 rounded-full bg-gradient-to-br from-blue-500 to-indigo-500 text-white flex items-center justify-center text-xs font-bold shadow-sm">
            C
          </div>
          <span className="text-xs font-medium text-slate-700 truncate">chienphammi...</span>
        </div>
      </div>
    </div>
  );
}

function NavItem({ icon, label, active = false, onClick }) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-150 cursor-pointer ${active
          ? 'bg-blue-50 text-blue-600 shadow-sm'
          : 'text-slate-600 hover:bg-slate-50 hover:text-slate-800'
        }`}
    >
      {icon}
      <span>{label}</span>
    </button>
  );
}
