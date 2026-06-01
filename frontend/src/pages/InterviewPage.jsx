import { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import {
  startInterview,
  sendInterviewResponse,
  endInterview,
  getInterviews,
} from '../api/client';
import { Mic, Square, Send, Clock, Zap, Brain, Target } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import Editor from '@monaco-editor/react';
import './InterviewPage.css';

export default function InterviewPage() {
  const { user, isAuthenticated, loading: authLoading, refreshUser } = useAuth();
  const navigate = useNavigate();

  // Setup state
  const [level, setLevel] = useState('');
  const [duration, setDuration] = useState(0);

  // Active interview state
  const [interviewId, setInterviewId] = useState(null);
  const [chatHistory, setChatHistory] = useState([]);
  const [code, setCode] = useState('// Write your code here...\n');
  const [isStarting, setIsStarting] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const [interviewStatus, setInterviewStatus] = useState('setup'); // setup | active | completed
  const [feedback, setFeedback] = useState('');

  // Timer state
  const [timeLeft, setTimeLeft] = useState(0);
  const timerRef = useRef(null);

  // Audio recording
  const [isRecording, setIsRecording] = useState(false);
  const [audioBlob, setAudioBlob] = useState(null);
  const mediaRecorderRef = useRef(null);
  const audioChunksRef = useRef([]);
  const [recordingTime, setRecordingTime] = useState(0);
  const recordingTimerRef = useRef(null);
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);

  // Idle tracking
  const [silentStrikes, setSilentStrikes] = useState(0);
  const isNudgedRef = useRef(false);
  const idleTimerRef = useRef(null);

  // Past interviews
  const [pastInterviews, setPastInterviews] = useState([]);

  const chatEndRef = useRef(null);

  // ─── End Interview ──────────────────────────────────
  const handleEndInterview = useCallback(async () => {
    if (!interviewId) return;
    clearInterval(timerRef.current);
    setIsProcessing(true);

    try {
      const data = await endInterview(interviewId);
      setFeedback(data.feedback || 'Interview completed.');
      setInterviewStatus('completed');
      refreshUser();

      // Refresh past interviews
      getInterviews()
        .then(setPastInterviews)
        .catch(() => {});
    } catch (err) {
      console.error('Failed to end interview:', err);
      setInterviewStatus('completed');
      setFeedback('Interview ended. Feedback could not be generated.');
    } finally {
      setIsProcessing(false);
    }
  }, [interviewId, refreshUser]);

  // Load past interviews
  useEffect(() => {
    if (isAuthenticated) {
      getInterviews()
        .then(setPastInterviews)
        .catch(() => setPastInterviews([]));
    }
  }, [isAuthenticated]);

  // Auto-scroll chat
  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatHistory, isProcessing]);

  // Idle Nudge Logic
  const handleSystemAction = useCallback(async (actionType) => {
    setIsProcessing(true);
    try {
      const data = await sendInterviewResponse(interviewId, "", code, false, actionType);
      const newEntries = [{ role: 'interviewer', text: data.question_text }];
      setChatHistory(prev => [...prev, ...newEntries]);
      if (data.audio_base64) {
        playAudioBase64(data.audio_base64);
      }
    } catch (err) {
      console.error('System action failed:', err);
    } finally {
      setIsProcessing(false);
    }
  }, [interviewId]);

  const resetIdleTimer = useCallback(() => {
    if (idleTimerRef.current) clearTimeout(idleTimerRef.current);
    if (interviewStatus !== 'active' || isRecording || isProcessing || isPlayingAudio || timeLeft === 0 || audioBlob) return;

    idleTimerRef.current = setTimeout(async () => {
      if (isNudgedRef.current) {
        const newStrikes = silentStrikes + 1;
        setSilentStrikes(newStrikes);
        if (newStrikes >= 3) {
          handleEndInterview();
        } else {
          isNudgedRef.current = false;
          await handleSystemAction('idle_skip');
        }
      } else {
        isNudgedRef.current = true;
        await handleSystemAction('idle_nudge');
      }
    }, 60000); // 1 minute
  }, [interviewStatus, isRecording, isProcessing, isPlayingAudio, timeLeft, audioBlob, silentStrikes, handleEndInterview, handleSystemAction]);

  useEffect(() => {
    resetIdleTimer();
    return () => { if (idleTimerRef.current) clearTimeout(idleTimerRef.current); };
  }, [resetIdleTimer]);

  // Timer countdown
  useEffect(() => {
    if (interviewStatus === 'active' && timeLeft > 0) {
      timerRef.current = setInterval(() => {
        setTimeLeft(prev => {
          if (prev <= 1) {
            clearInterval(timerRef.current);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
      return () => clearInterval(timerRef.current);
    }
  }, [interviewStatus, timeLeft]);

  // Auto-end if time is up and not doing anything
  useEffect(() => {
    if (interviewStatus === 'active' && timeLeft === 0 && !isRecording && !audioBlob && !isProcessing && !isPlayingAudio) {
      // Add a small delay to prevent race condition with mediaRecorder.onstop which fires asynchronously after stop()
      const timeout = setTimeout(() => {
        handleEndInterview();
      }, 500);
      return () => clearTimeout(timeout);
    }
  }, [timeLeft, isRecording, audioBlob, isProcessing, isPlayingAudio, interviewStatus, handleEndInterview]);

  const formatTime = (seconds) => {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  // ─── Start Interview ────────────────────────────────
  const handleStartInterview = async () => {
    if (!level || !duration) return;
    setIsStarting(true);
    try {
      const data = await startInterview(level, duration);
      setInterviewId(data.interview_id);
      setChatHistory([{ role: 'interviewer', text: data.question_text }]);
      setTimeLeft(duration * 60);
      setInterviewStatus('active');
      refreshUser();

      // Play AI audio
      if (data.audio_base64) {
        playAudioBase64(data.audio_base64);
      }
    } catch (err) {
      console.error('Failed to start interview:', err);
      const msg = err.response?.data?.error || 'Failed to start interview';
      alert(msg);
    } finally {
      setIsStarting(false);
    }
  };

  // ─── Recording ──────────────────────────────────────
  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const mediaRecorder = new MediaRecorder(stream, { mimeType: 'audio/webm' });
      mediaRecorderRef.current = mediaRecorder;
      audioChunksRef.current = [];

      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) audioChunksRef.current.push(e.data);
      };

      mediaRecorder.onstop = () => {
        const blob = new Blob(audioChunksRef.current, { type: 'audio/webm' });
        setAudioBlob(blob);
        stream.getTracks().forEach(t => t.stop());
      };

      mediaRecorder.start();
      setIsRecording(true);
      setRecordingTime(0);

      const maxRecTime = duration === 5 ? 180 : duration === 10 ? 240 : 300;

      recordingTimerRef.current = setInterval(() => {
        setRecordingTime(prev => {
          if (prev >= maxRecTime - 1) { 
            stopRecording();
            return maxRecTime;
          }
          return prev + 1;
        });
      }, 1000);
    } catch (err) {
      console.error('Microphone access denied:', err);
      alert('Please allow microphone access to participate in the interview.');
    }
  };

  const stopRecording = useCallback(() => {
    if (mediaRecorderRef.current && mediaRecorderRef.current.state === 'recording') {
      mediaRecorderRef.current.stop();
      setIsRecording(false);
      if (recordingTimerRef.current) {
        clearInterval(recordingTimerRef.current);
      }
    }
  }, []);

  // ─── Send Audio Response ─────────────────────────────
  const handleSendResponse = async () => {
    if (!audioBlob || isProcessing) return;
    setIsProcessing(true);
    isNudgedRef.current = false; // Reset nudge state on successful answer

    try {
      // Convert blob to base64
      const reader = new FileReader();
      const base64Promise = new Promise((resolve) => {
        reader.onloadend = () => {
          const base64 = reader.result.split(',')[1];
          resolve(base64);
        };
        reader.readAsDataURL(audioBlob);
      });
      const audioBase64 = await base64Promise;

      const isTimeUp = timeLeft <= 60; // Conclude if less than a minute left
      const data = await sendInterviewResponse(interviewId, audioBase64, code, isTimeUp);

      // Update chat
      const newEntries = [];
      if (data.user_transcript) {
        newEntries.push({ role: 'candidate', text: data.user_transcript });
      }
      newEntries.push({ role: 'interviewer', text: data.question_text });
      setChatHistory(prev => [...prev, ...newEntries]);

      setAudioBlob(null);

      // Play AI audio
      if (data.audio_base64) {
        playAudioBase64(data.audio_base64, () => {
          if (isTimeUp) handleEndInterview();
        });
      } else {
        if (isTimeUp) handleEndInterview();
      }
    } catch (err) {
      console.error('Failed to process response:', err);
      alert('Failed to process your response. Please try again.');
      if (timeLeft === 0) handleEndInterview();
    } finally {
      setIsProcessing(false);
    }
  };


  // ─── Audio Playback ──────────────────────────────────
  const playAudioBase64 = (base64Data, onEnded) => {
    try {
      const byteChars = atob(base64Data);
      const byteArray = new Uint8Array(byteChars.length);
      for (let i = 0; i < byteChars.length; i++) {
        byteArray[i] = byteChars.charCodeAt(i);
      }
      const blob = new Blob([byteArray], { type: 'audio/wav' });
      const url = URL.createObjectURL(blob);
      const audio = new Audio(url);
      
      setIsPlayingAudio(true);
      audio.onended = () => {
        setIsPlayingAudio(false);
        if (onEnded) onEnded();
      };
      
      audio.play().catch(() => {
        setIsPlayingAudio(false);
        if (onEnded) onEnded();
      });
    } catch (e) {
      console.error('Audio playback error:', e);
      setIsPlayingAudio(false);
      if (onEnded) onEnded();
    }
  };

  // ─── Auth guard ──────────────────────────────────────
  if (authLoading) return null;
  if (!isAuthenticated) {
    return (
      <div className="page-content fade-in">
        <div className="container" style={{ textAlign: 'center', padding: '80px 0' }}>
          <h2 style={{ color: 'var(--text-primary)' }}>Sign in to access AI Interviews</h2>
          <p style={{ color: 'var(--text-muted)', marginTop: 8 }}>
            Practice your coding interview skills with our AI interviewer.
          </p>
        </div>
      </div>
    );
  }

  // ─── SETUP VIEW ──────────────────────────────────────
  if (interviewStatus === 'setup') {
    return (
      <div className="page-content fade-in">
        <div className="interview-container">
          <div className="interview-hero">
            <h1><span>AI Mock Interview</span></h1>
            <p>Practice with an AI interviewer. Choose your difficulty and duration to begin.</p>
          </div>

          {/* Level Selection */}
          <div className="setup-section">
            <h2>Select Difficulty</h2>
            <div className="level-grid">
              {[
                { key: 'easy', icon: <Target size={28} />, desc: 'Arrays, Strings, Sorting, Basics' },
                { key: 'medium', icon: <Zap size={28} />, desc: 'DP, Graphs, Binary Search, Greedy' },
                { key: 'hard', icon: <Brain size={28} />, desc: 'Segment Trees, Flows, Advanced DP' },
              ].map(item => (
                <div
                  key={item.key}
                  className={`level-card ${item.key} ${level === item.key ? 'selected' : ''}`}
                  onClick={() => setLevel(item.key)}
                >
                  <div className="level-icon">{item.icon}</div>
                  <div className="level-label">{item.key}</div>
                  <div className="level-desc">{item.desc}</div>
                </div>
              ))}
            </div>
          </div>

          {/* Duration Selection */}
          <div className="setup-section">
            <h2>Select Duration</h2>
            <div className="duration-grid">
              <div
                className={`duration-card ${duration === 5 ? 'selected' : ''}`}
                onClick={() => setDuration(5)}
              >
                <div className="duration-label">5 Minutes</div>
                <div className="duration-cost"><span>25</span> tokens</div>
              </div>
              <div
                className={`duration-card ${duration === 10 ? 'selected' : ''}`}
                onClick={() => setDuration(10)}
              >
                <div className="duration-label">10 Minutes</div>
                <div className="duration-cost"><span>40</span> tokens</div>
              </div>
              <div
                className={`duration-card ${duration === 15 ? 'selected' : ''}`}
                onClick={() => setDuration(15)}
              >
                <div className="duration-label">15 Minutes</div>
                <div className="duration-cost"><span>50</span> tokens</div>
              </div>
              <div
                className={`duration-card ${duration === 30 ? 'selected' : ''}`}
                onClick={() => setDuration(30)}
              >
                <div className="duration-label">30 Minutes</div>
                <div className="duration-cost"><span>70</span> tokens</div>
              </div>
            </div>
          </div>

          {/* Start Button */}
          <button
            className="start-interview-btn"
            disabled={!level || !duration || isStarting}
            onClick={handleStartInterview}
          >
            {isStarting ? 'Starting Interview...' : (!level || !duration) ? 'Select options to begin...' : `Start ${level} Interview (${duration} min)`}
          </button>

          {/* Past Interviews */}
          {pastInterviews.length > 0 && (
            <div className="past-interviews">
              <h2>Past Interviews</h2>
              <div className="past-interviews-list">
                {pastInterviews.map(iv => (
                  <div
                    key={iv.id}
                    className="past-interview-card"
                    onClick={() => navigate(`/interview/${iv.id}`)}
                  >
                    <div className="past-interview-info">
                      <span className={`past-interview-level ${iv.level}`}>{iv.level}</span>
                      <span className="past-interview-meta">{iv.duration} min • {iv.tokens_deducted} tokens</span>
                    </div>
                    <span className={`past-interview-status ${iv.status}`}>{iv.status}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    );
  }

  // ─── ACTIVE INTERVIEW VIEW ───────────────────────────
  if (interviewStatus === 'active') {
    return (
      <div className="page-content fade-in" style={{ maxWidth: '100vw', paddingLeft: '20px', paddingRight: '20px' }}>
        <div className="interview-container" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px', maxWidth: '1400px', margin: '0 auto' }}>
          <div className="interview-active" style={{ height: 'calc(100vh - 120px)', display: 'flex', flexDirection: 'column' }}>
            {/* Timer Bar */}
            <div className="interview-timer-bar">
              <div>
                <div className={`timer-display ${timeLeft < 60 ? 'warning' : ''}`}>
                  <Clock size={20} />
                  {formatTime(timeLeft)}
                </div>
                <div className="timer-label">Time Remaining</div>
              </div>
              <button className="end-btn" onClick={handleEndInterview}>
                End Interview
              </button>
            </div>

            {/* Chat Area */}
            <div className="interview-chat">
              {chatHistory.map((msg, i) => (
                <div key={i} className={`chat-bubble ${msg.role}`}>
                  <div className="chat-role">
                    {msg.role === 'interviewer' ? '🤖 Interviewer' : '👤 You'}
                  </div>
                  {msg.text}
                </div>
              ))}
              {isProcessing && (
                <div className="thinking-indicator">
                  <div className="thinking-dots">
                    <span /><span /><span />
                  </div>
                  <span className="thinking-text">Thinking...</span>
                </div>
              )}
              <div ref={chatEndRef} />
            </div>

            {/* Audio Controls */}
            <div className="audio-controls">
              <button
                className={`record-btn ${isRecording ? 'recording' : ''}`}
                onClick={isRecording ? stopRecording : startRecording}
                disabled={isProcessing || isPlayingAudio || (timeLeft === 0 && !isRecording)}
                title={timeLeft === 0 && !isRecording ? 'Time is up' : ''}
              >
                {isRecording ? <div className="record-icon" /> : <Mic size={24} />}
              </button>

              <div className="audio-status">
                {isRecording ? (
                  <>
                    <div className="audio-waveform">
                      <div className="wave-bar" />
                      <div className="wave-bar" />
                      <div className="wave-bar" />
                      <div className="wave-bar" />
                      <div className="wave-bar" />
                    </div>
                    <span className="audio-status-text" style={{ color: '#e74c3c' }}>
                      Recording... {formatTime(recordingTime)} / {duration === 5 ? '03:00' : duration === 10 ? '04:00' : '05:00'}
                    </span>
                  </>
                ) : audioBlob ? (
                  <>
                    <span className="audio-status-text">Recording ready</span>
                    <span className="audio-status-hint">Click send to submit your answer</span>
                  </>
                ) : (
                  <>
                    <span className="audio-status-text">Click the mic to record your answer</span>
                    <span className="audio-status-hint">Speak clearly into your microphone</span>
                  </>
                )}
              </div>

              <button
                className="send-btn"
                disabled={!audioBlob || isProcessing}
                onClick={handleSendResponse}
              >
                <Send size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
                {isProcessing ? 'Processing...' : 'Send'}
              </button>
            </div>
          </div>
          <div className="interview-editor-container" style={{ display: 'flex', flexDirection: 'column', borderRadius: '12px', overflow: 'hidden', border: '1px solid var(--border-color)', height: 'calc(100vh - 120px)' }}>
            <div style={{ padding: '12px 16px', background: 'var(--bg-secondary)', borderBottom: '1px solid var(--border-color)', fontWeight: 600, color: 'var(--text-primary)' }}>
              Interview Code Editor
            </div>
            <Editor
              height="100%"
              defaultLanguage="cpp"
              theme="vs-dark"
              value={code}
              onChange={(value) => {
                setCode(value);
                if (!isRecording && !isStarting && interviewStatus === 'active') {
                  startRecording().catch(console.error);
                }
              }}
              options={{
                minimap: { enabled: false },
                fontSize: 14,
                wordWrap: 'on'
              }}
            />
          </div>
        </div>
      </div>
    );
  }

  // ─── COMPLETED VIEW ──────────────────────────────────
  return (
    <div className="page-content fade-in">
      <div className="interview-container">
        <div className="interview-hero">
          <h1><span>Interview Complete</span></h1>
          <p>Here's your performance feedback from the AI interviewer.</p>
        </div>

        {/* Chat transcript */}
        <div className="interview-chat" style={{ marginBottom: 20 }}>
          {chatHistory.map((msg, i) => (
            <div key={i} className={`chat-bubble ${msg.role}`}>
              <div className="chat-role">
                {msg.role === 'interviewer' ? '🤖 Interviewer' : '👤 You'}
              </div>
              {msg.text}
            </div>
          ))}
        </div>

        {/* Feedback */}
        <div className="interview-feedback">
          <h2>📋 Performance Feedback</h2>
          <div className="feedback-content">
    <ReactMarkdown>
        {feedback}
    </ReactMarkdown>
</div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', marginTop: 20 }}>
          <button
            className="start-interview-btn"
            onClick={() => {
              setInterviewStatus('setup');
              setChatHistory([]);
              setFeedback('');
              setInterviewId(null);
              setLevel('');
              setDuration(0);
              setAudioBlob(null);
            }}
          >
            Start Another Interview
          </button>
          
          <button
            onClick={() => navigate('/')}
            style={{
              padding: '14px',
              border: '1px solid var(--border-default)',
              borderRadius: 'var(--radius-lg)',
              background: 'transparent',
              color: 'var(--text-primary)',
              fontSize: '1rem',
              fontWeight: '600',
              cursor: 'pointer',
              transition: 'all 0.2s ease',
            }}
            onMouseOver={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.05)';
              e.currentTarget.style.borderColor = 'var(--border-hover)';
            }}
            onMouseOut={(e) => {
              e.currentTarget.style.background = 'transparent';
              e.currentTarget.style.borderColor = 'var(--border-default)';
            }}
          >
            Back to Home
          </button>
        </div>
      </div>
    </div>
  );
}
