import React from 'react'
import ReactDOM from 'react-dom/client'
import { ClerkProvider } from '@clerk/clerk-react'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import App from './App.tsx'
import './index.css'

// Import pages
import Home from './pages/Home.tsx'
import Admin from './pages/Admin.tsx'
import LinkAccount from './pages/LinkAccount.tsx'
import Profile from './pages/Profile.tsx'
import Standings from './pages/Standings.tsx'
import Results from './pages/Results.tsx'

// Import admin pages
import LeagueSetup from './pages/admin/LeagueSetup.tsx'
import PlayerManagement from './pages/admin/PlayerManagement.tsx'
import CourseManagement from './pages/admin/CourseManagement.tsx'
import MatchScheduling from './pages/admin/MatchScheduling.tsx'
import ScoreEntry from './pages/admin/ScoreEntry.tsx'
import SeasonPlayerManagement from './pages/admin/SeasonPlayerManagement.tsx'

// Import league pages
import LeagueList from './pages/leagues/LeagueList.tsx'
import CreateLeague from './pages/leagues/CreateLeague.tsx'

const PUBLISHABLE_KEY = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY

if (!PUBLISHABLE_KEY) {
    throw new Error("Missing Publishable Key")
}

const router = createBrowserRouter([
    {
        path: "/",
        element: <App />,
        children: [
            { path: "/", element: <Home /> },
            { path: "/leagues", element: <LeagueList /> },
            { path: "/leagues/create", element: <CreateLeague /> },
            { path: "/leagues/:leagueId/admin", element: <Admin /> },
            { path: "/leagues/:leagueId/admin/league-setup", element: <LeagueSetup /> },
            { path: "/leagues/:leagueId/admin/players", element: <PlayerManagement /> },
            { path: "/leagues/:leagueId/admin/courses", element: <CourseManagement /> },
            { path: "/leagues/:leagueId/admin/matches", element: <MatchScheduling /> },
            { path: "/leagues/:leagueId/admin/scores", element: <ScoreEntry /> },
            { path: "/leagues/:leagueId/admin/seasons/:seasonId/players", element: <SeasonPlayerManagement /> },
            { path: "/link-account", element: <LinkAccount /> },
            { path: "/profile/:playerId", element: <Profile /> },
            { path: "/leagues/:leagueId/standings", element: <Standings /> },
            { path: "/leagues/:leagueId/results", element: <Results /> },
        ],
    },
]);

ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
        <ClerkProvider publishableKey={PUBLISHABLE_KEY}>
            <RouterProvider router={router} />
        </ClerkProvider>
    </React.StrictMode>,
)
