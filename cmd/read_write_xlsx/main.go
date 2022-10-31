// Пустой шаблон проекта на Golang
package main

import (
	"fmt"
	"log"
	"os"
	"read_write_xlsx/internal/config"
	"read_write_xlsx/internal/services"

	"read_write_xlsx/internal/glogger"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Файл не задан")
		fmt.Println("Используйте: " + os.Args[0] + " <Имя файла>")
		os.Exit(1)
	}
	filename := os.Args[1]

	cfgPath := "config.json"
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	logger := glogger.BuildLogger("", cfg.LogLevel) // STD LOGRUS ZAP

	logger.Debugf("%v", cfg)

	s := services.New(cfg, logger)
	if err := s.Run(filename); err != nil {
		logger.Fatal("services.Run:", err)
	}

}
