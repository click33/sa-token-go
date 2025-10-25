module github.com/click33/sa-token-go/examples/session-demo

go 1.21

require (
	github.com/click33/sa-token-go/core v0.1.1
	github.com/click33/sa-token-go/storage/memory v0.1.1
	github.com/click33/sa-token-go/stputil v0.1.1
)

replace (
	github.com/click33/sa-token-go/core => ../../core
	github.com/click33/sa-token-go/storage/memory => ../../storage/memory
	github.com/click33/sa-token-go/stputil => ../../stputil
)
