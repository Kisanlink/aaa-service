TEST_DIR = ./test

.PHONY: proto migrate migrate-reset run test

# Generate gRPC code in pb folder
window_proto:
	if not exist pb mkdir pb
	protoc -Iproto --go_out=pb --go-grpc_out=pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/auth.proto
	protoc -Iproto --go_out=pb --go-grpc_out=pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/role_permission.proto
	protoc -Iproto --go_out=pb --go-grpc_out=pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/connectRolePermission.proto
	protoc -Iproto --go_out=pb --go-grpc_out=pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/service.proto



linux_proto:
	mkdir -p pb
	protoc -Iproto --go_out=pb --go-grpc_out=pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/auth.proto
	protoc -Iproto --go_out=pb --go-grpc_out=pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/role_permission.proto
	protoc -Iproto --go_out=pb --go-grpc_out=pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/connectRolePermission.proto
	protoc -Iproto --go_out=pb --go-grpc_out=pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/service.proto


migrate:
	go run main.go migrate

migrate-reset:
	go run main.go migrate-reset

air:
	air
run:
	go run main.go


test:
	@echo "Running Go tests..."
	go test -v $(TEST_DIR)/...

.PHONY: test
