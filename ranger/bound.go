package ranger

//go:generate stringer -type=BoundType --output=bound_string.go

// BoundType определяет тип границы диапазона.
const (
	// INCLUSIVE — включающая граница. Значение границы входит в диапазон.
	// Пример: [5, 10] — число 5 входит в диапазон.
	INCLUSIVE BoundType = iota

	// EXCLUSIVE — исключающая граница. Значение границы не входит в диапазон.
	// Пример: (5, 10) — число 5 не входит в диапазон.
	EXCLUSIVE

	// OPEN — открытая граница. Означает отсутствие ограничения (бесконечность).
	// Для нижней границы OPEN означает -∞, для верхней — +∞.
	OPEN
)

// BoundType представляет тип границы.
type BoundType uint8

// LowerBounder определяет контракт для нижней границы диапазона.
// Комбинирует общий интерфейс Bounder со специфичным методом проверки.
type LowerBounder[T Numeric] interface {
	Bounder[T]

	// IsAboveLowerBound проверяет, находится ли значение выше нижней границы.
	// Поведение зависит от типа границы:
	//   - INCLUSIVE: val >= граница
	//   - EXCLUSIVE: val > граница
	//   - OPEN: всегда true
	IsAboveLowerBound(T) bool
}

// UpperBounder определяет контракт для верхней границы диапазона.
// Комбинирует общий интерфейс Bounder со специфичным методом проверки.
type UpperBounder[T Numeric] interface {
	Bounder[T]

	// IsBelowUpperBound проверяет, находится ли значение ниже верхней границы.
	// Поведение зависит от типа границы:
	//   - INCLUSIVE: val <= граница
	//   - EXCLUSIVE: val < граница
	//   - OPEN: всегда true
	IsBelowUpperBound(T) bool
}

// Bounder определяет базовый контракт для любой границы диапазона.
// Предоставляет доступ к типу и значению границы.
type Bounder[T Numeric] interface {
	// Type возвращает тип границы: INCLUSIVE, EXCLUSIVE или OPEN.
	Type() BoundType

	// Value возвращает числовое значение границы.
	// Для OPEN границы возвращает экстремальное значение типа (minValue или maxValue).
	Value() T
}

// baseBoundValue — базовая структура для всех типов границ.
// Содержит общие поля: тип границы и значение.
type baseBoundValue[T Numeric] struct {
	boundType BoundType
	value     T
}

// Type возвращает тип границы.
func (boundValue *baseBoundValue[T]) Type() BoundType {
	return boundValue.boundType
}

// Value возвращает числовое значение границы.
func (boundValue *baseBoundValue[T]) Value() T {
	return boundValue.value
}
