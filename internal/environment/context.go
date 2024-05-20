package environment

import "context"

type key string

const (
	keyEnv key = "environment"
)

// CtxWithEnv puts passed env into the context.
func CtxWithEnv(ctx context.Context, env Env) context.Context {
	return context.WithValue(ctx, keyEnv, env)
}

// EnvFromCtx returns environment, if any, previously
// put in the context with CtxWithEnv.
func EnvFromCtx(ctx context.Context) Env {
	v := ctx.Value(keyEnv)
	if v == nil {
		return ""
	}

	env, ok := v.(Env)
	if !ok {
		return ""
	}

	return env
}
