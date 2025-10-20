module github.com/click33/sa-token-go/examples/echo-example

go 1.21

require (
	github.com/click33/sa-token-go/core v0.1.1
	github.com/click33/sa-token-go/integrations/echo v0.1.1
	github.com/click33/sa-token-go/storage/memory v0.1.1
	github.com/labstack/echo/v4 v4.11.4
)

require (
	github.com/click33/sa-token-go/stputil v0.0.0-20251017234446-3cf2bdee68cc // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/time v0.5.0 // indirect
)

replace (
	github.com/click33/sa-token-go/core => ../../../core
	github.com/click33/sa-token-go/integrations/echo => ../../../integrations/echo
	github.com/click33/sa-token-go/storage/memory => ../../../storage/memory
)
