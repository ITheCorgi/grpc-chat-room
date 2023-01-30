IMPORTS_DIR=./plugins
API_SRC_DIR=./api
GEN_DST_DIR=./pkg/api
SERVER_DIR_NAME = b2b-chat
SERVER_BINARY_NAME = b2b-chat
CLIENT_DIR_NAME = b2b-client
CLIENT_BINARY_NAME = b2b-client
CURRENT_PATH=$(shell pwd)

.PHONY: api
api: update-protoc
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	protoc -I=$(IMPORTS_DIR) --go_out=$(GEN_DST_DIR) --go-grpc_out=$(GEN_DST_DIR) --validate_out="lang=go:./pkg/api" --proto_path $(API_SRC_DIR) $(API_SRC_DIR)/chat.proto

.PHONY: docker-up
docker-up:
	docker-compose -f docker-compose.yaml up --build

.PHONY: docker-down
docker-down: ## Stop docker containers and clear artefacts.
	docker-compose -f docker-compose.yaml down
	docker system prune

.PHONY: build_server
build_server:
	go build -race -o $(SERVER_BINARY_NAME) ./cmd/$(SERVER_DIR_NAME)/main.go

.PHONY: build_client
build_client:
	go build -race -o $(CLIENT_BINARY_NAME).exe ./cmd/$(CLIENT_DIR_NAME)/main.go

.PHONY: run_client
run_client:
	@:go run ./$(CLIENT_BINARY_NAME) $(ARGS)

.PHONY: update-protoc
 update-protoc:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
