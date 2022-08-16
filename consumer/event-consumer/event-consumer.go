package event_consumer

import (
	"log"
	"telegram-bot-for-articles/events"
	"time"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

func (c Consumer) Start() error {
	for {
		gotEvents, err := c.fetcher.Fetch(c.batchSize)
		if err != nil {
			log.Printf("[ERR] consumer: %s", err.Error())

			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)

			continue
		}

		if err := c.handleEvents(gotEvents); err != nil {
			log.Print(err)

			continue
		}
	}
}

/*
1. Потеря событий: ретраи, возвращение в хранилище, фоллбэк (сохранение не в сторедж, а в локальный файл или
в оперативной памяти), подтверждение для фетчера.
2. Обработка всей пачки: останавливаеться после первой ошибки, счетчик ошибок.
3. Параллельная обработка
*/
func (c *Consumer) handleEvents(events []events.Event) error {
	for _, event := range events { // Перебирает события.
		log.Printf("got new event: %s", event.Text)

		if err := c.processor.Process(event); err != nil { // Пытается их обработать и, если что-то пошло не так с одним из них, то функция пропускает обработку
			log.Printf("can't handle event: %s", err.Error())

			continue
		}
	}

	return nil
}
