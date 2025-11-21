import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLeague } from '../../contexts/LeagueContext';
import { api } from '../../lib/api';
import { League, LeagueMember } from '../../types';
import { Plus, Trophy, Users, ArrowRight } from 'lucide-react';

export default function LeagueList() {
    const navigate = useNavigate();
    const { selectLeague, userLeagues } = useLeague();
    const [leagues, setLeagues] = useState<League[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const loadLeagues = async () => {
            try {
                const allLeagues = await api.listLeagues();
                setLeagues(allLeagues);
            } catch (error) {
                console.error('Failed to load leagues:', error);
            } finally {
                setIsLoading(false);
            }
        };

        loadLeagues();
    }, []);

    const handleSelectLeague = (leagueId: string) => {
        selectLeague(leagueId);
        navigate('/');
    };

    if (isLoading) {
        return (
            <div className="flex justify-center items-center min-h-screen">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-emerald-600"></div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-4xl mx-auto">
                <div className="flex justify-between items-center mb-8">
                    <div>
                        <h1 className="text-3xl font-bold text-gray-900">Your Leagues</h1>
                        <p className="mt-2 text-gray-600">Select a league to manage or view</p>
                    </div>
                    <button
                        onClick={() => navigate('/leagues/create')}
                        className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-emerald-600 hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-emerald-500"
                    >
                        <Plus className="h-5 w-5 mr-2" />
                        Create New League
                    </button>
                </div>

                {leagues.length === 0 ? (
                    <div className="text-center py-12 bg-white rounded-lg shadow">
                        <Trophy className="mx-auto h-12 w-12 text-gray-400" />
                        <h3 className="mt-2 text-sm font-medium text-gray-900">No leagues found</h3>
                        <p className="mt-1 text-sm text-gray-500">Get started by creating a new league.</p>
                        <div className="mt-6">
                            <button
                                onClick={() => navigate('/leagues/create')}
                                className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-emerald-600 hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-emerald-500"
                            >
                                <Plus className="h-5 w-5 mr-2" />
                                Create League
                            </button>
                        </div>
                    </div>
                ) : (
                    <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
                        {leagues.map((league) => {
                            const member = userLeagues.find((l: LeagueMember) => l.league_id === league.id);
                            const role = member?.role || 'none';

                            return (
                                <div
                                    key={league.id}
                                    className="bg-white overflow-hidden shadow rounded-lg hover:shadow-md transition-shadow duration-200 cursor-pointer border border-gray-100"
                                    onClick={() => handleSelectLeague(league.id)}
                                >
                                    <div className="p-6">
                                        <div className="flex items-center justify-between mb-4">
                                            <div className="bg-emerald-100 rounded-full p-3">
                                                <Trophy className="h-6 w-6 text-emerald-600" />
                                            </div>
                                            {role === 'admin' && (
                                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                                    Admin
                                                </span>
                                            )}
                                        </div>
                                        <h3 className="text-lg font-medium text-gray-900 truncate">
                                            {league.name}
                                        </h3>
                                        <p className="mt-1 text-sm text-gray-500 line-clamp-2 min-h-[2.5rem]">
                                            {league.description || 'No description provided'}
                                        </p>
                                        <div className="mt-4 flex items-center justify-between text-sm text-gray-500">
                                            <div className="flex items-center">
                                                <Users className="flex-shrink-0 mr-1.5 h-4 w-4 text-gray-400" />
                                                <span>View League</span>
                                            </div>
                                            <ArrowRight className="h-4 w-4 text-gray-400" />
                                        </div>
                                    </div>
                                    <div className="bg-gray-50 px-6 py-3 border-t border-gray-100">
                                        <div className="text-xs text-gray-500">
                                            Created {new Date(league.created_at).toLocaleDateString()}
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
}
