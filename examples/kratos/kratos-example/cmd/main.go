package main

import (
	"context"
	"fmt"
	"github.com/click33/sa-token-go/core/adapter"
	v1 "github.com/click33/sa-token-go/examples/kratos/kratos-example/api/helloworld/v1"
	sakratos "github.com/click33/sa-token-go/integrations/kratos"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/click33/sa-token-go/stputil"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/transport/http"
)

var (
	Name string
)

type server struct {
	v1.UnimplementedUserServer
}

func (s server) GetUserInfo(ctx context.Context, request *v1.GetUserInfoRequest) (*v1.GetUserInfoReply, error) {
	fmt.Println("==============GetUserInfo==============")

	kratosCtx := sakratos.NewKratosContext(ctx)

	// 测试 GetHeader
	fmt.Printf("GetHeader(Authorization): %s\n", kratosCtx.GetHeader("Authorization"))

	// 测试 GetQuery
	fmt.Printf("GetQuery(id): %s\n", kratosCtx.GetQuery("id"))

	// 测试 GetCookie
	fmt.Printf("GetCookie(session): %s\n", kratosCtx.GetCookie("session"))

	// 测试 GetClientIP
	fmt.Printf("GetClientIP: %s\n", kratosCtx.GetClientIP())

	// 测试 GetMethod
	fmt.Printf("GetMethod: %s\n", kratosCtx.GetMethod())

	// 测试 GetPath
	fmt.Printf("GetPath: %s\n", kratosCtx.GetPath())

	// 测试 GetURL
	fmt.Printf("GetURL: %s\n", kratosCtx.GetURL())

	// 测试 GetUserAgent
	fmt.Printf("GetUserAgent: %s\n", kratosCtx.GetUserAgent())

	// 测试 Set 和 Get
	kratosCtx.Set("test_key", "test_value")
	if val, exists := kratosCtx.Get("test_key"); exists {
		fmt.Printf("Get(test_key): %v\n", val)
	}

	// 测试 GetString
	kratosCtx.Set("string_key", "string_value")
	fmt.Printf("GetString(string_key): %s\n", kratosCtx.GetString("string_key"))

	// 测试 MustGet
	kratosCtx.Set("must_key", "must_value")
	fmt.Printf("MustGet(must_key): %v\n", kratosCtx.MustGet("must_key"))

	// 测试 GetHeaders
	headers := kratosCtx.GetHeaders()
	fmt.Printf("GetHeaders: %v\n", headers)

	// 测试 GetQueryAll
	queryAll := kratosCtx.GetQueryAll()
	fmt.Printf("GetQueryAll: %v\n", queryAll)

	// 测试 SetHeader
	kratosCtx.SetHeader("X-Custom-Header", "custom-value")
	fmt.Println("SetHeader: X-Custom-Header set to custom-value")

	// 测试 SetCookie
	kratosCtx.SetCookie("test_cookie", "cookie_value", 3600, "/", "", false, true)
	fmt.Println("SetCookie: test_cookie set")

	// 测试 SetCookieWithOptions
	kratosCtx.SetCookieWithOptions(&adapter.CookieOptions{
		Name:     "test_cookie_options",
		Value:    "options_value",
		MaxAge:   7200,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: true,
		SameSite: "Lax",
	})
	fmt.Println("SetCookieWithOptions: test_cookie_options set")

	// 测试 IsAborted
	fmt.Printf("IsAborted (before): %v\n", kratosCtx.IsAborted())

	// 测试 Abort
	kratosCtx.Abort()
	fmt.Printf("IsAborted (after Abort): %v\n", kratosCtx.IsAborted())

	fmt.Println("==============All context functions tested==============")

	return &v1.GetUserInfoReply{}, nil
}

func (s server) Login(ctx context.Context, request *v1.LoginRequest) (*v1.LoginReply, error) {
	tokenInfo, _ := stputil.Login(request.LoginId)
	_ = stputil.SetPermissions(request.LoginId, []string{"user:some:info", "other:permission"})

	kratosCtx := sakratos.NewKratosContext(ctx)

	// 测试 GetBody
	body, err := kratosCtx.GetBody()
	if err != nil {
		fmt.Printf("GetBody error: %v\n", err)
	} else {
		fmt.Printf("GetBody: %s\n", string(body))
		fmt.Printf("GetBody length: %d\n", len(body))
	}

	return &v1.LoginReply{
		Token: tokenInfo,
	}, nil
}

func main() {
	// 初始化存储
	storage := memory.NewStorage()
	config := sakratos.DefaultConfig()
	manager := sakratos.NewManager(storage, config)
	// 创建 sa-token 中间件
	saPlugin := sakratos.NewPlugin(manager)
	stputil.SetManager(manager)

	// 配置路由规则
	saPlugin.
		// 跳过公开路由
		Skip(v1.OperationUserLogin).
		// 用户信息需要登录
		ExactMatcher(v1.OperationUserGetUserInfo).RequireLogin().RequirePermission("user:some:info").Build()

	httpSrv := http.NewServer(
		http.Address(":8000"),
		http.Middleware(
			saPlugin.Server(),
		),
	)
	s := &server{}
	v1.RegisterUserHTTPServer(httpSrv, s)
	app := kratos.New(
		kratos.Name(Name),
		kratos.Server(
			httpSrv,
		),
	)
	fmt.Println("Server running on port 8000")

	if err := app.Run(); err != nil {
		panic(err)
	}
}
