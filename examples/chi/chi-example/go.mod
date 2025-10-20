module github.com/click33/sa-token-go/examples/chi-example

go 1.21

require (
	github.com/click33/sa-token-go/core v0.1.1
	github.com/click33/sa-token-go/integrations/chi v0.1.1
	github.com/click33/sa-token-go/storage/memory v0.1.1
	github.com/go-chi/chi/v5 v5.0.11
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace (
	github.com/click33/sa-token-go/core => ../../../core
	github.com/click33/sa-token-go/integrations/chi => ../../../integrations/chi
	github.com/click33/sa-token-go/storage/memory => ../../../storage/memory
)
