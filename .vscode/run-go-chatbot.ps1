$ErrorActionPreference = "Stop"

$env:Path = "C:\msys64\ucrt64\bin;C:\Program Files\Go\bin;" + $env:Path
$env:CGO_ENABLED = "1"

Get-Content .env | ForEach-Object {
  if ($_ -match "^\s*([^#=]+)=(.*)$") {
    Set-Item -Path ("Env:" + $matches[1].Trim()) -Value $matches[2].Trim()
  }
}

go run .
