package main

import (
	"fmt"
	"os"

	"log"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	fmt.Println("Bot start Working...")

	dbWorker := NewDataWorker()
	dbWorker.CreateTableIfNotExists()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))

	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message != nil { // If we got a message

			text := update.Message.Text
			switch {
			case strings.HasPrefix(text, "/new"):
				go newPosition(bot, &update, &dbWorker)

			case strings.HasPrefix(text, "/start"):
				go helloHandler(bot, &update)

			case strings.HasPrefix(text, "/show"):
				go ShowData(bot, &update, &dbWorker)

			case strings.HasPrefix(text, "/ticker"):
				go getCryptoPrice(bot, &update)

			case strings.HasPrefix(text, "/del"):
				go deletePosition(bot, &update, &dbWorker)

			case strings.HasPrefix(text, "@"):
				go chatBot(bot, &update, &dbWorker)

			default:
				fmt.Println(text)
			}
		}
	}

}

func helloHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	text := `/new Add new data in database
message stucture: /new [float64] [description]
Example: /new 1000 Find money
Example: /new -200

/show Show your data
message stucture: /show

/del Delete your last data
message stucture: /del

/ticker Show ticker price
message stucture: /ticker [<VALUES>]
example: /ticker EUR BTC

@ call chat gpt
@/ clear chat gpt history1
example: @ Find answer in equation: 2+x=4`

	message := tgbotapi.NewMessage(update.Message.Chat.ID, text)

	bot.Send(message)
}

func getCryptoPrice(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	text := update.Message.Text

	text = strings.Replace(text, "/ticker", "", 1)
	text = strings.TrimSpace(text)
	texts := strings.Split(text, " ")

	answerText := ""
	if strings.TrimSpace(text) == "" {
		message := tgbotapi.NewMessage(update.Message.Chat.ID, "Not enough arguments")

		bot.Send(message)
		return
	}

	dataCh := make(chan BinanceApiJson)
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for _, text := range texts {
		wg.Add(1)
		go func(text string) {
			data := getCryptoAPI(text)
			if data.Msg != "" {
				message := tgbotapi.NewMessage(update.Message.Chat.ID, "VALUE: "+text+" does not exist")

				bot.Send(message)
			}
			dataCh <- data
		}(text)
	}

	stopChan := make(chan struct{})
	go func(dataCH chan BinanceApiJson, stopChan chan struct{}) {
		for {
			select {
			case json := <-dataCH:
				wg.Done()
				fmt.Println("GET [" + json.Symbol + "]")
				if json.Symbol != "" {
					price, _ := strconv.ParseFloat(json.Price, 64)
					mu.Lock()
					answerText += "🌟Name:" + json.Symbol + "\t 💰Price:" + fmt.Sprintf("%.2f", price) + "$\n"
					mu.Unlock()
				}

			case <-stopChan:
				return
			}
		}

	}(dataCh, stopChan)
	wg.Wait()
	close(stopChan)

	message := tgbotapi.NewMessage(update.Message.Chat.ID, answerText)

	bot.Send(message)

}

func newPosition(bot *tgbotapi.BotAPI, update *tgbotapi.Update, dbWorker *DataWorker) {
	text := update.Message.Text
	text = strings.Replace(text, "/new", "", 1)
	text = strings.TrimSpace(text)
	texts := strings.Split(text, " ")
	if text == "" || strings.TrimSpace(text) == "" {
		message := tgbotapi.NewMessage(update.Message.Chat.ID, "Not enough arguments")

		bot.Send(message)
		return
	}
	value, _ := strconv.ParseFloat(texts[0], 32)

	description := ""
	if len(texts) > 1 {
		description = strings.Join(texts[1:], " ")
	}

	dbWorker.InsertData(int(update.Message.From.ID), float32(value), description)

	message := tgbotapi.NewMessage(update.Message.Chat.ID, "New value *"+texts[0]+"* Inserted")

	bot.Send(message)
}

func ShowData(bot *tgbotapi.BotAPI, update *tgbotapi.Update, dbWorker *DataWorker) {
	data, err := dbWorker.GetData(int(update.Message.From.ID))
	if err != nil {
		message := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())

		bot.Send(message)
		return
	}

	text := ""

	if len(data) == 0 {
		message := tgbotapi.NewMessage(update.Message.Chat.ID, "No data")

		bot.Send(message)
		return
	}

	totalSum := 0.0
	for num, val := range data {
		text += fmt.Sprintf("№%v: %v\nDescription: %v\n", num+1, val.Value, val.Description)
		totalSum += float64(val.Value)
	}

	message := tgbotapi.NewMessage(update.Message.Chat.ID, text+fmt.Sprintf("\nTotal: %v", totalSum))

	bot.Send(message)
}

func deletePosition(bot *tgbotapi.BotAPI, update *tgbotapi.Update, dbWorker *DataWorker) {
	row, err := dbWorker.DelData(int(update.Message.From.ID))
	if err != nil {
		message := tgbotapi.NewMessage(update.Message.Chat.ID, "Error while deleting data")
		bot.Send(message)
	}
	message := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Value: %v Description: %v deleted", row.Value, row.Description))

	bot.Send(message)
}

func chatBot(bot *tgbotapi.BotAPI, update *tgbotapi.Update, dbWorker *DataWorker) {
	text := update.Message.Text
	text = strings.Replace(text, "@", "", 1)
	text = strings.TrimSpace(text)
	if text[0] == '/' {
		clearHistory(int(update.Message.From.ID))
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "History clear!"))
	}
	text = chatGPT(int(update.Message.From.ID), text)
	text = strings.Replace(text, "ChatGPT:", "", 1)
	message := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(message)

}
