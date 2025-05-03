@echo off
setlocal

set PROTO_DIR=proto

if not exist "%PROTO_DIR%\gen" mkdir "%PROTO_DIR%\gen"

protoc --go_out=. --go_opt=paths=source_relative ^
    --go-grpc_out=. --go-grpc_opt=paths=source_relative ^
    "%PROTO_DIR%\task.proto"

echo Proto files successfully generated 