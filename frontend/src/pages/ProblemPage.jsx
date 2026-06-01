import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { Panel, Group as PanelGroup, Separator as PanelResizeHandle } from 'react-resizable-panels';
import { Play, Check, X, Clock, HardDrive, Code2, TriangleAlert, Settings, TerminalSquare, FileText, List, RotateCcw, Copy } from 'lucide-react';
import { fetchProblem, submitCode, getSubmission, getUserSubmissions, getAIHint, getAIFeedback, runCode } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { generateSubmissionId, formatDateTime, getVerdictInfo } from '../utils/helpers';
import CodeEditor from '../components/CodeEditor';
import VerdictBadge from '../components/VerdictBadge';
import ReactMarkdown from 'react-markdown';
import remarkMath from 'remark-math';
import rehypeKatex from 'rehype-katex';
import 'katex/dist/katex.min.css';
import confetti from 'canvas-confetti';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import './ProblemPage.css';

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

const TEMPLATES = {
  cpp: `#include <bits/stdc++.h>
using namespace std;

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    
    // Your code here
    
    return 0;
}`,
  python: `# Your code here
def main():
    pass

if __name__ == '__main__':
    main()`,
  java: `import java.util.*;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        // Your code here
    }
}`
};

/** localStorage helpers for code persistence */
function getSavedCode(problemId, language) {
  return localStorage.getItem(`code_${problemId}_${language}`) || '';
}
function saveCode(problemId, language, code) {
  localStorage.setItem(`code_${problemId}_${language}`, code);
}
function getLastSubmittedCode(problemId) {
  return localStorage.getItem(`lastsubmit_${problemId}`) || '';
}
function saveLastSubmittedCode(problemId, code) {
  localStorage.setItem(`lastsubmit_${problemId}`, code);
}

function ResizeHandle({ direction = "horizontal" }) {
  return (
    <PanelResizeHandle className={`resize-handle ${direction}`}>
      <div className="resize-handle-inner" />
    </PanelResizeHandle>
  );
}

/** Parses raw AI text into bullet points for UI */
function renderAIResponse(text) {
  if (!text) return null;
  const lines = text.split('\n').filter(line => line.trim() !== '');
  
  // If it doesn't look like a list, just return text
  if (!lines.some(line => line.trim().startsWith('-') || line.trim().startsWith('*'))) {
    return <p>{text}</p>;
  }

  return (
    <ul className="ai-feedback-list">
      {lines.map((line, idx) => {
        let content = line.trim();
        if (content.startsWith('-') || content.startsWith('*')) {
          content = content.substring(1).trim();
        }
        
        // Bold headers (e.g. "Time Complexity:")
        const colonIndex = content.indexOf(':');
        if (colonIndex !== -1 && colonIndex < 35) {
          const title = content.substring(0, colonIndex + 1);
          const desc = content.substring(colonIndex + 1);
          return (
            <li key={idx}>
              <strong>{title}</strong> {desc}
            </li>
          );
        }
        return <li key={idx}>{content}</li>;
      })}
    </ul>
  );
}

export default function ProblemPage() {
  const { id } = useParams();
  const { user, isAuthenticated, login, refreshUser } = useAuth();
  const [problem, setProblem] = useState(null);
  const [code, setCode] = useState('');
  const [language, setLanguage] = useState('cpp');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [verdict, setVerdict] = useState(null);
  const [activeTab, setActiveTab] = useState('description');
  const [submissions, setSubmissions] = useState([]);
  const [subsLoading, setSubsLoading] = useState(false);
  const [selectedSubmission, setSelectedSubmission] = useState(null);
  const [editorSettings, setEditorSettings] = useState({ minimap: false, wordWrap: 'on' });
  
  // AI State
  const [aiHint, setAiHint] = useState('');
  const [aiLoading, setAiLoading] = useState(false);
  const [aiFeedback, setAiFeedback] = useState('');
  const [aiFeedbackLoading, setAiFeedbackLoading] = useState(false);

  // Run State
  const [running, setRunning] = useState(false);
  const [runResult, setRunResult] = useState(null);
  const [consoleTab, setConsoleTab] = useState('testcases'); // 'testcases' | 'runresult'
  
  const [showPointsAnimation, setShowPointsAnimation] = useState(false);
  const [editorial, setEditorial] = useState(null);
  const [editorialLocked, setEditorialLocked] = useState(false);
  const [editorialLoading, setEditorialLoading] = useState(false);

  const pollRef = useRef(null);

  // Load problem + saved code
  useEffect(() => {
    async function loadProblem() {
      try {
        const data = await fetchProblem(id);
        setProblem(data);
      } catch (err) {
        console.error('Failed to load problem:', err);
      } finally {
        setLoading(false);
      }
    }
    loadProblem();

    // Restore saved code or use template
    const saved = getSavedCode(id, language);
    setCode(saved || TEMPLATES[language]);

    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, [id]);

  // Auto-save code on change
  useEffect(() => {
    if (code && code !== TEMPLATES[language]) {
      saveCode(id, language, code);
    }
  }, [code, id, language]);

  // Load submissions when tab switches
  useEffect(() => {
    if (activeTab === 'submissions' && isAuthenticated && user?.user_id) {
      loadSubmissions();
    }
  }, [activeTab, isAuthenticated, user]);

  async function loadSubmissions() {
    setSubsLoading(true);
    try {
      const all = await getUserSubmissions(user.user_id);
      const filtered = (all || [])
        .filter(s => s.problem_id === id)
        .sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
      setSubmissions(filtered);
    } catch (err) {
      console.error('Failed to load submissions:', err);
    } finally {
      setSubsLoading(false);
    }
  }

  // Fetch editorial
  useEffect(() => {
    if (activeTab === 'editorial' && isAuthenticated) {
      loadEditorial();
    }
  }, [activeTab, isAuthenticated, id]);

  const loadEditorial = async () => {
    setEditorialLoading(true);
    try {
      const token = localStorage.getItem('token');
      const res = await fetch(`http://localhost:8080/api/problem/editorial?id=${id}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (res.status === 403) {
        setEditorialLocked(true);
        setEditorial(null);
      } else if (res.ok) {
        const data = await res.json();
        setEditorialLocked(false);
        setEditorial(data.content);
      }
    } catch (err) {
      console.error('Failed to fetch editorial', err);
    } finally {
      setEditorialLoading(false);
    }
  };

  const handleUnlockEditorial = async () => {
    try {
      const token = localStorage.getItem('token');
      const res = await fetch('http://localhost:8080/api/problem/editorial/unlock', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ problem_id: id })
      });

      if (res.ok) {
        await loadEditorial();
        refreshUser();
        alert('Editorial unlocked! 10 coins deducted.');
      } else if (res.status === 402 || res.status === 412 || res.status === 402) {
        alert('Not enough coins to unlock this editorial.');
      }
    } catch (err) {
      console.error('Failed to unlock editorial', err);
    }
  };

  const pollForResult = useCallback((submissionId) => {
    pollRef.current = setInterval(async () => {
      try {
        const result = await getSubmission(submissionId);
        if (result.verdict && result.verdict !== 'pending') {
          setVerdict(result);
          setSubmitting(false);

          if (result.verdict === 'Compilation Error' || result.verdict === 'Runtime Error') {
            setRunResult({
              status: result.verdict,
              error: result.message
            });
            setConsoleTab('runresult');
          } else {
            setSelectedSubmission(result);
          }

          if (result.verdict === 'Accepted') {
            // Fire confetti with z-index above the modal (z-index: 1000)
            const myCanvas = document.createElement('canvas');
            myCanvas.style.position = 'fixed';
            myCanvas.style.top = '0';
            myCanvas.style.left = '0';
            myCanvas.style.width = '100vw';
            myCanvas.style.height = '100vh';
            myCanvas.style.pointerEvents = 'none';
            myCanvas.style.zIndex = '99999';
            document.body.appendChild(myCanvas);

            const myConfetti = confetti.create(myCanvas, { resize: true });
            
            // Fire multiple celebration bursts
            myConfetti({ particleCount: 80, spread: 70, origin: { x: 0.3, y: 0.6 } });
            setTimeout(() => myConfetti({ particleCount: 80, spread: 70, origin: { x: 0.7, y: 0.6 } }), 200);
            setTimeout(() => myConfetti({ particleCount: 100, spread: 100, origin: { x: 0.5, y: 0.5 } }), 500);
            
            // Cleanup canvas after animations finish
            setTimeout(() => { myCanvas.remove(); }, 4000);

            if (result.tokens_awarded && result.tokens_awarded > 0) {
              setShowPointsAnimation(true);
              setTimeout(() => setShowPointsAnimation(false), 3000);
            }
            refreshUser();
          }

          clearInterval(pollRef.current);
          pollRef.current = null;
          // Refresh submissions list
          if (isAuthenticated && user?.user_id) loadSubmissions();
        }
      } catch (err) {
        console.error('Poll error:', err);
      }
    }, 2000);
  }, [isAuthenticated, user, refreshUser]);

  const handleSubmit = async () => {
    if (!isAuthenticated) {
      alert('Please sign in to submit');
      return;
    }
    if (!code.trim()) {
      alert('Please write some code first');
      return;
    }

    setSubmitting(true);
    setVerdict(null);
    const submissionId = generateSubmissionId(id, user.user_id);
    saveLastSubmittedCode(id, code);

    try {
      await submitCode({
        submission_id: submissionId,
        problem_id: id,
        user_id: String(user.user_id),
        language: language,
        code: code,
      });
      setVerdict({ verdict: 'pending' });
      pollForResult(submissionId);
    } catch (err) {
      console.error('Submit error:', err);
      setSubmitting(false);
      
      let msg = 'Submission failed';
      if (err.response && typeof err.response.data === 'string') {
        msg = err.response.data.replace('15 seconds', 'a few seconds');
      } else if (err.message) {
        msg = err.message;
      }
      
      setVerdict({ verdict: 'Error', message: msg });
      setSelectedSubmission({ 
        verdict: 'Error', 
        message: msg, 
        code: code, 
        created_at: new Date().toISOString() 
      });
    }
  };

  const handleRunCode = async () => {
    if (!isAuthenticated) return alert('Please sign in to run code');
    if (!code.trim()) return alert('Please write some code first');
    if (!problem?.sample_cases || problem.sample_cases.length === 0) return alert('No sample test case available to run against');

    setRunning(true);
    setRunResult(null);
    setConsoleTab('runresult');
    
    try {
      // Test against the first sample case
      const testCase = problem.sample_cases[0];
      const res = await runCode({
        language: language,
        code: code,
        input_data: testCase.input,
        time_limit: problem.time_limit
      });

      // Simple diff (normalize \r\n to \n for cross-platform comparison)
      const normalize = (s) => s.replace(/\r\n/g, '\n').replace(/\r/g, '\n').trim();
      const actualOut = normalize(res.stdout || '');
      const expectedOut = normalize(testCase.output || '');
      
      let status = "Accepted";
      if (res.is_tle) status = "Time Limit Exceeded";
      else if (res.stderr) status = "Runtime Error";
      else if (actualOut !== expectedOut) status = "Wrong Answer";

      setRunResult({
        status,
        input: testCase.input,
        expected: expectedOut,
        actual: actualOut,
        error: res.stderr,
      });

    } catch (err) {
      console.error('Run error:', err);
      let alertMsg = 'Failed to execute code';
      if (err.response && typeof err.response.data === 'string') {
        alertMsg = err.response.data.replace('10 seconds', 'a few seconds');
      }
      alert(alertMsg);
    } finally {
      setRunning(false);
    }
  };

  const handleGetHint = async () => {
    if (!isAuthenticated) return alert("Sign in first");
    if (user.tokens < 5) return alert("Not enough tokens (5 required)");
    setAiLoading(true);
    try {
      const res = await getAIHint(id, code);
      setAiHint(res.hint);
      login(localStorage.getItem('token'), { ...user, tokens: user.tokens - 5 });
    } catch (e) {
      alert(e.response?.data || "Failed to get hint");
    } finally {
      setAiLoading(false);
    }
  };

  const handleGetFeedback = async () => {
    if (!isAuthenticated) return alert("Sign in first");
    if (user.tokens < 3) return alert("Not enough tokens (3 required)");
    setAiFeedbackLoading(true);
    try {
      const submissionCode = selectedSubmission?.code || code;
      const res = await getAIFeedback(id, submissionCode);
      setAiFeedback(res.feedback);
      login(localStorage.getItem('token'), { ...user, tokens: user.tokens - 3 });
    } catch (e) {
      alert(e.response?.data || "Failed to get feedback");
    } finally {
      setAiFeedbackLoading(false);
    }
  };

  const handleReset = () => {
    if (confirm('Reset code to template? This cannot be undone.')) {
      setCode(TEMPLATES[language]);
      saveCode(id, language, TEMPLATES[language]);
    }
  };

  const handleLanguageChange = (e) => {
    const newLang = e.target.value;
    
    // Save current language's code before switching if it's not a template
    if (code !== TEMPLATES[language] && code !== '') {
        saveCode(id, language, code);
    }

    setLanguage(newLang);
    
    // Load new language's code
    const saved = getSavedCode(id, newLang);
    setCode(saved || TEMPLATES[newLang]);
  };

  const handleRetrieveLastSubmission = () => {
    const last = getLastSubmittedCode(id);
    if (last) {
      setCode(last);
    } else {
      alert('No previous submission found for this problem');
    }
  };

  if (loading) {
    return (
      <div className="page-content">
        <div className="problem-loading">
          <div className="problem-loading-left">
            <div className="skeleton" style={{ height: 28, width: '60%', marginBottom: 16 }} />
            <div className="skeleton" style={{ height: 16, width: '100%', marginBottom: 8 }} />
            <div className="skeleton" style={{ height: 16, width: '90%', marginBottom: 8 }} />
            <div className="skeleton" style={{ height: 16, width: '95%', marginBottom: 8 }} />
            <div className="skeleton" style={{ height: 16, width: '70%', marginBottom: 24 }} />
            <div className="skeleton" style={{ height: 120, width: '100%' }} />
          </div>
          <div className="problem-loading-right">
            <div className="skeleton" style={{ height: '100%', width: '100%' }} />
          </div>
        </div>
      </div>
    );
  }

  if (!problem) {
    return (
      <div className="page-content">
        <div className="container">
          <div className="not-found">
            <div className="not-found-icon">404</div>
            <h2>Problem not found</h2>
            <p>The problem "{id}" does not exist or could not be loaded.</p>
          </div>
        </div>
      </div>
    );
  }

  const isAccepted = verdict?.verdict === 'Accepted';
  const isWrong = verdict && verdict.verdict !== 'Accepted' && verdict.verdict !== 'pending' && verdict.verdict !== 'Error';

  return (
    <>
      <div className="page-content problem-page fade-in">
        <PanelGroup orientation="horizontal" className="problem-layout">
        {/* ─── Left Panel ─── */}
        <Panel defaultSize={45} minSize={25} className="problem-statement-panel">
          <div className="panel-header">
            <div className="panel-tabs">
              <button
                className={`panel-tab ${activeTab === 'description' ? 'active' : ''}`}
                onClick={() => setActiveTab('description')}
              >
                <FileText size={16} />
                Description
              </button>
              <button
                className={`panel-tab ${activeTab === 'submissions' ? 'active' : ''}`}
                onClick={() => setActiveTab('submissions')}
              >
                <List size={16} />
                Submissions
                {submissions.length > 0 && (
                  <span className="tab-badge">{submissions.length}</span>
                )}
              </button>
              <button
                className={`panel-tab ${activeTab === 'editorial' ? 'active' : ''}`}
                onClick={() => setActiveTab('editorial')}
              >
                <FileText size={16} />
                Editorial
              </button>
            </div>
          </div>

          <div className="panel-body">
            {activeTab === 'description' && (
              <div className="description-content" id="problem-description">
                {/* Title + Difficulty */}
                <div className="problem-title-bar">
                  <div className="title-left">
                    <span className="problem-id-tag">{problem.id}</span>
                    <h2>{problem.title || problem.id}</h2>
                  </div>
                  <span className={`difficulty-chip ${problem.difficulty?.toLowerCase()}`}>
                    {problem.difficulty}
                  </span>
                </div>

                {/* Constraints */}
                <div className="constraints-bar">
                  <div className="constraint">
                    <Clock size={16} />
                    <span>Time limit: <strong>{problem.time_limit}s</strong></span>
                  </div>
                  <div className="constraint">
                    <HardDrive size={16} />
                    <span>Memory limit: <strong>{problem.memory_limit} MB</strong></span>
                  </div>
                </div>

                {/* Statement Body using ReactMarkdown */}
                <div className="statement-section markdown-body">
                    <ReactMarkdown 
                      remarkPlugins={[remarkMath]}
                      rehypePlugins={[rehypeKatex]}
                      components={{ code: CodeBlock }}
                    >
                      {problem.statement || 'No problem statement available.'}
                    </ReactMarkdown>
                </div>

                {/* Sample Cases (Show only first one) */}
                {problem.sample_cases?.length > 0 && (
                  <div className="statement-section">
                    <h3 className="section-heading">
                      <span className="heading-accent" />
                      Example
                    </h3>
                    {problem.sample_cases.slice(0, 1).map((sample, i) => (
                      <div key={i} className="sample-case">
                        <div className="sample-block">
                          <div className="sample-header">
                            <span className="sample-label">Input</span>
                            <button
                              className="copy-btn"
                              onClick={() => { navigator.clipboard.writeText(sample.input); }}
                              title="Copy input"
                            >
                              <Copy size={14} />
                            </button>
                          </div>
                          <pre className="sample-content">{sample.input}</pre>
                        </div>
                        <div className="sample-block">
                          <div className="sample-header">
                            <span className="sample-label">Output</span>
                            <button
                              className="copy-btn"
                              onClick={() => { navigator.clipboard.writeText(sample.output); }}
                              title="Copy output"
                            >
                              <Copy size={14} />
                            </button>
                          </div>
                          <pre className="sample-content">{sample.output}</pre>
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {/* Tags */}
                {problem.tags?.length > 0 && (
                  <div className="statement-section tags-section">
                    <details className="topics-dropdown">
                      <summary className="tags-summary">Topics</summary>
                      <div className="tags-list" style={{ marginTop: '10px' }}>
                        {problem.tags.map((tag) => (
                          <span key={tag} className="tag-pill">{tag}</span>
                        ))}
                      </div>
                    </details>
                  </div>
                )}
              </div>
            )}

            {/* Submissions Tab */}
            {activeTab === 'submissions' && (
              <div className="submissions-tab">
                {!isAuthenticated ? (
                  <div className="empty-state">
                    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ opacity: 0.3 }}>
                      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                      <circle cx="12" cy="7" r="4" />
                    </svg>
                    <p>Sign in to view your submissions</p>
                  </div>
                ) : subsLoading ? (
                  <div className="subs-loading">
                    {[...Array(4)].map((_, i) => (
                      <div key={i} className="skeleton" style={{ height: 48, marginBottom: 4, borderRadius: 8 }} />
                    ))}
                  </div>
                ) : submissions.length === 0 ? (
                  <div className="empty-state">
                    <TriangleAlert size={36} style={{ opacity: 0.3 }} />
                    <p>No submissions yet</p>
                    <span className="empty-hint">Submit a solution to see results here</span>
                  </div>
                ) : (
                  <div className="submissions-list">
                    {submissions.map((sub, i) => {
                      const vi = getVerdictInfo(sub.verdict);
                      return (
                        <div 
                           key={i} 
                           className={`sub-row interactive ${i % 2 === 0 ? 'even' : ''}`}
                           onClick={() => setSelectedSubmission(sub)}
                        >
                          <div className="sub-verdict">
                            <span className="sub-verdict-dot" style={{ background: vi.color }} />
                            <span className="sub-verdict-text" style={{ color: vi.color }}>{vi.label}</span>
                          </div>
                          <div className="sub-meta">
                            <span className="sub-lang">{sub.language === 'python' ? 'Python' : sub.language === 'java' ? 'Java' : 'C++'}</span>
                            {sub.execution_time > 0 && <span className="sub-stat">{sub.execution_time}ms</span>}
                            {sub.memory_used > 0 && <span className="sub-stat">{sub.memory_used}KB</span>}
                          </div>
                          <span className="sub-date">{formatDateTime(sub.created_at)}</span>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            )}

            {/* Editorial Tab */}
            {activeTab === 'editorial' && (
              <div className="editorial-tab" style={{ padding: '24px' }}>
                {!isAuthenticated ? (
                  <div className="empty-state">
                    <p>Sign in to view the editorial</p>
                  </div>
                ) : editorialLoading ? (
                  <div className="spinner" />
                ) : editorialLocked ? (
                  <div className="locked-editorial" style={{ textAlign: 'center', padding: '40px 20px' }}>
                    <h3>Editorial Locked</h3>
                    <p style={{ color: 'var(--text-tertiary)', marginBottom: '20px' }}>
                      You haven't solved this problem yet. You can unlock the editorial using your coins.
                    </p>
                    <button 
                      className="btn-primary" 
                      onClick={handleUnlockEditorial}
                      style={{ padding: '8px 16px', background: 'var(--accent-primary)', color: 'var(--text-inverse)', borderRadius: '4px', border: 'none', cursor: 'pointer', fontWeight: 600 }}
                    >
                      Unlock for 10 🪙
                    </button>
                  </div>
                ) : (
                  <div className="markdown-body">
                    <ReactMarkdown 
                      remarkPlugins={[remarkMath]}
                      rehypePlugins={[rehypeKatex]}
                      components={{ code: CodeBlock }}
                    >
                      {editorial || 'No editorial content available.'}
                    </ReactMarkdown>
                  </div>
                )}
              </div>
            )}
          </div>
        </Panel>

        <ResizeHandle direction="horizontal" />

        {/* ─── Right Panel: Code Editor ─── */}
        <Panel defaultSize={55} minSize={30} className="code-editor-panel">
          <PanelGroup orientation="vertical" className="vertical-split">
            <Panel defaultSize={70} minSize={30}>
              <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
              <div className="editor-header">
                <div className="editor-left-controls">
                  <div className="lang-selector">
                    <span className="lang-dot" />
                    <select 
                      className="lang-select" 
                      value={language} 
                      onChange={handleLanguageChange}
                      style={{ background: 'transparent', border: 'none', color: 'inherit', outline: 'none', cursor: 'pointer', fontSize: '14px', fontWeight: '500' }}
                    >
                      <option value="cpp" style={{ color: 'black' }}>C++</option>
                      <option value="python" style={{ color: 'black' }}>Python 3</option>
                      <option value="java" style={{ color: 'black' }}>Java</option>
                    </select>
                  </div>
                  <div className="editor-toolbar">
                    <button className="toolbar-btn" onClick={handleReset} title="Reset to template">
                      <RotateCcw size={16} />
                    </button>
                    <button className="toolbar-btn" onClick={handleRetrieveLastSubmission} title="Load last submitted code">
                      <Code2 size={16} />
                    </button>
                    <button 
                      className="toolbar-btn" 
                      onClick={() => setEditorSettings(s => ({ ...s, minimap: !s.minimap }))} 
                      title="Toggle Minimap"
                    >
                      <Settings size={16} />
                    </button>
                  </div>
                </div>
                <div className="editor-actions">
                  <button
                    className="btn-hint"
                    onClick={handleGetHint}
                    disabled={aiLoading}
                    title="Get a hint (Costs 5 tokens)"
                  >
                    {aiLoading ? <span className="spinner" /> : "💡 Hint (5 🪙)"}
                  </button>
                  {verdict && <VerdictBadge verdict={verdict.verdict} />}
                  <button
                    className="btn-run"
                    onClick={handleRunCode}
                    disabled={running || submitting}
                    title="Run code against sample test case"
                    style={{
                      display: 'flex', alignItems: 'center', gap: '7px', padding: '8px 16px',
                      background: 'var(--bg-tertiary)', color: 'var(--text-primary)', border: '1px solid var(--border-default)',
                      fontSize: '0.8rem', fontWeight: '600', borderRadius: 'var(--radius-md)', cursor: 'pointer',
                      transition: 'all var(--duration-fast)'
                    }}
                    onMouseEnter={(e) => !running && !submitting && (e.currentTarget.style.background = 'var(--bg-card-hover)')}
                    onMouseLeave={(e) => !running && !submitting && (e.currentTarget.style.background = 'var(--bg-tertiary)')}
                  >
                    {running ? <span className="spinner" /> : <><TerminalSquare size={16} /> Run</>}
                  </button>
                  <button
                    className="btn-submit"
                    onClick={handleSubmit}
                    disabled={submitting}
                    id="submit-btn"
                    style={{ position: 'relative' }}
                  >
                    {submitting ? (
                      <>
                        <span className="spinner" />
                        Judging…
                      </>
                    ) : (
                      <>
                        <Play size={16} fill="currentColor" />
                        Submit
                      </>
                    )}
                    {showPointsAnimation && (
                      <span className="floating-points">+2 Points!</span>
                    )}
                  </button>
                </div>
              </div>

              <div className="editor-body" style={{ flex: 1 }}>
                <CodeEditor 
                  value={code} 
                  onChange={setCode} 
                  language={language === 'cpp' ? 'cpp' : language === 'python' ? 'python' : 'java'} 
                  options={{
                    minimap: { enabled: editorSettings.minimap },
                    wordWrap: editorSettings.wordWrap,
                  }}
                />
              </div>
              </div>
            </Panel>

            <ResizeHandle direction="vertical" />

            {/* ─── Integrated Console Panel ─── */}
            <Panel defaultSize={30} minSize={10} className="console-panel">
              <div className="console-header">
                <div className="console-tabs">
                  <button 
                    className={`console-tab ${consoleTab === 'testcases' ? 'active' : ''}`}
                    onClick={() => setConsoleTab('testcases')}
                  >
                    <TerminalSquare size={14} /> Test Cases
                  </button>
                  <button 
                    className={`console-tab ${consoleTab === 'runresult' ? 'active' : ''}`}
                    onClick={() => setConsoleTab('runresult')}
                  >
                    <Check size={14} /> Test Result
                  </button>
                </div>
              </div>
              <div className="console-body">
                {consoleTab === 'testcases' && (
                  <div className="console-testcases">
                    {problem.sample_cases?.length > 0 ? (
                      <div className="testcase-item">
                        <div className="io-col">
                          <h5>Input</h5>
                          <pre>{problem.sample_cases[0].input}</pre>
                        </div>
                        <div className="io-col mt-4">
                          <h5>Expected Output</h5>
                          <pre>{problem.sample_cases[0].output}</pre>
                        </div>
                      </div>
                    ) : (
                      <div className="console-empty">No test cases available</div>
                    )}
                  </div>
                )}
                
                {consoleTab === 'runresult' && (
                  <div className="console-runresult">
                    {!runResult ? (
                      <div className="console-empty">
                        <p>Run your code to see results here</p>
                      </div>
                    ) : (
                      <div className="run-result-content">
                        <div className={`run-status ${runResult.status === 'Accepted' ? 'accepted' : 'rejected'}`}>
                          {runResult.status}
                        </div>
                        
                        {runResult.error ? (
                          <div className="run-error-box">
                            <div className="run-error-header">
                              <TriangleAlert size={16} /> Runtime/Compilation Error
                            </div>
                            <pre className="run-error-trace">{runResult.error}</pre>
                          </div>
                        ) : (
                          <div className="run-io-grid">
                            <div className="io-col">
                              <h5>Input</h5>
                              <pre>{runResult.input}</pre>
                            </div>
                            <div className="io-col">
                              <h5>Your Output</h5>
                              <pre className={runResult.status === 'Accepted' ? 'text-success' : 'text-danger'}>{runResult.actual || '(Empty)'}</pre>
                            </div>
                            <div className="io-col">
                              <h5>Expected Output</h5>
                              <pre>{runResult.expected}</pre>
                            </div>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </Panel>
          </PanelGroup>
        </Panel>
      </PanelGroup>
      </div>

      {/* ─── Submission Details Modal ─── */}
      {selectedSubmission && (
        <div className="submission-modal-overlay fadeIn" onClick={() => setSelectedSubmission(null)}>
          <div className="submission-modal slideUp" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div className="modal-title">
                <h3>Submission Details</h3>
                <span className="modal-date">{formatDateTime(selectedSubmission.created_at)}</span>
              </div>
              <button className="modal-close" onClick={() => setSelectedSubmission(null)}>
                <X size={20} />
              </button>
            </div>
            
            <div className={`modal-verdict-banner ${selectedSubmission.verdict === 'Accepted' ? 'accepted' : 'rejected'}`}>
              <div className="banner-left">
                <VerdictBadge verdict={selectedSubmission.verdict} />
                {selectedSubmission.message && selectedSubmission.message !== 'NA' && (
                  <span className="banner-message">{selectedSubmission.message}</span>
                )}
              </div>
              <div className="banner-stats">
                {selectedSubmission.execution_time > 0 && (
                  <div className="banner-stat">
                    <span>Runtime</span>
                    <strong>{selectedSubmission.execution_time}ms</strong>
                  </div>
                )}
                {selectedSubmission.memory_used > 0 && (
                  <div className="banner-stat">
                    <span>Memory</span>
                    <strong>{selectedSubmission.memory_used}KB</strong>
                  </div>
                )}
              </div>
            </div>

            {aiFeedback && (
              <div className="ai-feedback-box">
                <h4>🧠 AI Feedback</h4>
                {renderAIResponse(aiFeedback)}
              </div>
            )}
            
            <div className="modal-actions" style={{ padding: '0 24px 16px', display: 'flex', justifyContent: 'flex-end' }}>
              <button
                className="btn-ai-feedback"
                onClick={handleGetFeedback}
                disabled={aiFeedbackLoading}
                title="Get AI Optimization Feedback (Costs 3 tokens)"
              >
                {aiFeedbackLoading ? <span className="spinner" style={{ width: 14, height: 14 }} /> : "🧠 Get AI Feedback (3 🪙)"}
              </button>
            </div>

            <div className="modal-code-section">
              <div className="modal-code-header">
                <span className="code-lang-label">
                  <span className="lang-dot" />
                  C++
                </span>
                <button 
                  className="code-copy-btn"
                  onClick={() => navigator.clipboard.writeText(selectedSubmission.code)}
                  title="Copy Code"
                >
                  <Copy size={16} />
                  Copy
                </button>
              </div>
              <div className="modal-code-body" style={{ height: '350px' }}>
                <CodeEditor 
                  value={selectedSubmission.code || ''} 
                  language="cpp"
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
        </div>
      )}

      {/* ─── AI Hint Modal ─── */}
      {aiHint && (
        <div className="submission-modal-overlay fadeIn" onClick={() => setAiHint('')}>
          <div className="submission-modal slideUp" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '600px' }}>
            <div className="modal-header">
              <div className="modal-title">
                <h3>💡 AI Hint</h3>
              </div>
              <button className="modal-close" onClick={() => setAiHint('')}>
                <X size={20} />
              </button>
            </div>
            
            <div className="ai-feedback-box" style={{ margin: '20px', padding: '24px', fontSize: '15px' }}>
              {renderAIResponse(aiHint)}
            </div>
            
            <div className="modal-actions" style={{ padding: '0 24px 20px', display: 'flex', justifyContent: 'flex-end' }}>
              <button
                className="btn-submit"
                onClick={() => setAiHint('')}
                style={{ padding: '8px 24px' }}
              >
                Got it
              </button>
            </div>
          </div>
        </div>
      )}

    </>
  );
}
