import { Outlet } from 'react-router-dom'
import { LeagueProvider } from './contexts/LeagueContext'

function App() {
    return (
        <LeagueProvider>
            <div className="min-h-screen">
                <Outlet />
            </div>
        </LeagueProvider>
    )
}

export default App
