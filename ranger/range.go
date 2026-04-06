package ranger

import "cmp"

type Ranger[T cmp.Ordered] interface {
	GetLowerBoundValue() *RangeBound[T]
	GetUpperBoundValue() *RangeBound[T]
	Contains(T) bool
}

type Range[T cmp.Ordered] struct {
	LowerBoundValue RangeBound[T]
	UpperBoundValue RangeBound[T]
}

func NewRange[T cmp.Ordered](lowerBoundValue RangeBound[T], upperBoundValue RangeBound[T]) *Range[T] {
	return &Range[T]{LowerBoundValue: lowerBoundValue, UpperBoundValue: upperBoundValue}
}

func (ranger *Range[T]) GetLowerBoundValue() *RangeBound[T] {
	return &ranger.LowerBoundValue
}

func (ranger *Range[T]) GetUpperBoundValue() *RangeBound[T] {
	return &ranger.UpperBoundValue
}

func (ranger *Range[T]) Contains(v T) bool {
	minValue := false
	maxValue := false

	switch ranger.LowerBoundValue.Type() {
	case INCLUSIVE:
		minValue = v >= ranger.LowerBoundValue.Value()
	case EXCLUSIVE:
		minValue = v > ranger.LowerBoundValue.Value()
	default:
		minValue = true
	}

	switch ranger.UpperBoundValue.Type() {
	case INCLUSIVE:
		maxValue = v <= ranger.UpperBoundValue.Value()
	case EXCLUSIVE:
		maxValue = v < ranger.UpperBoundValue.Value()
	default:
		maxValue = true
	}

	return minValue && maxValue
}
