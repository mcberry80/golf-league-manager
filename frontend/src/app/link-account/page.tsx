'use client'

import { SignedIn, useUser } from '@clerk/nextjs'
import { useState, useEffect } from 'react'
import { useAuthenticatedAPI } from '@/lib/useAuthenticatedAPI'
import Link from 'next/link'

export default function LinkAccountPage() {
  const { user } = useUser()
  const { api, isReady } = useAuthenticatedAPI()
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(true)
  const [linking, setLinking] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [userInfo, setUserInfo] = useState<any>(null)

  useEffect(() => {
    if (isReady) {
      checkExistingLink()
    }
  }, [isReady])

  useEffect(() => {
    if (user?.primaryEmailAddress?.emailAddress) {
      setEmail(user.primaryEmailAddress.emailAddress)
    }
  }, [user])

  const checkExistingLink = async () => {
    try {
      const info = await api.getCurrentUser()
      setUserInfo(info)
      setLoading(false)
    } catch (err) {
      console.error('Failed to check user link:', err)
      setLoading(false)
    }
  }

  const handleLinkAccount = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLinking(true)

    try {
      const player = await api.linkPlayerAccount(email)
      setSuccess(true)
      setUserInfo({ linked: true, player })
    } catch (err: any) {
      setError(err.message || 'Failed to link account')
    } finally {
      setLinking(false)
    }
  }

  return (
    <SignedIn>
      <div className="container mx-auto p-6 max-w-2xl">
        <h1 className="text-4xl font-bold mb-8">Link Your Account</h1>
        
        {loading ? (
          <div className="text-center py-8">
            <p>Loading...</p>
          </div>
        ) : userInfo?.linked ? (
          <div className="bg-green-50 border border-green-300 rounded-lg p-6">
            <h2 className="text-2xl font-semibold mb-4 text-green-800">Account Already Linked!</h2>
            <p className="mb-4">Your Clerk account is linked to:</p>
            <div className="bg-white rounded p-4 mb-4">
              <p><strong>Name:</strong> {userInfo.player.name}</p>
              <p><strong>Email:</strong> {userInfo.player.email}</p>
              <p><strong>Status:</strong> {userInfo.player.active ? 'Active' : 'Inactive'}</p>
              {userInfo.player.is_admin && (
                <p className="text-blue-600 font-semibold">Role: Admin</p>
              )}
            </div>
            <div className="mt-6">
              <Link 
                href="/"
                className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded inline-block"
              >
                Go to Home
              </Link>
            </div>
          </div>
        ) : success ? (
          <div className="bg-green-50 border border-green-300 rounded-lg p-6">
            <h2 className="text-2xl font-semibold mb-4 text-green-800">Account Linked Successfully!</h2>
            <p className="mb-4">You can now access the league features.</p>
            <Link 
              href="/"
              className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded inline-block"
            >
              Go to Home
            </Link>
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow p-6">
            <p className="mb-6 text-gray-700">
              To access league features, you need to link your Clerk account to your player profile.
              Enter the email address that the league administrator used when creating your player account.
            </p>
            
            <form onSubmit={handleLinkAccount}>
              <div className="mb-4">
                <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                  Email Address
                </label>
                <input
                  type="email"
                  id="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="your.email@example.com"
                />
                <p className="mt-2 text-sm text-gray-500">
                  This should match the email your league administrator has on file.
                </p>
              </div>

              {error && (
                <div className="mb-4 bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded">
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={linking}
                className="w-full bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded disabled:bg-gray-400"
              >
                {linking ? 'Linking...' : 'Link Account'}
              </button>
            </form>

            <div className="mt-6 text-center">
              <Link href="/" className="text-blue-500 hover:text-blue-700">
                ← Back to Home
              </Link>
            </div>
          </div>
        )}
      </div>
    </SignedIn>
  )
}
