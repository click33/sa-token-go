module github.com/sa-tokens/sa-token-go/integrations/echo

go 1.25.0

require (
	github.com/labstack/echo/v4 v4.15.2
	github.com/sa-tokens/sa-token-go/core v0.2.1
	github.com/sa-tokens/sa-token-go/stputil v0.2.1
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/labstack/gommon v0.5.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/panjf2000/ants/v2 v2.12.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.51.0 // indirect
	golang.org/x/net v0.54.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.37.0 // indirect
)

replace github.com/sa-tokens/sa-token-go/core => ../../core

replace github.com/sa-tokens/sa-token-go/stputil => ../../stputil

replace github.com/sa-tokens/sa-token-go/storage/memory => ../../storage/memory
