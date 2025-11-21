import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLeague } from '../../contexts/LeagueContext';
import { api } from '../../lib/api';
import { Trophy, ArrowLeft } from 'lucide-react';

export default function CreateLeague() {
    const navigate = useNavigate();
    const { selectLeague, refreshLeagues } = useLeague();
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSubmitting(true);
        setError(null);

        try {
            const league = await api.createLeague({
                name,
                description,
            });

            // Refresh user's leagues to include the new one
            await refreshLeagues();

            // Select the new league and navigate to dashboard
            selectLeague(league.id);
            navigate('/');
        } catch (err) {
            console.error('Failed to create league:', err);
            setError('Failed to create league. Please try again.');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md mx-auto">
                <div className="mb-8">
                    <button
                        onClick={() => navigate('/leagues')}
                        className="flex items-center text-sm text-gray-500 hover:text-gray-700"
                    >
                        <ArrowLeft className="h-4 w-4 mr-1" />
                        Back to Leagues
                    </button>
                </div>

                <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
                    <div className="sm:mx-auto sm:w-full sm:max-w-md mb-6">
                        <div className="flex justify-center">
                            <div className="bg-emerald-100 rounded-full p-3">
                                <Trophy className="h-8 w-8 text-emerald-600" />
                            </div>
                        </div>
                        <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
                            Create a New League
                        </h2>
                        <p className="mt-2 text-center text-sm text-gray-600">
                            Start a new golf league and invite players
                        </p>
                    </div>

                    <form className="space-y-6" onSubmit={handleSubmit}>
                        {error && (
                            <div className="rounded-md bg-red-50 p-4">
                                <div className="flex">
                                    <div className="ml-3">
                                        <h3 className="text-sm font-medium text-red-800">{error}</h3>
                                    </div>
                                </div>
                            </div>
                        )}

                        <div>
                            <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                                League Name
                            </label>
                            <div className="mt-1">
                                <input
                                    id="name"
                                    name="name"
                                    type="text"
                                    required
                                    value={name}
                                    onChange={(e) => setName(e.target.value)}
                                    className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-emerald-500 focus:border-emerald-500 sm:text-sm"
                                    placeholder="e.g. Sunday Morning Golfers"
                                />
                            </div>
                        </div>

                        <div>
                            <label htmlFor="description" className="block text-sm font-medium text-gray-700">
                                Description
                            </label>
                            <div className="mt-1">
                                <textarea
                                    id="description"
                                    name="description"
                                    rows={3}
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-emerald-500 focus:border-emerald-500 sm:text-sm"
                                    placeholder="Brief description of your league..."
                                />
                            </div>
                        </div>

                        <div>
                            <button
                                type="submit"
                                disabled={isSubmitting}
                                className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-emerald-600 hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-emerald-500 disabled:opacity-50"
                            >
                                {isSubmitting ? 'Creating...' : 'Create League'}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    );
}
