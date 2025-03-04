$exec_name = "dsbg.exe"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}


# go build -ldflags="-s -w" .
go build .

$Host.UI.RawUI.ForegroundColor = "green"
Write-Host "dsbg.exe -h:"
$Host.UI.RawUI.ForegroundColor = "white"

./dsbg.exe -h

$Host.UI.RawUI.ForegroundColor = "green"
Write-Host "dsbg.exe:"
$Host.UI.RawUI.ForegroundColor = "white"

./dsbg.exe