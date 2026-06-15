package ranger

// exclusiveBoundValue представляет исключающую границу диапазона.
// Значение границы не входит в диапазон.
//
// Примеры:
//   - Нижняя исключающая граница: значение > 5
//   - Верхняя исключающая граница: значение < 10
type exclusiveBoundValue[T Numeric] struct {
	baseBoundValue[T]
}

// NewExclusiveLowerBoundValue создаёт новую исключающую нижнюю границу.
// Возвращает интерфейс LowerBounder для использования в качестве нижней границы диапазона.
//
// Пример:
//
//	lower := NewExclusiveLowerBoundValue(5) // нижняя граница: значение > 5
func NewExclusiveLowerBoundValue[T Numeric](value T) LowerBounder[T] {
	return newExclusiveBoundValue[T](value)
}

// NewExclusiveUpperBoundValue создаёт новую исключающую верхнюю границу.
// Возвращает интерфейс UpperBounder для использования в качестве верхней границы диапазона.
//
// Пример:
//
//	upper := NewExclusiveUpperBoundValue(10) // верхняя граница: значение < 10
func NewExclusiveUpperBoundValue[T Numeric](value T) UpperBounder[T] {
	return newExclusiveBoundValue[T](value)
}

// newExclusiveBoundValue — внутренний конструктор для создания исключающей границы.
func newExclusiveBoundValue[T Numeric](value T) *exclusiveBoundValue[T] {
	return &exclusiveBoundValue[T]{
		baseBoundValue[T]{
			value:     value,
			boundType: EXCLUSIVE,
		},
	}
}

// IsAboveLowerBound проверяет, что значение находится строго выше нижней границы.
// Для исключающей границы: val > boundValue.value
func (boundValue *exclusiveBoundValue[T]) IsAboveLowerBound(val T) bool {
	return val > boundValue.value
}

// IsBelowUpperBound проверяет, что значение находится строго ниже верхней границы.
// Для исключающей границы: val < boundValue.value
func (boundValue *exclusiveBoundValue[T]) IsBelowUpperBound(val T) bool {
	return val < boundValue.value
}
