import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { getGoogleLoginURL } from '../api/client';
import { Flame, X, Coins } from 'lucide-react';
import './Navbar.css';

export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();
  const [isTokenModalOpen, setIsTokenModalOpen] = useState(false);

  const handleLogin = () => {
    window.location.href = getGoogleLoginURL();
  };

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <nav className="navbar" id="main-navbar">
      <div className="navbar-inner">
        <Link to="/" className="navbar-brand">
          <span className="brand-icon">⚡</span>
          <span className="brand-text">CodeForPeace</span>
        </Link>

        <div className="navbar-links">
          <Link to="/" className="nav-link">Problems</Link>
          <Link to="/interview" state={{ reset: Date.now() }} className="nav-link">AI Interview</Link>
        </div>

        <div className="navbar-actions">
          {isAuthenticated ? (
            <div className="user-menu">
              <span className={`user-streak ${(user?.current_streak || 0) > 0 ? 'active' : ''}`} title={`Current Streak: ${user?.current_streak || 0} days. Solve an unsolved problem to increase your streak!`}>
                <Flame size={16} fill="currentColor" className="streak-fire" />
                {user?.current_streak || 0}
              </span>
              <span 
                className="user-tokens" 
                title="Your tokens"
                onClick={() => setIsTokenModalOpen(true)}
                style={{ cursor: 'pointer' }}
              >
                🪙 {user?.tokens || 0}
              </span>
              <Link to="/profile" className="user-profile-link" id="navbar-profile-link">
                <img
                  src={user?.avatar || 'https://ui-avatars.com/api/?name=User&background=1a1a1a&color=e8e8e8&size=32'}
                  alt="avatar"
                  className="user-avatar"
                />
                <span className="user-name">{user?.username || 'User'}</span>
              </Link>
              <button onClick={handleLogout} className="btn-logout" id="logout-btn">
                Sign Out
              </button>
            </div>
          ) : (
            <button onClick={handleLogin} className="btn-login" id="login-btn">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" />
                <polyline points="10 17 15 12 10 7" />
                <line x1="15" y1="12" x2="3" y2="12" />
              </svg>
              Sign In
            </button>
          )}
        </div>
      </div>

      {/* Token Modal */}
      {isTokenModalOpen && (
        <div className="token-modal-overlay" onClick={() => setIsTokenModalOpen(false)}>
          <div className="token-modal-content" onClick={e => e.stopPropagation()}>
            <button className="token-modal-close" onClick={() => setIsTokenModalOpen(false)}>
              <X size={20} />
            </button>
            
            <div className="token-modal-header">
              <Coins size={28} />
              How to Earn Tokens
            </div>

            <div className="token-reward-list">
              <div className="token-reward-item">
                <span className="token-reward-desc">Solve a new problem</span>
                <span className="token-reward-amount">+2</span>
              </div>
              <div className="token-reward-item">
                <span className="token-reward-desc">First 10 problems solved</span>
                <span className="token-reward-amount">+20</span>
              </div>
              <div className="token-reward-item">
                <span className="token-reward-desc">First 50 problems solved</span>
                <span className="token-reward-amount">+50</span>
              </div>
              <div className="token-reward-item">
                <span className="token-reward-desc">First 100 problems solved</span>
                <span className="token-reward-amount">+100</span>
              </div>
              <div className="token-reward-item">
                <span className="token-reward-desc">First 500 problems solved</span>
                <span className="token-reward-amount">+500</span>
              </div>
              <div className="token-reward-item">
                <span className="token-reward-desc">First 1000 problems solved</span>
                <span className="token-reward-amount">+1000</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </nav>
  );
}
