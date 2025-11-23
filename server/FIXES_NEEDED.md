# Summary of fixes needed for server.go:
# 1. Add respondWithError method
# 2. Fix handleCreateSeason to parse dates and use respondWithError  
# 3. Fix handleCreateMatch to parse dates with noon UTC and use respondWithError
# 4. Replace all http.Error calls with s.respondWithError

Due to the complexity and number of changes needed, I recommend:
1. Manually add the respondWithError method after ServeHTTP (around line 56)
2. Update handleCreateSeason to use custom date parsing (around line 400)
3. Update handleCreateMatch to use custom date parsing (around line 534)
4. Run the fix-errors.ps1 script again to replace http.Error calls

The key fix for the timezone issue is to set parsed dates to noon UTC:
```go
matchDate = time.Date(matchDate.Year(), matchDate.Month(), matchDate.Day(), 12, 0, 0, 0, time.UTC)
```

This prevents the date from shifting when displayed in different timezones.
