import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useLeague, LeagueProvider } from './LeagueContext';
import { api } from '../lib/api';
import { useAuth } from '@clerk/clerk-react';

// Mock dependencies
vi.mock('@clerk/clerk-react', () => ({
    useAuth: vi.fn(),
}));

vi.mock('../lib/api', () => ({
    api: {
        getCurrentUser: vi.fn(),
        listLeagues: vi.fn(),
        listLeagueMembers: vi.fn(),
        getLeague: vi.fn(),
        setAuthTokenProvider: vi.fn(),
    },
}));

describe('LeagueContext', () => {
    const mockUserId = 'user_123';
    const mockPlayerId = 'player_456';
    const mockLeagueId = 'league_789';
    const mockGetToken = vi.fn().mockResolvedValue('mock_token');

    beforeEach(() => {
        vi.clearAllMocks();
        (useAuth as any).mockReturnValue({
            userId: mockUserId,
            getToken: mockGetToken,
        });
    });

    it('should set admin role when user is admin of selected league', async () => {
        const mockLeague = {
            id: mockLeagueId,
            name: 'Test League',
            description: 'Test Description',
            createdBy: mockPlayerId,
            createdAt: new Date().toISOString(),
        };

        const mockUserInfo = {
            linked: true,
            player: {
                id: mockPlayerId,
                name: 'Test Player',
                email: 'test@example.com',
                active: true,
                established: false,
                createdAt: new Date().toISOString(),
            },
        };

        const mockMembers = [
            {
                id: 'member_1',
                leagueId: mockLeagueId,
                playerId: mockPlayerId,
                role: 'admin',
                joinedAt: new Date().toISOString(),
            },
        ];

        (api.getCurrentUser as any).mockResolvedValue(mockUserInfo);
        (api.listLeagues as any).mockResolvedValue([mockLeague]);
        (api.listLeagueMembers as any).mockResolvedValue(mockMembers);
        (api.getLeague as any).mockResolvedValue(mockLeague);

        const wrapper = ({ children }: { children: React.ReactNode }) => (
            <LeagueProvider>{children}</LeagueProvider>
        );

        const { result } = renderHook(() => useLeague(), { wrapper });

        // Wait for initial load
        await waitFor(() => {
            expect(result.current.isLoading).toBe(false);
        });

        // Select the league
        await result.current.selectLeague(mockLeagueId);

        // Wait for league selection to complete
        await waitFor(() => {
            expect(result.current.currentLeague).toEqual(mockLeague);
            expect(result.current.userRole).toBe('admin');
        });
    });

    it('should fetch membership info when selecting league directly via URL (empty userLeagues)', async () => {
        const mockLeague = {
            id: mockLeagueId,
            name: 'Test League',
            description: 'Test Description',
            createdBy: mockPlayerId,
            createdAt: new Date().toISOString(),
        };

        const mockUserInfo = {
            linked: true,
            player: {
                id: mockPlayerId,
                name: 'Test Player',
                email: 'test@example.com',
                active: true,
                established: false,
                createdAt: new Date().toISOString(),
            },
        };

        const mockMembers = [
            {
                id: 'member_1',
                leagueId: mockLeagueId,
                playerId: mockPlayerId,
                role: 'admin',
                joinedAt: new Date().toISOString(),
            },
        ];

        // Simulate direct URL navigation - no leagues loaded yet
        (api.getCurrentUser as any).mockResolvedValue(mockUserInfo);
        (api.listLeagues as any).mockResolvedValue([]);
        (api.listLeagueMembers as any).mockResolvedValue(mockMembers);
        (api.getLeague as any).mockResolvedValue(mockLeague);

        const wrapper = ({ children }: { children: React.ReactNode }) => (
            <LeagueProvider>{children}</LeagueProvider>
        );

        const { result } = renderHook(() => useLeague(), { wrapper });

        // Wait for initial load (which will have empty leagues)
        await waitFor(() => {
            expect(result.current.isLoading).toBe(false);
        });

        // Now select a league directly (simulating URL navigation)
        await result.current.selectLeague(mockLeagueId);

        // Wait for league selection to complete
        await waitFor(() => {
            expect(result.current.currentLeague).toEqual(mockLeague);
            expect(result.current.userRole).toBe('admin');
        });

        // Verify that we fetched membership info
        expect(api.listLeagueMembers).toHaveBeenCalledWith(mockLeagueId);
    });

    it('should set player role when user is not admin', async () => {
        const mockLeague = {
            id: mockLeagueId,
            name: 'Test League',
            description: 'Test Description',
            createdBy: 'other_player',
            createdAt: new Date().toISOString(),
        };

        const mockUserInfo = {
            linked: true,
            player: {
                id: mockPlayerId,
                name: 'Test Player',
                email: 'test@example.com',
                active: true,
                established: false,
                createdAt: new Date().toISOString(),
            },
        };

        const mockMembers = [
            {
                id: 'member_1',
                leagueId: mockLeagueId,
                playerId: mockPlayerId,
                role: 'player',
                joinedAt: new Date().toISOString(),
            },
        ];

        (api.getCurrentUser as any).mockResolvedValue(mockUserInfo);
        (api.listLeagues as any).mockResolvedValue([mockLeague]);
        (api.listLeagueMembers as any).mockResolvedValue(mockMembers);
        (api.getLeague as any).mockResolvedValue(mockLeague);

        const wrapper = ({ children }: { children: React.ReactNode }) => (
            <LeagueProvider>{children}</LeagueProvider>
        );

        const { result } = renderHook(() => useLeague(), { wrapper });

        // Wait for initial load
        await waitFor(() => {
            expect(result.current.isLoading).toBe(false);
        });

        // Select the league
        await result.current.selectLeague(mockLeagueId);

        // Wait for league selection to complete
        await waitFor(() => {
            expect(result.current.currentLeague).toEqual(mockLeague);
            expect(result.current.userRole).toBe('player');
        });
    });
});
