/**
 * Generate a unique submission ID using problem_id + user_id + timestamp + random
 */
export function generateSubmissionId(problemId, userId) {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 8);
  return `${problemId}-${userId}-${timestamp}-${random}`;
}

/**
 * Format a date string to a readable format
 */
export function formatDate(dateString) {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

/**
 * Format a date string to include time
 */
export function formatDateTime(dateString) {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/**
 * Get difficulty color CSS variable name
 */
export function getDifficultyColor(difficulty) {
  switch (difficulty?.toLowerCase()) {
    case 'easy': return 'var(--easy)';
    case 'medium': return 'var(--medium)';
    case 'hard': return 'var(--hard)';
    default: return 'var(--text-secondary)';
  }
}

/**
 * Get verdict display info
 */
export function getVerdictInfo(verdict) {
  switch (verdict) {
    case 'Accepted':
      return { label: 'Accepted', color: 'var(--success)', bg: 'var(--success-dim)', icon: '✓' };
    case 'Wrong Answer':
      return { label: 'Wrong Answer', color: 'var(--danger-text)', bg: 'var(--danger-dim)', icon: '✗' };
    case 'Time Limit Exceeded':
      return { label: 'TLE', color: 'var(--warning-text)', bg: 'var(--warning-dim)', icon: '⏱' };
    case 'Memory Limit Exceeded':
      return { label: 'MLE', color: 'var(--warning-text)', bg: 'var(--warning-dim)', icon: '💾' };
    case 'Runtime Error':
      return { label: 'Runtime Error', color: 'var(--danger-text)', bg: 'var(--danger-dim)', icon: '!' };
    case 'Compilation Error':
      return { label: 'CE', color: 'var(--warning-text)', bg: 'var(--warning-dim)', icon: '⚠' };
    case 'pending':
      return { label: 'Running…', color: 'var(--info-text)', bg: 'var(--info-dim)', icon: '⟳' };
    default:
      return { label: verdict || 'Unknown', color: 'var(--text-secondary)', bg: 'var(--bg-tertiary)', icon: '?' };
  }
}

/**
 * Get number of days between two dates
 */
export function daysBetween(date1, date2) {
  const d1 = new Date(date1);
  const d2 = new Date(date2);
  return Math.floor((d2 - d1) / (1000 * 60 * 60 * 24));
}
