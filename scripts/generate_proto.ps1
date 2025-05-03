$ErrorActionPreference = "Stop"

$PROTO_DIR = "proto"

if (-not (Test-Path -Path "$PROTO_DIR\gen")) {
    New-Item -ItemType Directory -Force -Path "$PROTO_DIR\gen"
}

protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    "$PROTO_DIR\task.proto"

Write-Host "Proto files successfully generated" 