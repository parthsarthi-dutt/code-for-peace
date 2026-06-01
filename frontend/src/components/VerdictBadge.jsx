import { getVerdictInfo } from '../utils/helpers';
import './VerdictBadge.css';

export default function VerdictBadge({ verdict, showIcon = true }) {
  const info = getVerdictInfo(verdict);

  return (
    <span
      className={`verdict-badge ${verdict === 'pending' ? 'pending' : ''}`}
      style={{ color: info.color, background: info.bg }}
    >
      {showIcon && <span className="verdict-icon">{info.icon}</span>}
      <span className="verdict-label">{info.label}</span>
    </span>
  );
}
