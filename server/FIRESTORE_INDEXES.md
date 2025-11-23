# Firestore Indexes

This directory contains Firestore composite index definitions required for the application's queries.

## Required Indexes

The application requires the following composite indexes:

1. **Seasons Index** (`league_id` + `start_date DESC`)
   - Used by: `ListSeasons` query
   - Allows filtering seasons by league and ordering by start date

2. **Rounds Index** (`league_id` + `player_id` + `date DESC`)
   - Used by: `GetPlayerRounds` query
   - Allows filtering rounds by league and player, ordered by date

3. **Matches Index** (`league_id` + `status`)
   - Used by: `ListMatches` query with status filter
   - Allows filtering matches by league and status

## Deploying Indexes

### Option 1: Using Firebase CLI (Recommended)

```bash
cd server
firebase deploy --only firestore:indexes --project elite-league-manager
```

Or use the provided script:

```powershell
cd server
.\deploy-indexes.ps1
```

### Option 2: Using the Firebase Console

If you encounter an index error, the error message will include a direct link to create the index in the Firebase Console. Click the link and Firebase will automatically configure the index for you.

Example error:
```
The query requires an index. You can create it here: https://console.firebase.google.com/...
```

## Index Build Time

After deploying indexes, it may take a few minutes for Firebase to build them, especially if you have existing data. You can monitor the build status in the [Firebase Console](https://console.firebase.google.com/project/elite-league-manager/firestore/indexes).

## Troubleshooting

If queries are failing with "requires an index" errors:

1. Check if the index exists in the Firebase Console
2. Verify the index status (building vs. enabled)
3. Ensure the `firestore.indexes.json` file is up to date
4. Redeploy indexes if needed

## Local Development

For local development with the Firestore Emulator, indexes are automatically created when queries are executed. No manual deployment is needed.
