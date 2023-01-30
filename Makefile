IMPORTS_DIR=./plugins
API_SRC_DIR=./api
GEN_DST_DIR=./pkg/api
DIR_NAME = b2b-chat
BINARY_NAME = b2b-chat
CURRENT_PATH=$(shell pwd)

.PHONY: api
api:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	protoc -I=$(IMPORTS_DIR) --go_out=$(GEN_DST_DIR) --go-grpc_out=$(GEN_DST_DIR) --validate_out="lang=go:./pkg/api" --proto_path $(API_SRC_DIR) $(API_SRC_DIR)/chat.proto

.PHONY: docker-up
docker-up:
	docker-compose -f docker-compose.yaml up --build

.PHONY: docker-down
docker-down: ## Stop docker containers and clear artefacts.
	docker-compose -f docker-compose.yaml down
	docker system prune

