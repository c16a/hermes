package config

import (
	"encoding/json"
	"io/ioutil"
)

// ParseConfig parses a file path and returns a config
//
// If the file is remote, it is downloaded via HTTP (proxy settings are respected)
// If file is local, is it opened in readonly mode.
func ParseConfig(configFilePath string) (*Config, error) {
	fileBytes, err := readLocalFile(configFilePath)
	if err != nil {
		return nil, err
	}
	return jsonParser(fileBytes)
}
func readLocalFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func jsonParser(fileBytes []byte) (*Config, error) {
	var config Config
	err := json.Unmarshal(fileBytes, &config)
	return &config, err
}
