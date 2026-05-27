module github.com/sa-tokens/sa-token-go/storage/redis

go 1.25.0

require (
	github.com/redis/go-redis/v9 v9.19.0
	github.com/sa-tokens/sa-token-go/core v0.2.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
)

replace github.com/sa-tokens/sa-token-go/core => ../../core
