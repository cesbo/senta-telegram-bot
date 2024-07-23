package config

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Token string `json:"token"`
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