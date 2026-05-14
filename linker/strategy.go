package linker

const (
	Optimization Strategy = iota
	CurrentLinker
	IncomingLinker
)

type Strategy uint8
