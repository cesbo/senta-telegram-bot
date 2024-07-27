package astraapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sentabot/internal/config"
)

type ProcessResponse struct {
	Message string `json:"message"`
}

type ProcessStatus struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

const apiUrlSuffix = "rest/api/v1"

func ProcessAction(action string, id string) (pr *ProcessResponse, err error) {
	url := fmt.Sprintf("%s/process/%s/%s", getApiUrl(), id, action)

	req, err := http.NewRequest("POST", url, nil)
	log.Println("Request: ", req)
	if err != nil {
		log.Println("Failed to create request: ", err)
		return
	}

	setToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	log.Println("Response: ", resp, err)
	if err != nil {
		log.Println("Failed to send request: ", err)
		return
	}
	defer resp.Body.Close()

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		var body []byte
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read response body: ", err)
			return
		}

		err = errors.New(string(body))
		return
	}

	if err = json.NewDecoder(resp.Body).Decode(pr); err != nil {
		log.Println("Failed to decode response: ", err)
		err = errors.New("Failed to decode response: " + err.Error())
		return
	}

	return
}

func GetProcessStarus() (ps *[]ProcessStatus, err error) {
	url := fmt.Sprintf("%s/process/list/status", getApiUrl())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Failed to create request: ", err)
		return
	}

	setToken(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to send request: ", err)
		return
	}
	defer resp.Body.Close()

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		err = errors.New("Request failed with status: " + resp.Status)
		log.Println("Request failed with status: ", resp.Status)
		return
	}

	if err = json.NewDecoder(resp.Body).Decode(ps); err != nil {
		log.Println("Failed to decode response: ", err)
		err = errors.New("Failed to decode response: " + err.Error())
		return
	}

	return
}

func getApiUrl() string {
	return config.GetConfig().Server + "/" + apiUrlSuffix
}

func setToken(req *http.Request) {
	token := config.GetConfig().APIToken
	req.Header.Set("accept", "application/json")
	req.Header.Set("api_key", token)
}
