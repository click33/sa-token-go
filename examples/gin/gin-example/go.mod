module github.com/click33/sa-token-go/examples/gin-example

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/click33/sa-token-go/core v0.1.1
	github.com/click33/sa-token-go/integrations/gin v0.1.1
	github.com/click33/sa-token-go/storage/memory v0.1.1
	github.com/spf13/viper v1.18.2
)

replace (
	github.com/click33/sa-token-go/core => ../../../core
	github.com/click33/sa-token-go/integrations/gin => ../../../integrations/gin
	github.com/click33/sa-token-go/storage/memory => ../../../storage/memory
)

