package xcontext

import (
	"context"
	"time"
)

type Options struct {
	timeout    time.Duration
	retryCount int
}

type Option func(option *Options)

func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.timeout = d
	}
}

func WithTryCount(retryCount int) Option {
	return func(o *Options) {
		o.retryCount = retryCount
	}
}

func Do(ctx context.Context, f func(context.Context) error, options ...Option) (err error) {
	var o Options
	for _, f := range options {
		f(&o)
	}
	tryCount := o.retryCount
	if tryCount <= 0 {
		tryCount = 1
	}
	for i := 0; i < tryCount; i++ {
		if o.timeout > 0 {
			func() {
				cctx, cancel := context.WithTimeout(ctx, o.timeout)
				err = f(cctx)
				defer cancel()
			}()
		} else {
			err = f(ctx)
		}
		if err == nil {
			return nil
		}
	}
	return err
}
