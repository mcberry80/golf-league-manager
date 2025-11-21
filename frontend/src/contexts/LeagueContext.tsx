import React, { createContext, useContext, useState, useEffect } from 'react';
import { League, LeagueMember } from '../types';
import { api } from '../lib/api';
import { useAuth } from '@clerk/clerk-react';

interface LeagueContextType {
    currentLeague: League | null;
    userRole: 'admin' | 'player' | null;
    userLeagues: LeagueMember[];
    isLoading: boolean;
    selectLeague: (leagueId: string) => void;
    refreshLeagues: () => Promise<void>;
}

const LeagueContext = createContext<LeagueContextType | undefined>(undefined);

export function LeagueProvider({ children }: { children: React.ReactNode }) {
    const { getToken, userId } = useAuth();
    const [currentLeague, setCurrentLeague] = useState<League | null>(null);
    const [userRole, setUserRole] = useState<'admin' | 'player' | null>(null);
    const [userLeagues, setUserLeagues] = useState<LeagueMember[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    // Load user's leagues on mount or auth change
    useEffect(() => {
        const loadLeagues = async () => {
            if (!userId) {
                setIsLoading(false);
                return;
            }

            try {
                const token = await getToken();
                if (!token) return;

                api.setAuthTokenProvider(async () => token);

                // Get current user info which includes leagues
                const userInfo = await api.getCurrentUser();
                if (userInfo.leagues) {
                    setUserLeagues(userInfo.leagues);

                    // Restore selected league from local storage if available
                    const savedLeagueId = localStorage.getItem('selectedLeagueId');
                    if (savedLeagueId) {
                        const member = userInfo.leagues.find(l => l.league_id === savedLeagueId);
                        if (member) {
                            selectLeague(savedLeagueId);
                        } else if (userInfo.leagues.length > 0) {
                            // Default to first league if saved one not found
                            selectLeague(userInfo.leagues[0].league_id);
                        }
                    } else if (userInfo.leagues.length > 0) {
                        // Default to first league
                        selectLeague(userInfo.leagues[0].league_id);
                    }
                }
            } catch (error) {
                console.error('Failed to load leagues:', error);
            } finally {
                setIsLoading(false);
            }
        };

        loadLeagues();
    }, [userId, getToken]);

    const selectLeague = async (leagueId: string) => {
        setIsLoading(true);
        try {
            const league = await api.getLeague(leagueId);
            setCurrentLeague(league);
            localStorage.setItem('selectedLeagueId', leagueId);

            // Set user role for this league
            const member = userLeagues.find(l => l.league_id === leagueId);
            setUserRole(member?.role || null);
        } catch (error) {
            console.error('Failed to select league:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const refreshLeagues = async () => {
        if (!userId) return;
        const token = await getToken();
        if (!token) return;

        const userInfo = await api.getCurrentUser();
        if (userInfo.leagues) {
            setUserLeagues(userInfo.leagues);
        }
    };

    return (
        <LeagueContext.Provider value={{
            currentLeague,
            userRole,
            userLeagues,
            isLoading,
            selectLeague,
            refreshLeagues
        }}>
            {children}
        </LeagueContext.Provider>
    );
}

export function useLeague() {
    const context = useContext(LeagueContext);
    if (context === undefined) {
        throw new Error('useLeague must be used within a LeagueProvider');
    }
    return context;
}
