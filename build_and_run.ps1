$exec_name = "./dsbg.exe"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}

# go build -ldflags="-s -w" .
go build .

./dsbg.exe

# ./codemerge -h

# ./codemerge -ignore="\.git.*,.+\.exe" -excluded-paths-file="excluded_paths.txt"