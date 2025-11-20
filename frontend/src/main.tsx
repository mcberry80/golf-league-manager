import React from 'react'
import ReactDOM from 'react-dom/client'
import { ClerkProvider } from '@clerk/clerk-react'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import App from './App.tsx'
import './index.css'

// Import pages (we will create these next)
import Home from './pages/Home.tsx'
// import Admin from './pages/Admin.tsx'
// import LinkAccount from './pages/LinkAccount.tsx'
// import Players from './pages/Players.tsx'
// import Standings from './pages/Standings.tsx'

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
            // { path: "/admin", element: <Admin /> },
            // { path: "/link-account", element: <LinkAccount /> },
            // { path: "/players", element: <Players /> },
            // { path: "/standings", element: <Standings /> },
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
