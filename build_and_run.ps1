$exec_name = "dsbg.exe"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}

# go build -ldflags="-s -w" .
go build .

Remove-Item "docs/*" -Recurse -Force

copy README.md sample_content/01_readme.md

magick -density 376 -background none "logo.svg" "sample_content/01_dsbg_logo.webp"
magick -background none "sample_content/01_dsbg_logo.webp" -fill red -opaque black -blur 0x1  -crop 167x167+0+0  "assets/favicon.ico"

./dsbg.exe -template -title "My Awesome Post" -description "A sample template" -output-path "sample_content"

start chrome http://localhost:666/index.html

./dsbg.exe -title "Dead Simple Blog Generator" -description "Welcome to the DSBG (Dead Simple Blog Generator) blog" -watch -open-in-new-tab -css-path "assets/style.css" -input-path "sample_content" -output-path "docs"