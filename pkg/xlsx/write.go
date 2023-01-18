package xlsx

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

// DataToExcel Записываем данные в Excel-файл
func (s *FieldsExcel) DataToExcel(filename string, startRow int, data []map[string]interface{}) error {
	var (
		err error
		f   *excelize.File
	)
	if startRow == 0 {
		startRow = 1
	}
	if s.sheetName == "" {
		s.sheetName = "Новый лист"
	}
	if _, err = os.Stat(filename); os.IsNotExist(err) {
		// файл не существует
		f = excelize.NewFile()
		f.SetSheetName("Sheet1", s.sheetName) // лист по умолчанию переименовываем
		s.log.Debugf("NewFile sheetName=%v", s.sheetName)
	} else {
		// файл существует
		f, err = excelize.OpenFile(filename)
		if err != nil {
			return fmt.Errorf("OpenFile %v", err)
		}
		if index, _ := f.GetSheetIndex(s.sheetName); index != -1 {
			f.DeleteSheet(s.sheetName)
		}
		f.NewSheet(s.sheetName)
		s.log.Debugf("OpenFile sheetName=%v", s.sheetName)
	}
	defer func() {
		if err := f.Close(); err != nil {
			s.log.Error(err)
		}
	}()
	// создание streamWriter для буферизированной записи
	streamWriter, err := f.NewStreamWriter(s.sheetName)
	if err != nil {
		return fmt.Errorf("NewStreamWriter %v", err)
	}

	headersColumns := s.fields
	if err = s.CreateStyle(f); err != nil {
		return err
	}
	countColumn := s.CountColumn()
	s.log.Debugf("countColumn=%v", countColumn)
	maxColumn := s.MaxColumn()
	s.log.Debugf("maxColumn=%v", maxColumn)

	s.log.Debug("Устанавливаем ширину столбцов")
	for i, v := range headersColumns {
		if v.Width > 0 {
			if err = streamWriter.SetColWidth(i, i, v.Width); err != nil {
				return fmt.Errorf("SetColWidth %v", err)
			}
		}
	}

	s.log.Debug("формируем строку заголовка")
	strColumns := make([]interface{}, maxColumn)
	for i := 0; i < maxColumn; i++ {
		if v, ok := headersColumns[i+1]; ok {
			strColumns[i] = v.Header
		} else {
			strColumns[i] = "-"
		}
	}
	s.log.Debugf("headersName=%v", strColumns)

	// формируем адрес первой ячейки для записи
	addrStart, _ := excelize.JoinCellName("A", startRow)
	s.log.Debugf("addrStart=%v", addrStart)

	// пишем строку заголовка
	if err := streamWriter.SetRow(addrStart, strColumns); err != nil {
		return fmt.Errorf("SetRow addrStart %v", err)
	}

	startData := startRow + 1 // шапка + заголовок
	countData := len(data)
	s.log.Debugf("countData=%v", countData)
	// Пишем данные
	for r, row := range data {

		// формируем строку данных заданного формата
		rowVal := make([]interface{}, maxColumn)
		for i := 0; i < maxColumn; i++ {

			if v, ok := headersColumns[i+1]; ok {
				val := row[v.Name]

				if v.Type == "float64" {
					if _, ok := val.(float64); !ok {
						res, err := strconv.ParseFloat(val.(string), 64)
						if err != nil {
							return err
						}
						val = res
					}
				}
				if v.Type == "date" {
					if _, ok := val.(time.Time); !ok {
						valStr := val.(string)
						t, err := time.Parse(v.ParseFormat, valStr)
						if err != nil {
							s.log.Debugf("time.Parse ParseFormat=%v val=%v error: %v", v.ParseFormat, valStr, err)
							resFloat, err := strconv.ParseFloat(valStr, 64)
							if err != nil {
								return err
							}
							res, err := excelize.ExcelDateToTime(resFloat, false)
							if err != nil {
								return err
							}
							val = res
						} else {
							val = t
						}
					}
				}
				if v.Type == "int64" {
					if _, ok := val.(int64); !ok {
						res, err := strconv.ParseInt(val.(string), 10, 64)
						if err != nil {
							return err
						}
						val = res
					}
				}
				rowVal[i] = excelize.Cell{StyleID: v.StyleID, Value: val}
			} else {
				rowVal[i] = nil
			}
		}
		// пишем строку данных
		addr, _ := excelize.CoordinatesToCellName(1, r+startData)
		if err := streamWriter.SetRow(addr, rowVal); err != nil {
			return fmt.Errorf("SetRow %v", err)
		}

		if r%1000 == 0 {
			fmt.Printf("Сохраняем данные в файл...%v\r", r)
		}
	}
	//============= создаём умную таблицу
	letterLastColumn, _ := excelize.ColumnNumberToName(maxColumn)
	rangeRef := fmt.Sprintf("%s:%s%d", addrStart, letterLastColumn, countData+startRow)
	tableName := "table" + strconv.Itoa(randomRange(1, 100))
	s.log.Debugf("rangeRef=%s, table_name=%s", rangeRef, tableName)

	disable := false
	if err := streamWriter.AddTable(rangeRef, &excelize.TableOptions{
		Name:              tableName,
		StyleName:         "TableStyleMedium2",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowRowStripes:    &disable,
		ShowColumnStripes: false,
	}); err != nil {
		return fmt.Errorf("AddTable %v %w", tableName, err)
	}

	//==========================================
	f.SetActiveSheet(0)

	if err := streamWriter.Flush(); err != nil {
		return fmt.Errorf("flush %v", err)
	}
	if err := f.SaveAs(filename); err != nil {
		return err
	}
	return nil
}

// CreatePivotTableFile ...
func (s *FieldsExcel) CreatePivotTableFile(filename, sheetNamePivot string,
	DataRange, PivotTableRange string,
	PivotTableRows []excelize.PivotTableField,
	PivotTableFilter []excelize.PivotTableField,
	PivotTableColumns []excelize.PivotTableField,
	PivotTableData []excelize.PivotTableField,
) error {
	//==========================================
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return fmt.Errorf("OpenFile %v", err)
	}
	defer f.Close()

	if index, _ := f.GetSheetIndex(sheetNamePivot); index != -1 {
		f.DeleteSheet(sheetNamePivot)
	}
	f.NewSheet(sheetNamePivot)
	if err := f.AddPivotTable(&excelize.PivotTableOptions{
		DataRange:       DataRange,
		PivotTableRange: PivotTableRange,
		Rows:            PivotTableRows,
		Filter:          PivotTableFilter,
		Columns:         PivotTableColumns,
		Data:            PivotTableData,
		RowGrandTotals:  true,
		ColGrandTotals:  true,
		ShowDrill:       true,
		ShowRowHeaders:  true,
		ShowColHeaders:  true,
		ShowLastColumn:  true,

		UseAutoFormatting: true,
		PageOverThenDown:  true,
		//MergeItem:         true,
		CompactData: false,
		//ShowError:         true,
	}); err != nil {
		return err
	}
	f.SetActiveSheet(0)
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("SaveAs %v", err)
	}

	return nil
}
