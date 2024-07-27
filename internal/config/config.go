package config

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	TlgToken      string   `json:"tlg_token"`
	Server        string   `json:"server"`
	APIToken      string   `json:"api_token"`
	AcceptedUsers []string `json:"accepted_users"`
}

var config Config

func LoadConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &config)
	return err
}

func GetConfig() Config {
	return config
}
