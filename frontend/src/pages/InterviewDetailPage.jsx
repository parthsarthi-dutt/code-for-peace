import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { getInterviewDetail } from '../api/client';
import { ArrowLeft, Copy, Check } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import './InterviewPage.css';

const CodeBlock = ({ inline, className, children, ...props }) => {
  const match = /language-(\w+)/.exec(className || '');
  const [copied, setCopied] = useState(false);
  
  if (!inline && match) {
    const codeString = String(children).replace(/\n$/, '');
    const handleCopy = () => {
      navigator.clipboard.writeText(codeString);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    };
    return (
      <div style={{ position: 'relative', marginTop: 16, marginBottom: 16, borderRadius: 8, overflow: 'hidden', border: '1px solid var(--border-default)' }}>
        <div style={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center', 
          background: 'rgba(0,0,0,0.3)', 
          padding: '6px 12px',
          borderBottom: '1px solid var(--border-subtle)'
        }}>
          <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase', fontWeight: 600 }}>{match[1]}</span>
          <button 
            onClick={handleCopy}
            style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 6, fontSize: '0.75rem', padding: '4px 8px', borderRadius: 4, transition: 'all 0.2s' }}
            onMouseOver={e => e.currentTarget.style.color = 'var(--text-primary)'}
            onMouseOut={e => e.currentTarget.style.color = 'var(--text-muted)'}
          >
            {copied ? <Check size={14} color="#4ade80" /> : <Copy size={14} />}
            {copied ? 'Copied!' : 'Copy'}
          </button>
        </div>
        <SyntaxHighlighter
          style={vscDarkPlus}
          language={match[1]}
          PreTag="div"
          customStyle={{ margin: 0, background: 'transparent', padding: '16px' }}
          {...props}
        >
          {codeString}
        </SyntaxHighlighter>
      </div>
    );
  }
  return <code className={className} style={{ background: 'rgba(255,255,255,0.1)', padding: '2px 4px', borderRadius: 4, fontFamily: 'monospace', fontSize: '0.9em' }} {...props}>{children}</code>;
};

export default function InterviewDetailPage() {
  const { id } = useParams();
  const { isAuthenticated, loading: authLoading } = useAuth();
  const navigate = useNavigate();
  const [interview, setInterview] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!isAuthenticated && !authLoading) {
      navigate('/');
      return;
    }
    if (isAuthenticated) {
      getInterviewDetail(id)
        .then(setInterview)
        .catch(() => navigate('/interview'))
        .finally(() => setLoading(false));
    }
  }, [id, isAuthenticated, authLoading, navigate]);

  if (authLoading || loading) {
    return (
      <div className="page-content">
        <div className="interview-container" style={{ padding: '60px 0', textAlign: 'center' }}>
          <div className="skeleton" style={{ height: 200, width: '100%', borderRadius: 12 }} />
        </div>
      </div>
    );
  }

  if (!interview) return null;

  return (
    <div className="page-content fade-in">
      <div className="interview-container">
        <button
          onClick={() => navigate('/interview')}
          style={{
            background: 'none',
            border: 'none',
            color: 'var(--text-muted)',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            fontSize: '0.85rem',
            marginBottom: 20,
          }}
        >
          <ArrowLeft size={16} /> Back to Interviews
        </button>

        <div className="interview-hero">
          <h1><span>{interview.level} Interview</span></h1>
          <p>{interview.duration} minutes • {interview.tokens_deducted} tokens • {interview.status}</p>
        </div>

        {/* Chat transcript */}
        <div className="interview-chat" style={{ maxHeight: 500, marginBottom: 20 }}>
          {(interview.history || []).map((msg, i) => (
            <div key={i} className={`chat-bubble ${msg.role}`}>
              <div className="chat-role">
                {msg.role === 'interviewer' ? '🤖 Interviewer' : '👤 You'}
              </div>
              {msg.text}
            </div>
          ))}
        </div>

        {/* Feedback */}
        {interview.feedback && (
          <div className="interview-feedback">
            <h2>📋 Performance Feedback</h2>
            <div className="feedback-content">
              <ReactMarkdown components={{ code: CodeBlock }}>{interview.feedback}</ReactMarkdown>
            </div>
            
          </div>
        )}
      </div>
    </div>
  );
}
