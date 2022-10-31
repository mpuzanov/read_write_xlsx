package services

import (
	"read_write_xlsx/internal/config"
	"read_write_xlsx/internal/glogger"
	"read_write_xlsx/pkg/xlsx"
)

// App ...
type App struct {
	cfg config.Config
	log glogger.Logger
}

// New ...
func New(cfg config.Config, logger glogger.Logger) *App {
	return &App{cfg: cfg, log: logger}
}

// Run ...
func (app *App) Run(filename string) error {

	app.log.Infof("Обрабатываем файл %v", filename)

	fileExcelRead := xlsx.NewFieldsExcel("", app.cfg.ReadFileSettings, app.log)
	app.log.Debugf("fileExcelRead: %v", fileExcelRead)

	data, err := fileExcelRead.ExcelToData(filename, 2)
	if err != nil {
		return err
	}

	// Покажем несколько записей для примера
	for i := 0; i < 5; i++ {
		app.log.Debug(i, data[i])
	}

	// Создадим в файле новый лист
	fileExcelWrite := xlsx.NewFieldsExcel("Вывод", app.cfg.WriteFileSettings, app.log)
	app.log.Debugf("fileExcelWrite: %v", fileExcelWrite)

	if err := fileExcelWrite.DataToExcel(filename, 1, data); err != nil {
		return err
	}

	app.log.Info("Выполнено")

	return nil
}
