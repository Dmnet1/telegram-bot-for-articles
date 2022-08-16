package telegram

import (
	"errors"
	"telegram-bot-for-articles/clients/telegram"
	"telegram-bot-for-articles/events"
	"telegram-bot-for-articles/lib/e"
	"telegram-bot-for-articles/storage"
)

type Processor struct {
	tg      *telegram.Client
	offset  int
	storage storage.Storage
}

type Meta struct {
	ChatID   int
	Username string
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

func New(client *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

/*
различие между ивентами и апдейтами заключается в том, что апдейты - это понятие телеграмма
(в другом мессенджере термина апдейт может не быть). Ивент более общая сущность, в нее можно преобразовывать
все, что получаем от других мессенджеров, в каком бы формате они не предоставляли информацию.
*/
func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit) // Получение апдейтов.
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	if len(updates) == 0 { // Если список апдейтов оказался пустым, то функция завершается.
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates)) // Готовим переменную для результата и заранее аллоцируем память для нее.

	for _, u := range updates { // Перебираем все апдейты и преобразуем в тип ивент.
		res = append(res, event(u))
	}

	p.offset = updates[len(updates)-1].ID + 1 // Обновляем параметр оффсет для того, чтобы в следующий раз получить следующую пачку изменений.

	return res, nil // Возвращаем результат.
}

func (p Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	default:
		return e.Wrap("can't process message", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	if err := p.doCmd(event.Text, meta.ChatID, meta.Username); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil

}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)

	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			Username: upd.Message.From.Username,
		}
	}

	return res
}

func fetchText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""
	}
	return upd.Message.Text
}

func fetchType(upd telegram.Update) events.Type {
	if upd.Message == nil {
		return events.Unknown
	}

	return events.Message
}
