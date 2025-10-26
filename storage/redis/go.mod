module github.com/click33/sa-token-go/storage/redis

go 1.21

require (
	github.com/click33/sa-token-go/core v0.1.2
	github.com/redis/go-redis/v9 v9.5.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace github.com/click33/sa-token-go/core => ../../core
