$exec_name = "dsbg.exe"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}

# go build -ldflags="-s -w" .
go build .


# ./dsbg.exe -template -title "MY Awesome Post" -description "My awesome description"
# ./dsbg.exe -title "My Awesome Blog" -description "My awesome description" -watch -css-path "assets/style-colorful.css"
# ./dsbg.exe -title "My Awesome Blog" -description "My awesome description" -watch -style "dark"
./dsbg.exe -h