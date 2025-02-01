package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
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

func lookupEnvDuration(env string, val *Duration) error {
	if v, ok := os.LookupEnv(env); ok {
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing env param %s: %w", env, err)
		}
		*val = Duration{time.Duration(i) * time.Second}
	}
	return nil
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String()) //nolint:wrapcheck // all ok
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err //nolint:wrapcheck // all ok
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err //nolint:wrapcheck // all ok
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}
