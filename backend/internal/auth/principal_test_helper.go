package auth

import "context"

// WithPrincipalForTest кладёт principal в контекст. Используется в тестах из других пакетов,
// которым нужно обойти auth middleware, но при этом протестировать handlerы, читающие PrincipalFrom.
func WithPrincipalForTest(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}
