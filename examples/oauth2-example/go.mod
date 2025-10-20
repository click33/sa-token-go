module github.com/click33/sa-token-go/examples/oauth2-example

go 1.21

require (
	github.com/click33/sa-token-go/core v0.1.1
	github.com/click33/sa-token-go/storage/memory v0.1.1
	github.com/gin-gonic/gin v1.10.0
)

replace (
	github.com/click33/sa-token-go/core => ../../core
	github.com/click33/sa-token-go/storage/memory => ../../storage/memory
)

