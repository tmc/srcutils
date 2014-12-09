package main

import "golang.org/x/net/context"

func x(ctx context.Context, a int, b string) {
	y(ctx, a, b)
}
