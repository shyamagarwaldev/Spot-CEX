GO_BIN := $(shell go env GOPATH)/bin
export PATH := $(GO_BIN):$(PATH)
gen:
	@protoc \
		--proto_path=protobuf "protobuf/matching.proto" \
		--go_out=genproto/matching \
		--go_opt=paths=source_relative \
		--go-grpc_out=genproto/matching \
		--go-grpc_opt=paths=source_relative
	@protoc \
		--proto_path=protobuf "protobuf/account.proto" \
		--go_out=genproto/account \
		--go_opt=paths=source_relative \
		--go-grpc_out=genproto/account \
		--go-grpc_opt=paths=source_relative	
		