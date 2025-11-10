module github.com/click33/sa-token-go/integrations/chi

go 1.21

require (
	github.com/click33/sa-token-go/core v0.1.3
	github.com/click33/sa-token-go/stputil v0.0.0-20251017234446-3cf2bdee68cc
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace (
	github.com/click33/sa-token-go/core => ../../core
	github.com/click33/sa-token-go/stputil => ../../stputil
)
