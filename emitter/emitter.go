// Package emitter предоставляет типобезопасную реализацию паттерна "Издатель-Подписчик" (Pub-Sub).
//
// Emitter позволяет отправлять события (значения типа T) всем подписанным подписчикам
// асинхронно в отдельных горутинах. Поддерживает контекст для отмены и graceful shutdown.
//
// Основные возможности:
//   - Типобезопасная рассылка значений
//   - Асинхронная обработка (каждый подписчик выполняется в своей горутине)
//   - Поддержка контекста для отмены операций
//   - Graceful shutdown через Wait()
//   - Потокобезопасное добавление/удаление подписчиков
//
// Пример использования:
//
//	// Создание эмиттера
//	emitter := NewEmitter[string]()
//
//	// Подписка на события
//	unsubscribe := emitter.Subscribe(func(ctx context.Context, msg string) {
//	    fmt.Println("Получено:", msg)
//	})
//	defer unsubscribe() // не забываем отписываться
//
//	// Отправка событий
//	ctx := context.Background()
//	emitter.Emit(ctx, "Hello, World!")
//	emitter.Emit(ctx, "Second message")
//
//	// Ожидание завершения всех обработчиков
//	emitter.Wait()
//
// Пример с отменой через контекст:
//
//	emitter.Subscribe(func(ctx context.Context, value int) {
//	    select {
//	    case <-ctx.Done():
//	        fmt.Println("Отменено:", ctx.Err())
//	        return
//	    default:
//	        fmt.Println("Обработка:", value)
//	    }
//	})
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	defer cancel()
//	emitter.Emit(ctx, 42)

package emitter

import (
	"context"
	"sync"
	"sync/atomic"
)

// Subscriber представляет функцию-обработчик события.
// Вызывается для каждого полученного значения T.
// Параметр ctx — контекст, переданный в Emit (может использоваться для отмены).
// Параметр value — переданное значение.
//
// Важно: подписчик выполняется в отдельной горутине. Если вам нужна
// синхронная обработка, организуйте её внутри подписчика (например, через каналы).
//
// Пример:
//
//	subscriber := func(ctx context.Context, value string) {
//	    select {
//	    case <-ctx.Done():
//	        log.Printf("Обработка отменена: %v", ctx.Err())
//	        return
//	    default:
//	        process(value)
//	    }
//	}
type Subscriber[T any] func(ctx context.Context, value T)

// UnsubscribeFunc — функция для отписки от событий.
// Вызов этой функции удаляет подписчика из эмиттера.
// Функция идемпотентна — повторные вызовы безопасны.
//
// Пример:
//
//	unsubscribe := emitter.Subscribe(handler)
//	defer unsubscribe() // отписка при выходе из функции
type UnsubscribeFunc func()

// Emitter определяет контракт для издателя событий.
// Позволяет подписываться на события, отправлять их и ожидать завершения обработки.
type Emitter[T any] interface {
	// Emit отправляет значение всем подписчикам.
	// Каждый подписчик выполняется в своей горутине асинхронно.
	// Параметр ctx передаётся каждому подписчику для поддержки отмены.
	Emit(ctx context.Context, value T)

	// Subscribe добавляет нового подписчика и возвращает функцию для отписки.
	// Подписчик будет получать все будущие события.
	// Функция отписки потокобезопасна и может быть вызвана в любой момент.
	Subscribe(subscriber Subscriber[T]) UnsubscribeFunc

	// Wait ожидает завершения обработки текущих событий всеми подписчиками.
	// Полезно для graceful shutdown — дождаться, пока все обработчики закончат работу.
	// Вызов Wait() не блокирует новые Emit.
	Wait()
}

// emitter — внутренняя реализация Emitter.
type emitter[T any] struct {
	subscribers map[uint64]Subscriber[T]
	mutex       sync.RWMutex
	nextID      atomic.Uint64
	wg          sync.WaitGroup
}

// NewEmitter создаёт новый эмиттер событий для типа T.
//
// Пример:
//
//	// Эмиттер для строк
//	strEmitter := NewEmitter[string]()
//
//	// Эмиттер для пользовательских типов
//	type Event struct { ID int; Data string }
//	eventEmitter := NewEmitter[Event]()
func NewEmitter[T any]() Emitter[T] {
	return &emitter[T]{
		subscribers: make(map[uint64]Subscriber[T]),
	}
}

// Emit отправляет значение всем подписчикам.
// Каждый подписчик выполняется в своей горутине, что позволяет обрабатывать
// события параллельно, но не гарантирует порядок выполнения.
//
// Блокировки:
//   - Чтение списка подписчиков блокируется на время копирования ссылок
//   - Сама рассылка не блокируется — подписчики выполняются конкурентно
//
// Пример:
//
//	ctx := context.Background()
//	emitter.Emit(ctx, "сообщение всем")
//
//	// С поддержкой таймаута
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	defer cancel()
//	emitter.Emit(ctx, 42)
func (emitter *emitter[T]) Emit(ctx context.Context, value T) {
	emitter.mutex.RLock()
	defer emitter.mutex.RUnlock()

	emitter.wg.Add(len(emitter.subscribers))
	for _, subscriber := range emitter.subscribers {
		go func(ctx context.Context, subscriber Subscriber[T]) {
			defer emitter.wg.Done()
			subscriber(ctx, value)
		}(ctx, subscriber)
	}
}

// Subscribe добавляет нового подписчика.
// Возвращает функцию UnsubscribeFunc, вызов которой удалит подписчика.
// Несколько подписчиков могут быть добавлены одновременно — всё потокобезопасно.
//
// Подписчик не получает события, отправленные до его подписки.
//
// Пример:
//
//	unsubscribe := emitter.Subscribe(func(ctx context.Context, val int) {
//	    fmt.Printf("Получено: %d\n", val)
//	})
//
//	// ... когда подписчик больше не нужен
//	unsubscribe()
//
// Важно: всегда вызывайте unsubscribe, если подписчик нужно удалить,
// иначе он будет существовать до завершения программы.
func (emitter *emitter[T]) Subscribe(callback Subscriber[T]) UnsubscribeFunc {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	id := emitter.nextID.Add(1)

	emitter.subscribers[id] = callback

	return UnsubscribeFunc(func() {
		emitter.unsubscribe(id)
	})
}

// unsubscribe — внутренний метод для удаления подписчика по ID.
// Вызывается из UnsubscribeFunc.
func (emitter *emitter[T]) unsubscribe(id uint64) {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	delete(emitter.subscribers, id)
}

// Wait ожидает завершения обработки всех текущих событий.
// Используйте перед завершением программы для graceful shutdown.
//
// Важно: Wait ожидает только подписчиков, запущенных через Emit.
// Новые вызовы Emit после Wait будут снова увеличивать счётчик.
//
// Пример graceful shutdown:
//
//	func main() {
//	    emitter := NewEmitter[string]()
//	    emitter.Subscribe(handleMessage)
//
//	    // Отправка сообщений
//	    emitter.Emit(context.Background(), "message")
//
//	    // Ждём завершения обработки
//	    emitter.Wait()
//	    fmt.Println("Все обработчики завершены")
//	}
//
// Пример с сигналами ОС:
//
//	sigChan := make(chan os.Signal, 1)
//	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
//	go func() {
//	    <-sigChan
//	    emitter.Wait()
//	    os.Exit(0)
//	}()
func (emitter *emitter[T]) Wait() {
	emitter.wg.Wait()
}
