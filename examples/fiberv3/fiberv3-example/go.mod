module github.com/sa-tokens/sa-token-go/examples/fiberv3-example

go 1.25.0

require (
	github.com/gofiber/fiber/v3 v3.4.0
	github.com/sa-tokens/sa-token-go/core v0.2.4
	github.com/sa-tokens/sa-token-go/integrations/fiberv3 v0.2.4
	github.com/sa-tokens/sa-token-go/storage/memory v0.2.4
	github.com/sa-tokens/sa-token-go/stputil v0.2.4
)

require (
	github.com/andybalholm/brotli v1.2.2 // indirect
	github.com/gofiber/schema v1.8.0 // indirect
	github.com/gofiber/utils/v2 v2.1.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.19.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/panjf2000/ants/v2 v2.12.0 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/tinylib/msgp v1.6.4 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.72.0 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)

// 本地 workspace 开发：fiberv3 尚未发布到代理时用 replace 指向仓库内模块
replace (
	github.com/sa-tokens/sa-token-go/core => ../../../core
	github.com/sa-tokens/sa-token-go/integrations/fiberv3 => ../../../integrations/fiberv3
	github.com/sa-tokens/sa-token-go/storage/memory => ../../../storage/memory
	github.com/sa-tokens/sa-token-go/stputil => ../../../stputil
)
