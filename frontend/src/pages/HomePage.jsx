import { useState, useEffect, useMemo } from 'react';
import { fetchProblems, getUserSubmissions } from '../api/client';
import { useAuth } from '../context/AuthContext';
import ProblemCard from '../components/ProblemCard';
import './HomePage.css';

const ITEMS_PER_PAGE = 10;

export default function HomePage() {
  const { user, isAuthenticated } = useAuth();
  const [problems, setProblems] = useState([]);
  const [submissions, setSubmissions] = useState([]);
  const [search, setSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [difficultyFilter, setDifficultyFilter] = useState('All');
  const [topicFilter, setTopicFilter] = useState('All');

  const ALL_TOPICS = [
    'All', 'Math', 'Implementation', 'Greedy', 'Sorting',
    'Binary Search', 'Two Pointers', 'Strings', 'DP',
    'Graphs', 'Trees', 'DFS', 'BFS', 'Brute Force',
    'Number Theory', 'Constructive', 'Combinatorics',
    'Data Structures', 'Bitmasks', 'Geometry'
  ];

  useEffect(() => {
    async function loadData() {
      try {
        const probs = await fetchProblems();
        setProblems(probs || []);

        if (isAuthenticated && user?.user_id) {
          const subs = await getUserSubmissions(user.user_id);
          setSubmissions(subs || []);
        }
      } catch (err) {
        console.error('Failed to load data:', err);
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, [isAuthenticated, user]);

  // Compute problem statuses from submissions
  const problemStatuses = useMemo(() => {
    const statuses = {};
    for (const sub of submissions) {
      if (sub.verdict === 'Accepted') {
        statuses[sub.problem_id] = 'solved';
      } else if (!statuses[sub.problem_id]) {
        statuses[sub.problem_id] = 'attempted';
      }
    }
    return statuses;
  }, [submissions]);

  // Filter and search
  const filteredProblems = useMemo(() => {
    return problems.filter((p) => {
      const matchesSearch = p.title?.toLowerCase().includes(search.toLowerCase()) ||
                            p.id?.toLowerCase().includes(search.toLowerCase());
      const matchesDifficulty = difficultyFilter === 'All' || p.difficulty === difficultyFilter;
      const matchesTopic = topicFilter === 'All' || (p.tags && p.tags.some(t => t.toLowerCase() === topicFilter.toLowerCase()));
      return matchesSearch && matchesDifficulty && matchesTopic;
    });
  }, [problems, search, difficultyFilter, topicFilter]);

  // Pagination
  const totalPages = Math.ceil(filteredProblems.length / ITEMS_PER_PAGE);
  const paginatedProblems = filteredProblems.slice(
    (currentPage - 1) * ITEMS_PER_PAGE,
    currentPage * ITEMS_PER_PAGE
  );

  useEffect(() => {
    setCurrentPage(1);
  }, [search, difficultyFilter, topicFilter]);

  if (loading) {
    return (
      <div className="page-content">
        <div className="container">
          <div className="home-loading">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="skeleton problem-skeleton" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="page-content fade-in">
      <div className="home-container">
        <div className="home-hero">
          <h1><span>Problem Set</span></h1>
          <p className="home-subtitle">Master algorithms and data structures through curated competitive programming challenges</p>
        </div>

        <div className="home-controls">
          <div className="search-wrapper" id="search-bar">
            <svg className="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="11" cy="11" r="8" />
              <path d="M21 21l-4.35-4.35" />
            </svg>
            <input
              type="text"
              placeholder="Search problems..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="search-input"
            />
          </div>

          <div className="difficulty-filters">
            {['All', 'Easy', 'Medium', 'Hard'].map((d) => (
              <button
                key={d}
                className={`filter-btn ${difficultyFilter === d ? 'active' : ''}`}
                onClick={() => setDifficultyFilter(d)}
                data-difficulty={d.toLowerCase()}
              >
                {d}
              </button>
            ))}
          </div>

          <div className="topic-filters" id="topic-filters">
            {ALL_TOPICS.map((t) => (
              <button
                key={t}
                className={`topic-pill ${topicFilter === t ? 'active' : ''}`}
                onClick={() => setTopicFilter(t)}
              >
                {t}
              </button>
            ))}
          </div>
        </div>

        <div className="problems-table" id="problems-list">
          <div className="table-header">
            <div className="th-status">Status</div>
            <div className="th-title">Title</div>
            <div className="th-difficulty">Difficulty</div>
          </div>

          {paginatedProblems.length > 0 ? (
            paginatedProblems.map((problem, index) => (
              <ProblemCard
                key={problem.id}
                problem={problem}
                status={problemStatuses[problem.id] || 'unattempted'}
                index={index}
              />
            ))
          ) : (
            <div className="no-results">
              <span className="no-results-icon">🔍</span>
              <p>No problems found</p>
            </div>
          )}
        </div>

        {totalPages > 1 && (
          <div className="pagination" id="pagination">
            <button
              className="page-btn"
              disabled={currentPage === 1}
              onClick={() => setCurrentPage(currentPage - 1)}
            >
              ← Prev
            </button>

            <div className="page-numbers">
              {[...Array(totalPages)].map((_, i) => (
                <button
                  key={i}
                  className={`page-num ${currentPage === i + 1 ? 'active' : ''}`}
                  onClick={() => setCurrentPage(i + 1)}
                >
                  {i + 1}
                </button>
              ))}
            </div>

            <button
              className="page-btn"
              disabled={currentPage === totalPages}
              onClick={() => setCurrentPage(currentPage + 1)}
            >
              Next →
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
