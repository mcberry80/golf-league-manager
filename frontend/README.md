# Golf League Manager - Frontend

React-based web application for managing golf league operations with Clerk authentication.

## Features

- **Clerk Authentication**: Secure user authentication and session management
- **Admin Dashboard**: Manage courses, players, matches, and scores
- **Player Portal**: View personal scores, handicaps, and match history
- **League Standings**: Real-time standings and player rankings
- **TypeScript**: Type-safe codebase for better developer experience
- **Tailwind CSS**: Modern, responsive UI design

## Getting Started

### Prerequisites

- Node.js 18+ installed
- Clerk account (get one at [clerk.com](https://clerk.com))
- Running backend API server

### Installation

```bash
# Install dependencies
npm install

# Copy environment template
cp .env.local.example .env.local
```

### Configuration

Edit `.env.local` with your credentials:

```bash
# Clerk Authentication Keys (from Clerk Dashboard)
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...

# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Development

```bash
# Start development server
npm run dev
```

Visit [http://localhost:3000](http://localhost:3000)

### Production Build

```bash
# Build for production
npm run build

# Start production server
npm start
```

## Project Structure

```
frontend/
├── src/
│   ├── app/              # Next.js app router pages
│   │   ├── admin/        # Admin pages
│   │   ├── standings/    # Standings page
│   │   ├── players/      # Player profile pages
│   │   └── page.tsx      # Home page
│   ├── components/       # Reusable React components
│   ├── lib/             # Utility functions and API client
│   │   └── api.ts       # API client for backend
│   └── types/           # TypeScript type definitions
│       └── index.ts     # Shared types
├── public/              # Static assets
└── package.json         # Dependencies and scripts
```

## Pages

### Public Pages
- `/` - Home page with sign in

### Authenticated Pages
- `/standings` - League standings
- `/players` - Player profile and history
- `/admin` - Admin dashboard (requires admin role)
- `/admin/courses` - Manage courses
- `/admin/players` - Manage players
- `/admin/matches` - Schedule matches
- `/admin/scores` - Enter scores

## API Integration

The frontend communicates with the Go backend API:

```typescript
import { api } from '@/lib/api'

// Example: Get player handicap
const handicap = await api.getPlayerHandicap(playerId)

// Example: Create a course
const course = await api.createCourse({
  name: 'Pine Valley',
  par: 36,
  course_rating: 35.5,
  slope_rating: 113,
  hole_handicaps: [1, 2, 3, 4, 5, 6, 7, 8, 9],
  hole_pars: [4, 3, 5, 4, 4, 3, 5, 4, 4]
})
```

## Authentication

Clerk provides:
- Email/password authentication
- Social login (Google, etc.)
- Session management
- User profile management
- Role-based access control

### Protecting Routes

```tsx
import { SignedIn, SignedOut } from '@clerk/nextjs'

export default function ProtectedPage() {
  return (
    <SignedIn>
      {/* Your protected content */}
    </SignedIn>
  )
}
```

## Styling

Uses Tailwind CSS for styling. Customize in:
- `tailwind.config.ts` - Tailwind configuration
- `src/app/globals.css` - Global styles

## Deployment

### Vercel (Recommended)

```bash
# Install Vercel CLI
npm i -g vercel

# Deploy
vercel
```

### Docker

```bash
# Build image
docker build -t golf-league-frontend .

# Run container
docker run -p 3000:3000 golf-league-frontend
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` | Clerk publishable key | Yes |
| `CLERK_SECRET_KEY` | Clerk secret key | Yes |
| `NEXT_PUBLIC_API_URL` | Backend API URL | Yes |

## Development Tips

- Use TypeScript for type safety
- Follow React hooks best practices
- Use the API client in `@/lib/api` for all backend calls
- Add new types to `@/types/index.ts`
- Keep components small and reusable

## Support

For issues or questions, please open an issue on GitHub.
