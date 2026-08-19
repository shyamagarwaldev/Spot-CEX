gen:
	@protoc \
		--proto_path=protobuf "protobuf/matching.proto" \
		--go_out=genproto/matching \
		--go_opt=paths=source_relative \
		--go-grpc_out=genproto/matching \
		--go-grpc_opt=paths=source_relative
		