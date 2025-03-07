package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const cConfigFilenameUsage = "config filename"

func loadConfigFile[T ConfigClient | ConfigServer](config *T) error {
	configFile, err := getFileConfig()
	if err != nil {
		return err
	}
	if configFile != "" {
		err = parseConfig(configFile, config)
		if err != nil {
			return err
		}
	}
	return nil
}

func getFileConfig() (string, error) {
	filename := ""
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-c" {
			// don't have next arg?
			if i+2 > len(os.Args) {
				return "", fmt.Errorf("config file name not found for arg %s", os.Args[i])
			}
			filename = os.Args[i+1]
			break
		} else if strings.HasPrefix(os.Args[i], "-c=") {
			s := strings.Split(os.Args[i], "=")
			if len(s) != 2 {
				return "", fmt.Errorf("config file name not found for arg %s", os.Args[i])
			}
			filename = s[1]
		}
	}

	if c, ok := os.LookupEnv("CONFIG"); ok {
		filename = c
	}

	return filename, nil
}

func readFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}

func parseConfig[T ConfigClient | ConfigServer](filename string, conf *T) error {
	data, err := readFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, conf)
	if err != nil {
		return fmt.Errorf("error parsing file config: %w", err)
	}

	return nil
}
