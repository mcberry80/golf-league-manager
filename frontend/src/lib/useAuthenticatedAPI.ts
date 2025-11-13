'use client'

import { useAuth } from '@clerk/nextjs';
import { useEffect, useState } from 'react';
import { api } from './api';

/**
 * Hook that provides an API client configured with Clerk authentication
 */
export function useAuthenticatedAPI() {
  const { getToken, isLoaded, isSignedIn } = useAuth();
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    if (isLoaded && isSignedIn) {
      // Configure the API client with a token getter function
      api.setTokenGetter(async () => {
        try {
          return await getToken();
        } catch (error) {
          console.error('Failed to get Clerk token:', error);
          return null;
        }
      });
      setIsReady(true);
    } else if (isLoaded && !isSignedIn) {
      setIsReady(false);
    }
  }, [isLoaded, isSignedIn, getToken]);

  return { api, isReady };
}
