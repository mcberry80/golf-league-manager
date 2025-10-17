import { SignInButton, SignedIn, SignedOut, UserButton } from '@clerk/nextjs'
import Link from 'next/link'

export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="z-10 max-w-5xl w-full items-center justify-between font-mono text-sm">
        <div className="fixed top-0 right-0 p-4">
          <SignedOut>
            <SignInButton mode="modal">
              <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
                Sign In
              </button>
            </SignInButton>
          </SignedOut>
          <SignedIn>
            <UserButton afterSignOutUrl="/" />
          </SignedIn>
        </div>

        <div className="text-center">
          <h1 className="text-6xl font-bold mb-8">
            Golf League Manager
          </h1>
          
          <p className="text-xl mb-12">
            Manage your golf league with handicaps and match play
          </p>

          <SignedIn>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-12">
              <Link 
                href="/standings"
                className="p-6 border border-gray-300 rounded-lg hover:border-blue-500 transition-colors"
              >
                <h2 className="text-2xl font-semibold mb-2">Standings</h2>
                <p className="text-gray-600">View league standings and player rankings</p>
              </Link>

              <Link 
                href="/players"
                className="p-6 border border-gray-300 rounded-lg hover:border-blue-500 transition-colors"
              >
                <h2 className="text-2xl font-semibold mb-2">My Profile</h2>
                <p className="text-gray-600">View your scores and handicap history</p>
              </Link>

              <Link 
                href="/admin"
                className="p-6 border border-gray-300 rounded-lg hover:border-blue-500 transition-colors"
              >
                <h2 className="text-2xl font-semibold mb-2">Admin</h2>
                <p className="text-gray-600">Manage league, players, and scores</p>
              </Link>
            </div>
          </SignedIn>

          <SignedOut>
            <p className="text-gray-600 mt-8">
              Please sign in to access the league management features
            </p>
          </SignedOut>
        </div>
      </div>
    </main>
  )
}
