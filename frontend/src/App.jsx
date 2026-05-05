import React, { useState } from 'react';
import Sidebar from './components/Sidebar';
import DashboardPage from './pages/DashboardPage';
import CreateProjectPage from './pages/CreateProjectPage';
import SettingsPage from './pages/SettingsPage';
import EditorPage from './pages/EditorPage';

function App() {
  const [currentPage, setCurrentPage] = useState('dashboard');
  const [editorTaskId, setEditorTaskId] = useState(null);

  const [error, setError] = useState(null);

  const handleNavigate = (page, taskId = null) => {
    try {
      setCurrentPage(page);
      if (page === 'editor' && taskId) {
        setEditorTaskId(taskId);
      }
    } catch (e) {
      setError(e.message);
    }
  };

  if (error) {
    return <div className="p-20 text-red-500">App Crash: {error}</div>;
  }

  const renderPage = () => {
    try {
      switch (currentPage) {
        case 'dashboard':
          return <DashboardPage onNavigate={handleNavigate} />;
        case 'create':
          return <CreateProjectPage onNavigate={handleNavigate} />;
        case 'settings':
          return <SettingsPage onNavigate={handleNavigate} />;
        case 'editor':
          return <EditorPage taskId={editorTaskId} onNavigate={handleNavigate} />;
        default:
          return <DashboardPage onNavigate={handleNavigate} />;
      }
    } catch (e) {
      return <div className="p-20 text-red-500">Page Render Crash: {e.message}</div>;
    }
  };

  return (
    <div className="flex h-screen w-full overflow-hidden bg-slate-50 text-slate-800">
      <Sidebar currentPage={currentPage} onNavigate={handleNavigate} />
      {renderPage()}
    </div>
  );
}

export default App;
