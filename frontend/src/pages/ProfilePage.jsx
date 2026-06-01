import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { useNavigate } from 'react-router-dom';
import { getUserProfile, updateUserProfile } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { formatDate } from '../utils/helpers';
import Heatmap from '../components/Heatmap';
import VerdictBadge from '../components/VerdictBadge';
import { X, Copy, Check } from 'lucide-react';
import CodeEditor from '@monaco-editor/react';
import './ProfilePage.css';

export default function ProfilePage() {
  const { user, isAuthenticated, loading: authLoading, updateLocalUser } = useAuth();
  const navigate = useNavigate();
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [selectedSolve, setSelectedSolve] = useState(null);
  
  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState('');
  const [editAvatar, setEditAvatar] = useState('');
  const [savingProfile, setSavingProfile] = useState(false);

  const PRESET_AVATARS = [
    'https://api.dicebear.com/7.x/bottts/svg?seed=Felix',
    'https://api.dicebear.com/7.x/bottts/svg?seed=Aneka',
    'https://api.dicebear.com/7.x/bottts/svg?seed=Jack',
    'https://api.dicebear.com/7.x/bottts/svg?seed=Spooky',
    'https://api.dicebear.com/7.x/pixel-art/svg?seed=John',
    'https://api.dicebear.com/7.x/pixel-art/svg?seed=Jane',
    'https://api.dicebear.com/7.x/avataaars/svg?seed=Mia',
    'https://api.dicebear.com/7.x/avataaars/svg?seed=Leo',
  ];

  const handleStartEdit = () => {
    setEditName(user?.username || '');
    setEditAvatar(user?.avatar || '');
    setIsEditing(true);
  };

  const handleSaveProfile = async () => {
    if (!editName.trim()) {
      alert('Username cannot be empty');
      return;
    }
    setSavingProfile(true);
    try {
      await updateUserProfile(editName, editAvatar);
      updateLocalUser({ username: editName, avatar: editAvatar });
      setIsEditing(false);
      if (profile) {
        setProfile(prev => ({ ...prev, username: editName }));
      }
    } catch (err) {
      console.error('Failed to update profile:', err);
      alert('Failed to update profile');
    } finally {
      setSavingProfile(false);
    }
  };

  useEffect(() => {
    // Wait for auth to finish loading before checking
    if (authLoading) return;

    if (!isAuthenticated) {
      navigate('/');
      return;
    }

    async function loadProfile() {
      try {
        const data = await getUserProfile();
        setProfile(data);
      } catch (err) {
        console.error('Failed to load profile:', err);
      } finally {
        setLoading(false);
      }
    }
    loadProfile();
  }, [isAuthenticated, authLoading, navigate]);

  if (loading) {
    return (
      <div className="page-content">
        <div className="container">
          <div className="profile-loading">
            <div className="skeleton" style={{ height: 120, width: '100%', marginBottom: 24 }} />
            <div className="skeleton" style={{ height: 80, width: '100%', marginBottom: 24 }} />
            <div className="skeleton" style={{ height: 200, width: '100%' }} />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="page-content fade-in">
      <div className="container">
        {/* ─── User Header ─── */}
        <div className="profile-header" id="profile-header">
          <img
            src={user?.avatar || 'https://ui-avatars.com/api/?name=User&background=1a1a1a&color=e8e8e8&size=80'}
            alt="avatar"
            className="profile-avatar"
          />
          <div className="profile-info">
            <div className="profile-name-row">
              <h1 className="profile-name">{user?.username || 'User'}</h1>
              <button className="btn-edit-profile" onClick={handleStartEdit}>
                Edit Profile
              </button>
            </div>
            <p className="profile-email">{user?.email || ''}</p>
          </div>
        </div>

        {/* ─── Stats Cards ─── */}
        <div className="stats-grid" id="stats-grid">
          <div className="stat-card stat-total">
            <div className="stat-number">{profile?.total_solved || 0}</div>
            <div className="stat-label">Total Solved</div>
          </div>
          <div className="stat-card stat-easy">
            <div className="stat-number">{profile?.easy_solved || 0}</div>
            <div className="stat-label">Easy</div>
            <div className="stat-bar"><div className="stat-fill easy" /></div>
          </div>
          <div className="stat-card stat-medium">
            <div className="stat-number">{profile?.medium_solved || 0}</div>
            <div className="stat-label">Medium</div>
            <div className="stat-bar"><div className="stat-fill medium" /></div>
          </div>
          <div className="stat-card stat-hard">
            <div className="stat-number">{profile?.hard_solved || 0}</div>
            <div className="stat-label">Hard</div>
            <div className="stat-bar"><div className="stat-fill hard" /></div>
          </div>
          <div className="stat-card stat-streak">
            <div className="stat-number">🔥 {profile?.current_streak || 0}</div>
            <div className="stat-label">Current Streak</div>
          </div>
          <div className="stat-card stat-streak">
            <div className="stat-number">🏆 {profile?.highest_streak || 0}</div>
            <div className="stat-label">Best Streak</div>
          </div>
        </div>

        {/* ─── Heatmap ─── */}
        <div className="profile-section">
          <h2 className="section-title">Submission Activity</h2>
          <Heatmap data={profile?.heatmap_data || {}} />
        </div>

        {/* ─── Recent Solves ─── */}
        {profile?.recent_solves?.length > 0 && (
          <div className="profile-section">
            <h2 className="section-title">Recently Solved</h2>
            <div className="recent-table" id="recent-solves">
              {profile.recent_solves.map((solve, i) => (
                <div
                  key={i}
                  className={`recent-row ${i % 2 === 0 ? 'even' : ''}`}
                  onClick={() => setSelectedSolve(solve)}
                >
                  <span className="recent-problem">{solve.problem_id}</span>
                  <VerdictBadge verdict={solve.verdict} />
                  <span className="recent-date">{formatDate(solve.solved_at)}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {selectedSolve && createPortal(
        <div className="modal-overlay" onClick={() => setSelectedSolve(null)}>
          <div className="modal-content solve-modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2>{selectedSolve.problem_id} - Solution</h2>
              <button className="btn-close" onClick={() => setSelectedSolve(null)}>
                <X size={20} />
              </button>
            </div>
            <div className="modal-body">
              <div className="solve-stats">
                <VerdictBadge verdict={selectedSolve.verdict} />
                <span className="stat-badge">⏱ {selectedSolve.execution_time}ms</span>
                <span className="stat-badge">💾 {selectedSolve.memory_used}KB</span>
                <span className="stat-badge">📅 {formatDate(selectedSolve.solved_at)}</span>
              </div>
              <div style={{ position: 'relative', marginTop: 16, border: '1px solid var(--border-default)', borderRadius: 8, overflow: 'hidden' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 16px', background: 'var(--bg-tertiary)', borderBottom: '1px solid var(--border-default)' }}>
                  <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)', textTransform: 'uppercase', fontWeight: 600 }}>Code</span>
                  <button 
                    onClick={(e) => {
                      navigator.clipboard.writeText(selectedSolve.code);
                      const target = e.currentTarget;
                      const original = target.innerHTML;
                      target.innerHTML = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#4ade80" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg> Copied!';
                      setTimeout(() => target.innerHTML = original, 2000);
                    }}
                    style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 6, fontSize: '0.85rem' }}
                  >
                    <Copy size={14} /> Copy
                  </button>
                </div>
                <div style={{ height: '350px' }}>
                  <CodeEditor 
                    value={selectedSolve.code || ''} 
                    language="cpp"
                    theme="vs-dark"
                    options={{
                      readOnly: true,
                      minimap: { enabled: false },
                      scrollBeyondLastLine: false,
                      padding: { top: 16, bottom: 16 }
                    }}
                  />
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <button 
                className="btn-primary"
                onClick={() => navigate(`/problem/${selectedSolve.problem_id}`)}
              >
                Go to Problem
              </button>
            </div>
          </div>
        </div>,
        document.body
      )}

      {isEditing && createPortal(
        <div className="modal-overlay" onClick={() => setIsEditing(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Edit Profile</h2>
              <button className="btn-close" onClick={() => setIsEditing(false)}>
                <X size={20} />
              </button>
            </div>
            <div className="modal-body">
              <div className="edit-profile-form">
                <div className="avatar-preview-container">
                  <img
                    src={editAvatar || 'https://ui-avatars.com/api/?name=User&background=1a1a1a&color=e8e8e8&size=80'}
                    alt="preview"
                    className="avatar-preview-img"
                  />
                  <div className="avatar-preview-info">
                    <span className="avatar-preview-title">Avatar Preview</span>
                    <span className="avatar-preview-desc">Choose a preset below or enter a custom image URL</span>
                  </div>
                </div>

                <div className="form-group">
                  <label htmlFor="username-input">Username</label>
                  <input
                    id="username-input"
                    type="text"
                    className="form-input"
                    value={editName}
                    onChange={e => setEditName(e.target.value)}
                    placeholder="Enter username"
                  />
                </div>

                <div className="form-group">
                  <label htmlFor="avatar-url-input">Custom Avatar URL</label>
                  <input
                    id="avatar-url-input"
                    type="text"
                    className="form-input"
                    value={editAvatar}
                    onChange={e => setEditAvatar(e.target.value)}
                    placeholder="Enter custom image URL"
                  />
                </div>

                <div className="form-group">
                  <label>Preset Avatars</label>
                  <div className="avatar-presets-grid">
                    {PRESET_AVATARS.map((presetUrl, idx) => (
                      <div
                        key={idx}
                        className={`avatar-preset-item ${editAvatar === presetUrl ? 'active' : ''}`}
                        onClick={() => setEditAvatar(presetUrl)}
                      >
                        <img src={presetUrl} alt={`preset-${idx}`} className="avatar-preset-img" />
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <button
                className="btn-secondary"
                style={{
                  marginRight: 12,
                  padding: '8px 16px',
                  background: 'transparent',
                  border: '1px solid var(--border-default)',
                  color: 'var(--text-secondary)',
                  borderRadius: 'var(--radius-md)',
                  cursor: 'pointer',
                  fontWeight: 600,
                }}
                onClick={() => setIsEditing(false)}
                disabled={savingProfile}
              >
                Cancel
              </button>
              <button
                className="btn-primary"
                onClick={handleSaveProfile}
                disabled={savingProfile}
              >
                {savingProfile ? 'Saving...' : 'Save Changes'}
              </button>
            </div>
          </div>
        </div>,
        document.body
      )}
    </div>
  );
}
