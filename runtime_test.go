package llb

func fatalRuntimePanic(err error) {
	panic(err.Error())
}
