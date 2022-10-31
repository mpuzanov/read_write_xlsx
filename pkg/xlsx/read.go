package xlsx

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// ExcelToData чтение Excel-файла
func (s *FieldsExcel) ExcelToData(filename string, startData int) ([]map[string]interface{}, error) {

	s.log.Debug("Читаем файл: ", filename)
	fmt.Print("Открываем файл...\r")

	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("OpenFile %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			s.log.Error("Close file")
		}
	}()
	if s.sheetName == "" {
		s.sheetName = f.GetSheetName(0) // получаем имя 1 листа
	}
	s.log.Debug("Лист: ", s.sheetName)

	i := 0
	data := make([]map[string]interface{}, 0)

	rows, err := f.Rows(s.sheetName)
	if err != nil {
		return nil, fmt.Errorf("f.Rows %v", err)
	}
	for rows.Next() {

		row, err := rows.Columns(excelize.Options{RawCellValue: true})
		if err != nil {
			return nil, fmt.Errorf("rows.Columns %v", err)
		}

		if len(row) == 0 { // пропускаем пустую строку
			continue
		}

		i++
		if i < startData { // пропускаем шапку таблицы
			continue
		}
		var val interface{}
		dt := make(map[string]interface{}, len(s.fields))
		for key, v := range s.fields {
			val = row[key-1]

			if v.Type == "date" {
				resFloat, err := strconv.ParseFloat(val.(string), 64)
				if err != nil {
					return nil, err
				}
				res, err := excelize.ExcelDateToTime(resFloat, false)
				if err != nil {
					return nil, err
				}
				val = res.Format(v.ParseFormat)
			}
			if v.Type == "float64" {
				res, err := strconv.ParseFloat(val.(string), 64)
				if err != nil {
					return nil, err
				}
				val = res
			}

			dt[s.fields[key].Name] = val
		}
		data = append(data, dt)

		if i%1000 == 0 {
			fmt.Printf("Идет чтение строк файла...%v\r", i)
		}
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}

	s.log.Debug("Прочитано: ", len(data))

	return data, nil
}
