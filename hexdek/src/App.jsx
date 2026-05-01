import { Routes, Route, Navigate } from 'react-router-dom'
import AppShell from './components/AppShell'
import { useAuth } from './context/AuthContext'
import { usePageTracking } from './hooks/useAnalytics'
import Splash from './screens/Splash'
import Dashboard from './screens/Dashboard'
import DeckArchive from './screens/DeckArchive'
import DeckList from './screens/DeckList'
import GameBoard from './screens/GameBoard'
import Spectator from './screens/Spectator'
import Report from './screens/Report'
import Forge from './screens/Forge'
import Leaderboard from './screens/Leaderboard'
import About from './screens/About'
import BugReport from './screens/BugReport'
import Donations from './screens/Donations'
import Profile from './screens/Profile'
import Login from './screens/Login'
import AuthCallback from './screens/AuthCallback'

function RequireAuth({ children }) {
  const { user, loading } = useAuth()
  if (loading) return null
  return user ? children : <Navigate to="/login" replace />
}

export default function App() {
  usePageTracking()
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<Splash />} />
        <Route path="login" element={<Login />} />
        <Route path="auth/callback" element={<AuthCallback />} />
        <Route path="dash" element={<Dashboard />} />
        <Route path="decks" element={<DeckList />} />
        <Route path="decks/:owner/:id" element={<DeckArchive />} />
        <Route path="play" element={<GameBoard />} />
        <Route path="forge" element={<Forge />} />
        <Route path="leaderboard" element={<Leaderboard />} />
        <Route path="spectate" element={<Spectator />} />
        <Route path="report" element={<Report />} />
        <Route path="report/:gameId" element={<Report />} />
        <Route path="about" element={<About />} />
        <Route path="feedback" element={<BugReport />} />
        <Route path="donations" element={<Donations />} />
        <Route path="profile" element={<Profile />} />
      </Route>
    </Routes>
  )
}
