'use client'

import { SignedIn } from '@clerk/nextjs'
import Link from 'next/link'

export default function AdminPage() {
  return (
    <SignedIn>
      <div className="container mx-auto p-6">
        <h1 className="text-4xl font-bold mb-8">League Administration</h1>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <Link 
            href="/admin/courses"
            className="p-6 bg-white rounded-lg shadow hover:shadow-lg transition-shadow"
          >
            <h2 className="text-2xl font-semibold mb-2">Courses</h2>
            <p className="text-gray-600">Add and manage golf courses</p>
          </Link>

          <Link 
            href="/admin/players"
            className="p-6 bg-white rounded-lg shadow hover:shadow-lg transition-shadow"
          >
            <h2 className="text-2xl font-semibold mb-2">Players</h2>
            <p className="text-gray-600">Manage player accounts</p>
          </Link>

          <Link 
            href="/admin/matches"
            className="p-6 bg-white rounded-lg shadow hover:shadow-lg transition-shadow"
          >
            <h2 className="text-2xl font-semibold mb-2">Matches</h2>
            <p className="text-gray-600">Schedule and manage matches</p>
          </Link>

          <Link 
            href="/admin/scores"
            className="p-6 bg-white rounded-lg shadow hover:shadow-lg transition-shadow"
          >
            <h2 className="text-2xl font-semibold mb-2">Enter Scores</h2>
            <p className="text-gray-600">Enter match and round scores</p>
          </Link>

          <Link 
            href="/admin/handicaps"
            className="p-6 bg-white rounded-lg shadow hover:shadow-lg transition-shadow"
          >
            <h2 className="text-2xl font-semibold mb-2">Handicaps</h2>
            <p className="text-gray-600">Recalculate player handicaps</p>
          </Link>
        </div>

        <div className="mt-8">
          <Link 
            href="/"
            className="text-blue-500 hover:text-blue-700"
          >
            ← Back to Home
          </Link>
        </div>
      </div>
    </SignedIn>
  )
}
