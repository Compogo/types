package linker

// Strategy определяет стратегию поведения при объединении (Union) двух Linker'ов.
// Используется в функции UnionByStrategy.
const (
	// Optimization (значение по умолчанию) выбирает наиболее эффективную стратегию
	// в зависимости от размера линкеров (клонируется меньший).
	Optimization Strategy = iota

	// CurrentLinker при объединении клонирует текущий Linker (CurrentLinker)
	// и добавляет в него элементы входящего линкера (IncomingLinker).
	CurrentLinker

	// IncomingLinker при объединении клонирует входящий Linker (IncomingLinker)
	// и добавляет в него элементы текущего линкера (CurrentLinker).
	IncomingLinker
)

// Strategy представляет стратегию слияния двух Linker'ов.
type Strategy uint8
