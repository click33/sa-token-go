package gozero

// ctxKey is an unexported type for context keys to avoid collisions.
// ctxKey 为未导出的 context 键类型，避免与其他库 context 键冲突。
type ctxKey int

const (
	// ctxKeyToken holds token string parsed by TokenInterceptor.
	// ctxKeyToken 存放 TokenInterceptor 解析出的 token 字符串。
	ctxKeyToken ctxKey = iota + 1
	// ctxKeySaToken holds *core.SaTokenContext after auth middleware.
	// ctxKeySaToken 存放鉴权中间件注入的 Sa-Token 上下文。
	ctxKeySaToken
	// ctxKeyLoginID holds login id written by PathAuthMiddleware.
	// ctxKeyLoginID 存放 PathAuth 校验通过后的 loginID。
	ctxKeyLoginID
)
