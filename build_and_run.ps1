$exec_name = "dsbg.exe"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}

# go build -ldflags="-s -w" .
go build .

Remove-Item "public/*" -Recurse -Force

./dsbg.exe