import { Link } from 'react-router-dom';
import { getDifficultyColor } from '../utils/helpers';
import './ProblemCard.css';

export default function ProblemCard({ problem, status, index }) {
  const diffColor = getDifficultyColor(problem.difficulty);

  return (
    <Link
      to={`/problem/${problem.id}`}
      className={`problem-card ${index % 2 === 0 ? 'even' : 'odd'}`}
      id={`problem-${problem.id}`}
    >
      <div className="problem-status">
        {status === 'solved' && <span className="status-icon solved">✓</span>}
        {status === 'attempted' && <span className="status-icon attempted">✗</span>}
        {status === 'unattempted' && <span className="status-icon unattempted">—</span>}
      </div>

      <div className="problem-info">
        <span className="problem-title">{problem.title}</span>
        <div className="problem-tags">
          {problem.tags?.map((tag) => (
            <span key={tag} className="problem-tag">{tag}</span>
          ))}
        </div>
      </div>

      <div className="problem-meta">
        <span
          className="difficulty-badge"
          style={{ color: diffColor, background: `${diffColor}15` }}
        >
          {problem.difficulty}
        </span>
      </div>
    </Link>
  );
}
