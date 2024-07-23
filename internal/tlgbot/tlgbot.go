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

	return nil
}

func handleProcessCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, action string) {
	apiUrl := config.GetConfig().Server
	apiToken := config.GetConfig().APIToken

	args := strings.Split(msg.Text, " ")
	if len(args) != 2 {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Usage: /"+action+"_process <id>"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	id := args[1]
	url := fmt.Sprintf("%s/%s/process/%s/%s", apiUrl, apiUrlSuffix, id, action)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to create request"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api_key", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to execute request"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Request failed with status: "+resp.Status))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	var processResp ProcessResponse
	if err := json.NewDecoder(resp.Body).Decode(&processResp); err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to decode response"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, processResp.Message))
}

func handleListProcesses(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	apiUrl := config.GetConfig().Server
	apiToken := config.GetConfig().APIToken

	url := fmt.Sprintf("%s/%s/process/list/status", apiUrl, apiUrlSuffix)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to create request"))
		if err != nil {
			log.Println("Failed to send message ", err)
		}
		return
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api_key", apiToken)

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

	if resp.StatusCode != http.StatusOK {
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
		response += fmt.Sprintf("ID: %d, Name: %s, Active: %t\n", int(process["id"].(float64)), process["name"].(string), process["is_active"].(bool))
	}

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
}
