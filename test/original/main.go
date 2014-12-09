package main

import "golang.org/x/net/context"

func main(ctx context.Context) {
	x(ctx, 42, "life")
}

func init(ctx context.Context) {
	z(ctx, 31, "foobar")
}

func init(ctx context.Context) {
	alreadyPresent(ctx, 42)
}

// shouldn't be modified
func alreadyPresent(ctx context.Context, foo int) {
	x(ctx, 31, "foobar")
	x(ctx, foo, "foobar")
}

func init(ctx context.Context) {
	new(dummy).test(ctx)
}

type dummy struct{}

func (d *dummy) test(ctx context.Context) {
	baz := &struct{ foo func() }{
		func() {
			x(ctx, 7, "oi")
		},
	}
	baz.foo()
}
