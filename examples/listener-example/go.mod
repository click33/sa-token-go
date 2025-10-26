module github.com/click33/sa-token-go/examples/listener-example

go 1.21

require (
	github.com/click33/sa-token-go/core v0.1.2
	github.com/click33/sa-token-go/storage/memory v0.1.2
	github.com/click33/sa-token-go/stputil v0.1.2
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace (
	github.com/click33/sa-token-go/core => ../../core
	github.com/click33/sa-token-go/storage/memory => ../../storage/memory
	github.com/click33/sa-token-go/stputil => ../../stputil
)
