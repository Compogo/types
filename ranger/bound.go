package ranger

import "cmp"

//go:generate stringer -type=BoundType --output=bound_string.go

const (
	INCLUSIVE BoundType = iota
	EXCLUSIVE
	OPEN
)

type BoundType uint8

type RangeBound[T cmp.Ordered] struct {
	boundType BoundType
	value     T
}

func NewRangeBound[T cmp.Ordered](boundType BoundType, value T) RangeBound[T] {
	return RangeBound[T]{boundType: boundType, value: value}
}

func (bound *RangeBound[T]) Value() T {
	return bound.value
}

func (bound *RangeBound[T]) SetValue(value T) {
	bound.value = value
}

func (bound *RangeBound[T]) Type() BoundType {
	return bound.boundType
}

func Inclusive[T cmp.Ordered](value T) RangeBound[T] {
	return NewRangeBound[T](INCLUSIVE, value)
}

func Exclusive[T cmp.Ordered](value T) RangeBound[T] {
	return NewRangeBound[T](EXCLUSIVE, value)
}

func Open[T cmp.Ordered]() RangeBound[T] {
	var zero T
	return NewRangeBound[T](OPEN, zero)
}
