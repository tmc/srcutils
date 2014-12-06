package main

func main() {
	x(42, "life")
}

func init() {
	z(31, "foobar")
}

func init() {
	foobar := func() {
		x(31, "foobar")
	}
	foobar()
}

func init() {
	new(dummy).test()
}

type dummy struct{}

func (d *dummy) test() {
	baz := &struct{ foo func() }{
		func() {
			x(7, "oi")
		},
	}
	baz.foo()
}
