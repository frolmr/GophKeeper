#!/bin/bash
go test -coverprofile=coverage.tmp ./...

grep -v "internal/server/db/" coverage.tmp |
	grep -v "internal/server/mocks/" |
	grep -v "internal/client/mocks/" |
	grep -v "internal/client/commands/" |
	# grep -v "internal/client/app/app.go" |
	# grep -v "internal/client/client/grpc_client.go" |
	# grep -v "internal/client/adapter/grpc/adapter.go" |
	# grep -v "internal/server/api/api.go" |
	# grep -v "internal/server/app/app.go" |
	# grep -v "internal/server/storage/storage.go" |
	grep -v "swagger/" |
	grep -v "cmd/" |
	grep -v "pkg/proto/" >coverage.out

go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html

rm coverage.tmp
