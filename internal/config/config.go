package config

import (
	"encoding/json"
	"os"
	"read_write_xlsx/pkg/xlsx"
)

var (
	// Version для установки версии релиза
	Version = "development"
)

// ShowVersion Вывод версии релиза программы
func ShowVersion() string {
	return Version
}

// Config ...
type Config struct {
	LogLevel string `json:"log_level"`

	ReadFileSettings struct {
		SheetName string                  `json:"sheet_name,omitempty"`
		StartRow  int                     `json:"start_row,omitempty" env:"START_ROW" env-default:"2"`
		Fields    map[int]xlsx.FieldExcel `json:"fields"`
	} `json:"read_file_settings"`

	WriteFileSettings map[int]xlsx.FieldExcel `json:"write_file_settings"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var c Config
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
