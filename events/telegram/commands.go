package telegram

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"telegram-bot-for-articles/lib/e"
	"telegram-bot-for-articles/storage"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

// Смотрим на текст сообщения и по его формату и содержанию будем понимать, какая это команда.
func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)
	log.Printf("got new command '%s' from '%s'", text, username)

	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHelo(chatID)
	default:
		return p.tg.SendMessage(chatID, msgUnknownCommand)

	}
}

func (p *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	//sendMsg := NewMessageSander(chatID, p.tg)

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	IsExists, err := p.storage.IsExists(page)
	if err != nil {
		return err
	}
	if IsExists {
		//return sendMsg(msgAlreadyExists)
		return p.tg.SendMessage(chatID, msgAlreadyExists)
	}

	if err := p.storage.Save(page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send random", err) }()

	page, err := p.storage.PickRandom(username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) { // Если пользователь ничего не сохранил, то сообщим ему об этом.
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessage(chatID, page.URL); err != nil { // Если удалось что-то найти, то отправляем ссылку пользователю.
		return err // Если не отправилось, то возвращаем ошибку.
	}

	return p.storage.Remove(page) // Если удалось найти и отправить, то необходимо ее удалить.
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendHelo(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

/*func NewMessageSander(chatID int, tg *telegram.Client) func(string) error {
	return func(msg string) error {
		return tg.SendMessage(chatID, msg)

	}
}*/

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
