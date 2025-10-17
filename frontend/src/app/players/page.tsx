'use client'

import { SignedIn, useUser } from '@clerk/nextjs'
import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import type { Round, HandicapRecord } from '@/types'
import Link from 'next/link'

export default function PlayersPage() {
  const { user } = useUser()
  const [rounds, setRounds] = useState<Round[]>([])
  const [handicap, setHandicap] = useState<HandicapRecord | null>(null)
  const [loading, setLoading] = useState(true)

  // In a real app, you'd map Clerk user ID to your player ID
  const playerId = user?.id || ''

  useEffect(() => {
    if (playerId) {
      loadPlayerData()
    }
  }, [playerId])

  const loadPlayerData = async () => {
    try {
      const [roundsData, handicapData] = await Promise.all([
        api.getPlayerRounds(playerId).catch(() => []),
        api.getPlayerHandicap(playerId).catch(() => null),
      ])
      setRounds(roundsData)
      setHandicap(handicapData)
    } catch (error) {
      console.error('Failed to load player data:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <SignedIn>
      <div className="container mx-auto p-6">
        <h1 className="text-4xl font-bold mb-8">My Profile</h1>
        
        {loading ? (
          <p>Loading your data...</p>
        ) : (
          <>
            {/* Handicap Card */}
            <div className="bg-white rounded-lg shadow p-6 mb-8">
              <h2 className="text-2xl font-semibold mb-4">Current Handicap</h2>
              {handicap ? (
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <p className="text-gray-600 text-sm">League Handicap</p>
                    <p className="text-3xl font-bold">{handicap.league_handicap.toFixed(1)}</p>
                  </div>
                  <div>
                    <p className="text-gray-600 text-sm">Course Handicap</p>
                    <p className="text-3xl font-bold">{handicap.course_handicap.toFixed(1)}</p>
                  </div>
                  <div>
                    <p className="text-gray-600 text-sm">Playing Handicap</p>
                    <p className="text-3xl font-bold">{handicap.playing_handicap}</p>
                  </div>
                </div>
              ) : (
                <p className="text-gray-600">No handicap data available yet. Play some rounds to establish your handicap!</p>
              )}
            </div>

            {/* Round History */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-2xl font-semibold mb-4">Round History</h2>
              {rounds.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="min-w-full">
                    <thead className="bg-gray-100">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Course</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Gross</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Adjusted</th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {rounds.map((round) => (
                        <tr key={round.id}>
                          <td className="px-6 py-4 whitespace-nowrap text-sm">
                            {new Date(round.date).toLocaleDateString()}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {round.course_id}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-semibold">
                            {round.total_gross}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {round.total_adjusted}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <p className="text-gray-600">No rounds recorded yet. Get out there and play!</p>
              )}
            </div>
          </>
        )}

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
