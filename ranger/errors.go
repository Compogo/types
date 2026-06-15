package ranger

import "errors"

// Ошибки, возвращаемые при создании диапазона.
var (
	// UpperBelowLowerError возникает, когда верхняя граница меньше или равна нижней.
	// Например, диапазон [10, 5] является некорректным.
	UpperBelowLowerError = errors.New("upper limit is below the lower limit")

	// OpenBoundError возникает, когда обе границы диапазона являются открытыми (OPEN).
	// Диапазон вида (-∞, +∞) не имеет смысла, так как включает все возможные значения.
	OpenBoundError = errors.New("both borders are open")
)
