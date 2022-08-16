package main

import (
	"flag"
	"log"
	tgClient "telegram-bot-for-articles/clients/telegram"
	event_consumer "telegram-bot-for-articles/consumer/event-consumer"
	"telegram-bot-for-articles/events/telegram"
	"telegram-bot-for-articles/storage/files"
)

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "storage"
	batchSize   = 100
)

func main() {
	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, mustToken()),
		files.New(storagePath),
	)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal(err)
	}
	// fetcher = fetcher.New()

	// processor = processor.New()

	// consumer.Start(fetcher, processor)
}

func mustToken() string {
	// bot -tg-bot-token 'my token'
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access to telegram bot",
	)
	flag.Parse()
	if *token == "" {
		log.Fatal("token is not specified")
	}
	return *token
}
