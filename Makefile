.PHONY: all build run proto clean server worker

GO=go
PORT?=8080
GRPC_PORT?=50051
WORKERS_COUNT?=2

all: proto build

proto:
	@echo "Generating proto files..."
	@mkdir -p proto/gen
	@chmod +x scripts/generate_proto.sh
	@./scripts/generate_proto.sh

build: proto
	@echo "Building server..."
	@$(GO) build -o bin/server  cmd/main.go
	@echo "Building gRPC worker..."
	@$(GO) build -o bin/grpc_worker cmd/grpc_worker/main.go

server:
	@echo "Starting server..."
	@./bin/server

worker:
	@echo "Starting gRPC worker..."
	@./bin/grpc_worker

run: build
	@echo "Starting both server and worker..."
	@./bin/server &
	@sleep 2
	@./bin/grpc_worker

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf proto/gen/ 