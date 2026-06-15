package errors

import "errors"

// Предопределённые ошибки проекта.
// Все ошибки доступны для сравнения через errors.Is().
var (
	// DoesNotExistError возникает при попытке доступа к несуществующему элементу.
	// Используется в пакетах:
	//   - mapper: при отсутствии ключа в маппере
	//   - linker: при отсутствии ключа в линкере
	//   - hash_slice: при попытке доступа к несуществующему элементу
	//
	// Пример проверки:
	//
	//	_, err := mapper.Get("unknown")
	//	if errors.Is(err, DoesNotExistError) {
	//	    // обрабатываем отсутствие элемента
	//	}
	DoesNotExistError = errors.New("does not exist")
)
