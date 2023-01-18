package services

import (
	"fmt"
	"read_write_xlsx/internal/config"
	"read_write_xlsx/internal/glogger"
	"read_write_xlsx/pkg/xlsx"

	"github.com/xuri/excelize/v2"
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

	fileExcelRead := xlsx.NewFieldsExcel(app.cfg.ReadFileSettings.SheetName, app.cfg.ReadFileSettings.Fields, app.log)
	app.log.Debugf("fileExcelRead: %v", fileExcelRead)
	data, err := fileExcelRead.ExcelToData(filename, app.cfg.ReadFileSettings.StartRow)
	if err != nil {
		return err
	}

	// Покажем несколько записей для примера
	for i := 0; i < 5; i++ {
		app.log.Debug(i, data[i])
	}

	// Создадим в файле новый лист
	sheetNameData := "Вывод"
	fileExcelWrite := xlsx.NewFieldsExcel(sheetNameData, app.cfg.WriteFileSettings, app.log)
	app.log.Debugf("fileExcelWrite: %v", fileExcelWrite)

	if err := fileExcelWrite.DataToExcel(filename, 1, data); err != nil {
		return err
	}

	app.log.Debug("Создадим в файле новый лист для сводной информации")
	sheetNamePivot := "Свод по платежам"
	letterLastColumn := "E"
	dataRange := fmt.Sprintf("%s!$A$%d:$%s$%d", sheetNameData, 1, letterLastColumn, len(data)+1)
	pivotTableRange := fmt.Sprintf("%s!$B$5:$E$20", sheetNamePivot)

	app.log.Debugf("dataRange: %v, pivotTableRange: %v", dataRange, pivotTableRange)

	if err := fileExcelWrite.CreatePivotTableFile(filename, sheetNamePivot,
		dataRange,
		pivotTableRange,
		[]excelize.PivotTableField{ //PivotTableRows
			{Data: "Дата платежа", Name: "Дата платежа"}, {Data: "Лицевой счет"},
		},
		[]excelize.PivotTableField{ //PivotTableFilter
		},
		[]excelize.PivotTableField{ //PivotTableColumns
		},
		[]excelize.PivotTableField{ //PivotTableData
			{Data: "Сумма платежа", Name: "Сумма платежа", Subtotal: "Sum"},
			{Data: "Лицевой счет", Name: "Количество", Subtotal: "Count"},
		},
	); err != nil {
		return err
	}

	app.log.Info("Выполнено")

	return nil
}
