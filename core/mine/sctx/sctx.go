
package sctx

import (
	"context"
	"errors"
	"math/big"
)

var (
	// ErrTargetPrefix is returned when target prefix decoding fails.
	ErrTargetPrefix = errors.New("error decoding prefix string")
)

type (
	HTTPRequestIDKey  struct{}
	requestHostKey    struct{}
	tagKey            struct{}
	targetsContextKey struct{}
	gasPriceKey       struct{}
	gasLimitKey       struct{}
)

// SetHost sets the http request host in the context
func SetHost(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, requestHostKey{}, domain)
}

// GetHost gets the request host from the context
func GetHost(ctx context.Context) string {
	v, ok := ctx.Value(requestHostKey{}).(string)
	if ok {
		return v
	}
	return ""
}

func SetGasLimit(ctx context.Context, limit uint64) context.Context {
	return context.WithValue(ctx, gasLimitKey{}, limit)
}

func GetGasLimit(ctx context.Context) uint64 {
	v, ok := ctx.Value(gasLimitKey{}).(uint64)
	if ok {
		return v
	}
	return 0
}

func SetGasPrice(ctx context.Context, price *big.Int) context.Context {
	return context.WithValue(ctx, gasPriceKey{}, price)
}

func GetGasPrice(ctx context.Context) *big.Int {
	v, ok := ctx.Value(gasPriceKey{}).(*big.Int)
	if ok {
		return v
	}
	return nil
}
