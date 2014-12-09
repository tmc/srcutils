package main

import "golang.org/x/net/context"

func y(ctx context.Context, a int, b string) {
	z(ctx, a, b)
}
