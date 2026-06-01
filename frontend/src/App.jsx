import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import Navbar from './components/Navbar';
import HomePage from './pages/HomePage';
import ProblemPage from './pages/ProblemPage';
import ProfilePage from './pages/ProfilePage';
import InterviewPage from './pages/InterviewPage';
import InterviewDetailPage from './pages/InterviewDetailPage';
import LoginCallback from './pages/LoginCallback';

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Navbar />
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/problem/:id" element={<ProblemPage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/interview" element={<InterviewPage />} />
          <Route path="/interview/:id" element={<InterviewDetailPage />} />
          <Route path="/auth/callback" element={<LoginCallback />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

