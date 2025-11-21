import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useLeague } from '../../contexts/LeagueContext'
import api from '../../lib/api'
import type { Course } from '../../types'

export default function CourseManagement() {
    const { currentLeague, userRole, isLoading: leagueLoading } = useLeague()
    const navigate = useNavigate()
    const [courses, setCourses] = useState<Course[]>([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState({
        name: '',
        par: 36,
        course_rating: 36.0,
        slope_rating: 113,
        hole_pars: [4, 4, 4, 4, 4, 4, 4, 4, 4],
        hole_handicaps: [1, 2, 3, 4, 5, 6, 7, 8, 9],
    })

    useEffect(() => {
        if (!leagueLoading && !currentLeague) {
            navigate('/leagues')
            return
        }

        if (currentLeague) {
            loadCourses()
        }
    }, [currentLeague, leagueLoading, navigate])

    async function loadCourses() {
        if (!currentLeague) return

        try {
            const data = await api.listCourses(currentLeague.id)
            setCourses(data)
        } catch (error) {
            console.error('Failed to load courses:', error)
        } finally {
            setLoading(false)
        }
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!currentLeague) return

        try {
            await api.createCourse(currentLeague.id, formData)
            setShowForm(false)
            setFormData({
                name: '',
                par: 36,
                course_rating: 36.0,
                slope_rating: 113,
                hole_pars: [4, 4, 4, 4, 4, 4, 4, 4, 4],
                hole_handicaps: [1, 2, 3, 4, 5, 6, 7, 8, 9],
            })
            loadCourses()
        } catch (error) {
            alert('Failed to create course: ' + (error instanceof Error ? error.message : 'Unknown error'))
        }
    }

    if (leagueLoading || loading) {
        return (
            <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <div className="spinner"></div>
            </div>
        )
    }

    if (!currentLeague || userRole !== 'admin') {
        return null // Will redirect or show access denied in Admin wrapper
    }

    return (
        <div className="min-h-screen" style={{ background: 'var(--gradient-dark)' }}>
            <div className="container animate-fade-in" style={{ paddingTop: 'var(--spacing-2xl)', paddingBottom: 'var(--spacing-2xl)' }}>
                <Link to="/admin" style={{ color: 'var(--color-primary)', textDecoration: 'none', marginBottom: 'var(--spacing-md)', display: 'inline-block' }}>
                    ← Back to Admin
                </Link>

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-xl)' }}>
                    <div>
                        <h1>Course Management</h1>
                        <p className="text-gray-400 mt-1">{currentLeague.name}</p>
                    </div>
                    <button onClick={() => setShowForm(!showForm)} className="btn btn-primary">
                        {showForm ? 'Cancel' : '+ Add Course'}
                    </button>
                </div>

                {showForm && (
                    <div className="card-glass" style={{ marginBottom: 'var(--spacing-xl)' }}>
                        <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Add New Course</h3>
                        <form onSubmit={handleSubmit}>
                            <div className="grid grid-cols-2" style={{ gap: 'var(--spacing-lg)' }}>
                                <div className="form-group">
                                    <label className="form-label">Course Name</label>
                                    <input
                                        type="text"
                                        className="form-input"
                                        value={formData.name}
                                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                        required
                                        placeholder="Pine Valley"
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Par</label>
                                    <input
                                        type="number"
                                        className="form-input"
                                        value={formData.par}
                                        onChange={(e) => setFormData({ ...formData, par: parseInt(e.target.value) })}
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Course Rating</label>
                                    <input
                                        type="number"
                                        step="0.1"
                                        className="form-input"
                                        value={formData.course_rating}
                                        onChange={(e) => setFormData({ ...formData, course_rating: parseFloat(e.target.value) })}
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">Slope Rating</label>
                                    <input
                                        type="number"
                                        className="form-input"
                                        value={formData.slope_rating}
                                        onChange={(e) => setFormData({ ...formData, slope_rating: parseInt(e.target.value) })}
                                        required
                                    />
                                </div>
                            </div>

                            <div className="form-group">
                                <label className="form-label">Hole Pars (9 holes, comma-separated)</label>
                                <input
                                    type="text"
                                    className="form-input"
                                    value={formData.hole_pars.join(',')}
                                    onChange={(e) => setFormData({
                                        ...formData,
                                        hole_pars: e.target.value.split(',').map(v => parseInt(v.trim()) || 4)
                                    })}
                                    placeholder="4,4,4,4,4,4,4,4,4"
                                />
                            </div>

                            <div className="form-group">
                                <label className="form-label">Hole Handicaps (1-9, difficulty rankings)</label>
                                <input
                                    type="text"
                                    className="form-input"
                                    value={formData.hole_handicaps.join(',')}
                                    onChange={(e) => setFormData({
                                        ...formData,
                                        hole_handicaps: e.target.value.split(',').map(v => parseInt(v.trim()) || 1)
                                    })}
                                    placeholder="1,2,3,4,5,6,7,8,9"
                                />
                            </div>

                            <button type="submit" className="btn btn-success">
                                Add Course
                            </button>
                        </form>
                    </div>
                )}

                <div className="card-glass">
                    <h3 style={{ marginBottom: 'var(--spacing-lg)', color: 'var(--color-text)' }}>Courses ({courses.length})</h3>
                    {courses.length === 0 ? (
                        <p style={{ color: 'var(--color-text-muted)' }}>No courses added yet.</p>
                    ) : (
                        <div className="table-container">
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Par</th>
                                        <th>Course Rating</th>
                                        <th>Slope Rating</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {courses.map((course) => (
                                        <tr key={course.id}>
                                            <td style={{ fontWeight: '600' }}>{course.name}</td>
                                            <td>{course.par}</td>
                                            <td>{course.course_rating.toFixed(1)}</td>
                                            <td>{course.slope_rating}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
