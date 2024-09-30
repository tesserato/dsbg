$exec_name = "dsbg.exe"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}

# go build -ldflags="-s -w" .
go build .

Remove-Item "public/*" -Recurse -Force

# ./dsbg.exe -template -title "MY Awesome Post" -description "My awesome description"
./dsbg.exe -title "MY Awesome Blog" -description "My awesome description"