package main

func main() {
	x(42, "life")
}

func init() {
	z(31, "foobar")
}

func init() {
	alreadyPresent(42)
}

// shouldn't be modified
func alreadyPresent(foo int) {
	x(31, "foobar")
	x(foo, "foobar")
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
