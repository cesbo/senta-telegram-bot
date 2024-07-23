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

func Pool() {
	token := config.GetConfig().Token
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if strings.HasPrefix(update.Message.Text, "/start_process") {
			handleProcessCommand(bot, update.Message, "start")
		} else if strings.HasPrefix(update.Message.Text, "/stop_process") {
			handleProcessCommand(bot, update.Message, "stop")
		} else if strings.HasPrefix(update.Message.Text, "/restart_process") {
			handleProcessCommand(bot, update.Message, "restart")
		} else if update.Message.Text == "/list_processes" {
			handleListProcesses(bot, update.Message)
		}
	}
}

func handleProcessCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, action string) {
	apiUrl := config.GetConfig().Server

	args := strings.Split(msg.Text, " ")
	if len(args) != 2 {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Usage: /"+action+"_process <id>"))
		return
	}

	id := args[1]
	url := fmt.Sprintf("%s/%s/process/%s/%s", apiUrl, apiUrlSuffix, id, action)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to create request"))
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to execute request"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Request failed with status: "+resp.Status))
		return
	}

	var processResp ProcessResponse
	if err := json.NewDecoder(resp.Body).Decode(&processResp); err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to decode response"))
		return
	}

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, processResp.Message))
}

func handleListProcesses(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	apiUrl := config.GetConfig().Server
	url := fmt.Sprintf("%s/%s/process/list/status", apiUrl, apiUrlSuffix)

	resp, err := http.Get(url)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to fetch processes"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Request failed with status: "+resp.Status))
		return
	}

	var processes []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&processes); err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to decode response"))
		return
	}

	var response string
	for _, process := range processes {
		response += fmt.Sprintf("ID: %d, Name: %s, Active: %t\n", int(process["id"].(float64)), process["name"].(string), process["is_active"].(bool))
	}

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
}
