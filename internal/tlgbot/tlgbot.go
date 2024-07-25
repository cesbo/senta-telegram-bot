package tlgbot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sentabot/internal/config"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const apiUrlSuffix = "rest/api/v1"

type ProcessResponse struct {
	Message string `json:"message"`
}

func Pool() error {
	token := config.GetConfig().TlgToken
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

		if update.CallbackQuery != nil {
			switch update.CallbackQuery.Data {
			case "list_processes":
				handleListProcesses(bot, update.CallbackQuery.Message)
			}

			log.Println("CallbackQuery: ", update.CallbackQuery.Data)

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
	apiUrl := config.GetConfig().Server

	args := strings.Split(callBack.Data, "_")
	if len(args) != 3 {
		_, err := bot.Send(tgbotapi.NewMessage(callBack.Message.Chat.ID, "Usage: /"+action+"_process <id>"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	id := args[2]

	log.Println("Process command: ", action, id)
	url := fmt.Sprintf("%s/%s/process/%s/%s", apiUrl, apiUrlSuffix, id, action)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(callBack.Message.Chat.ID, "Failed to create request"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	setToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(callBack.Message.Chat.ID, "Failed to execute request"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}
	defer resp.Body.Close()

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		_, err := bot.Send(tgbotapi.NewMessage(callBack.Message.Chat.ID, "Request failed with status: "+resp.Status))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	var processResp ProcessResponse
	if err := json.NewDecoder(resp.Body).Decode(&processResp); err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(callBack.Message.Chat.ID, "Failed to decode response"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	bot.Send(tgbotapi.NewMessage(callBack.Message.Chat.ID, processResp.Message))
}

func handleListProcesses(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	apiUrl := config.GetConfig().Server

	url := fmt.Sprintf("%s/%s/process/list/status", apiUrl, apiUrlSuffix)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to create request"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	setToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to fetch processes"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}
	defer resp.Body.Close()

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Request failed with status: "+resp.Status))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	var processes []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&processes); err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to decode response"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	var response string
	for _, process := range processes {
		flag := "🔴"
		if process["is_active"].(bool) {
			flag = "🟢"
		}
		response += fmt.Sprintf("📺: %s, %s /process_%d\n", process["name"].(string), flag, int(process["id"].(float64)))
	}

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))

	handleStartProcess(bot, msg)
}

func setToken(req *http.Request) {
	token := config.GetConfig().APIToken
	req.Header.Set("accept", "application/json")
	req.Header.Set("api_key", token)
}
