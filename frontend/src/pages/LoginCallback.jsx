import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function LoginCallback() {
  const [searchParams] = useSearchParams();
  const { login } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const token = searchParams.get('token');
    const userId = searchParams.get('user_id');
    const username = searchParams.get('username');
    const avatar = searchParams.get('avatar');
    const email = searchParams.get('email');
    const tokens = searchParams.get('tokens');

    if (token && userId) {
      login(token, {
        user_id: parseInt(userId),
        username: decodeURIComponent(username || ''),
        avatar: decodeURIComponent(avatar || ''),
        email: decodeURIComponent(email || ''),
        tokens: parseInt(tokens || '0'),
      });
      window.location.replace('/');
    } else {
      console.error('Missing auth parameters');
      window.location.replace('/');
    }
  }, [searchParams, login]);

  return (
    <div className="page-content" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <div style={{ textAlign: 'center', color: 'var(--text-secondary)' }}>
        <div className="spinner" style={{
          width: 32,
          height: 32,
          border: '3px solid var(--border-secondary)',
          borderTopColor: 'var(--text-primary)',
          borderRadius: '50%',
          animation: 'spin 0.6s linear infinite',
          margin: '0 auto 16px',
        }} />
        <p>Signing you in...</p>
      </div>
    </div>
  );
}
