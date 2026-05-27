module github.com/sa-tokens/sa-token-go/examples/multi-certification

go 1.25.3

require (
	github.com/sa-tokens/sa-token-go/core v0.2.1
	github.com/sa-tokens/sa-token-go/storage/memory v0.2.1
	github.com/sa-tokens/sa-token-go/stputil v0.2.1
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/panjf2000/ants/v2 v2.12.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
)

replace github.com/sa-tokens/sa-token-go/storage/memory => ../../storage/memory

replace github.com/sa-tokens/sa-token-go/core => ../../core

replace github.com/sa-tokens/sa-token-go/stputil => ../../stputil
