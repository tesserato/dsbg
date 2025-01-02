$exec_name = "dsbg.exe"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}

# go build -ldflags="-s -w" .
go build .

Remove-Item "public/*" -Recurse -Force

start chrome http://localhost:666/index.html

# ./dsbg.exe -template -title "MY Awesome Post" -description "My awesome description"
# ./dsbg.exe -title "My Awesome Blog" -description "My awesome description" -watch -css-path "assets/style-colorful.css"
# ./dsbg.exe -title "My Awesome Blog" -description "My awesome description" -watch -style "dark"
./dsbg.exe -title "Dead Simple Blog Generator" -description "Welcome to the DSBG (Dead Simple Blog Generator) blog" -watch -open-in-new-tab -css-path "assets/style.css" -input-path "sample_content"