$content = Get-Content 'c:\Dev\golf-league-manager\server\internal\api\server.go' -Raw

# Replace http.Error patterns with s.respondWithError
$content = $content -replace 'http\.Error\(w,\s*"([^"]+)",\s*(http\.Status\w+)\)', 's.respondWithError(w, $2, "$1")'
$content = $content -replace 'http\.Error\(w,\s*fmt\.Sprintf\("([^"]+)",\s*([^)]+)\),\s*(http\.Status\w+)\)', 's.respondWithError(w, $3, fmt.Sprintf("$1", $2))'

Set-Content 'c:\Dev\golf-league-manager\server\internal\api\server.go' $content

Write-Host "Replaced http.Error calls in server.go"

$content2 = Get-Content 'c:\Dev\golf-league-manager\server\internal\api\league_handlers.go' -Raw

# Replace http.Error patterns with s.respondWithError
$content2 = $content2 -replace 'http\.Error\(w,\s*"([^"]+)",\s*(http\.Status\w+)\)', 's.respondWithError(w, $2, "$1")'
$content2 = $content2 -replace 'http\.Error\(w,\s*fmt\.Sprintf\("([^"]+)",\s*([^)]+)\),\s*(http\.Status\w+)\)', 's.respondWithError(w, $3, fmt.Sprintf("$1", $2))'

Set-Content 'c:\Dev\golf-league-manager\server\internal\api\league_handlers.go' $content2

Write-Host "Replaced http.Error calls in league_handlers.go"
