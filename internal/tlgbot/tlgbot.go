package tlgbot

import (
	"fmt"
	"log"
	"sentabot/internal/astraapi"
	"sentabot/internal/config"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func Pool() error {
	token := config.GetConfig().TlgToken
	acceptedUsers := config.GetConfig().AcceptedUsers

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		msgUser := update.Message.From

		isAccepted := false

		for _, user := range acceptedUsers {
			if user == msgUser.UserName {
				isAccepted = true
				break
			}
		}

		if !isAccepted {
			log.Printf("User %s is not accepted", msgUser.UserName)
			continue
		}

		if update.CallbackQuery != nil {
			if update.CallbackQuery.Data == "list_processes" {
				handleListProcesses(bot, update.CallbackQuery.Message)
				continue
			}

			switch {
			case strings.HasPrefix(update.CallbackQuery.Data, "process_start"):
				handleProcessCommand(bot, update.CallbackQuery, "start")
			case strings.HasPrefix(update.CallbackQuery.Data, "process_stop"):
				handleProcessCommand(bot, update.CallbackQuery, "stop")
			case strings.HasPrefix(update.CallbackQuery.Data, "process_restart"):
				handleProcessCommand(bot, update.CallbackQuery, "restart")
			}

			continue
		}

		switch {
		case strings.HasPrefix(update.Message.Text, "/start"):
			handleStartProcess(bot, update.Message)
		case strings.HasPrefix(update.Message.Text, "/process"):
			handlerProcess(bot, update.Message)
		}
	}

	return nil
}

func handlerProcess(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := strings.Split(msg.Text, "_")
	if len(args) < 2 {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Invalid command format. Usage: /process_<id>"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	id := args[1]

	// Create a button for /process_start
	buttonStart := tgbotapi.NewInlineKeyboardButtonData("Start", fmt.Sprintf("process_start_%s", id))
	// Create a button for /process_stop
	buttonStop := tgbotapi.NewInlineKeyboardButtonData("Stop", fmt.Sprintf("process_stop_%s", id))
	// Create a button for /process_restart
	buttonRestart := tgbotapi.NewInlineKeyboardButtonData("Restart", fmt.Sprintf("process_restart_%s", id))
	row := tgbotapi.NewInlineKeyboardRow(buttonStart, buttonStop, buttonRestart)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

	// Send the button
	message := tgbotapi.NewMessage(msg.Chat.ID, "Choose an action:")
	message.ReplyMarkup = keyboard

	_, err := bot.Send(message)
	if err != nil {
		log.Println("Failed to send button: ", err)
	}
}

func handleStartProcess(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	// Create a button for /list_processes
	button := tgbotapi.NewInlineKeyboardButtonData("List Processes", "list_processes")
	row := tgbotapi.NewInlineKeyboardRow(button)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

	// Send the button
	message := tgbotapi.NewMessage(msg.Chat.ID, "Choose an action:")
	message.ReplyMarkup = keyboard

	_, err := bot.Send(message)
	if err != nil {
		log.Println("Failed to send button: ", err)
	}
}

func handleProcessCommand(bot *tgbotapi.BotAPI, callBack *tgbotapi.CallbackQuery, action string) {
	args := strings.Split(callBack.Data, "_")
	if len(args) != 3 {
		_, err := bot.Send(tgbotapi.NewMessage(callBack.Message.Chat.ID, "Usage: button '"+action+"_process_<id>'"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	id := args[2]

	log.Println("Process command: ", action, id)
	pr, err := astraapi.ProcessAction(action, id)
	if err != nil {
		_, err = bot.Send(tgbotapi.NewMessage(callBack.Message.Chat.ID, fmt.Sprintf("Failed to %s process: %s", action, err)))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	bot.Send(tgbotapi.NewMessage(callBack.Message.Chat.ID, pr.Message))
}

func handleListProcesses(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	processes, err := astraapi.GetProcessStarus()
	if err != nil {
		_, err = bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Failed to fetch processes: %s", err)))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
	}

	var response string
	for _, process := range *processes {
		flag := "🔴"
		if process.IsActive {
			flag = "🟢"
		}

		response += fmt.Sprintf("📺: %s, %s /process_%d\n", process.Name, flag, process.ID)
	}

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))

	handleStartProcess(bot, msg)
}
