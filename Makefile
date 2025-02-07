# Makefile
test:
    @echo "Makefile is working!"
# Define the protoc command to generate Go code and gRPC stubs
gen:
    @echo "Generating proto files..."
    @protoc \
        --proto_path=proto \
        "proto/auth.proto" \
        --go_out=pb --go_opt=paths=source_relative \
        --go-grpc_out=pb --go-grpc_opt=paths=source_relative

.PHONY: gen