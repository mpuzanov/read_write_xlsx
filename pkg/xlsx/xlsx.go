/*
Package xlsx используется для импорта или записи данных Excel-файла

документация к библиотеке
https://xuri.me/excelize/ru/

Пример из конфига для чтения и записи файла:

	"read_file_settings" : {
		"2": {"name":"fio", "header":"Ф.И.О."},
		"3": {"name":"data_paym", "header":"Дата платежа", "type":"date", "format":"dd.mm.yyyy"},
		"4": {"name":"account", "header":"Лицевой счет", "type":"int64"},
		"5": {"name":"address", "header":"Адрес"},
		"6": {"name":"paym_account", "header":"Сумма платежа", "type":"float64", "format":"#,##0.00"}
	},

	"write_file_settings": {
		"1": {"name":"fio", "header":"Ф.И.О.", "width":40},
		"2": {"name":"data_paym", "header":"Дата платежа", "type":"date"},
		"3": {"name":"account", "header":"Лицевой счет", "type":"int64"},
		"5": {"name":"paym_account", "header":"Сумма платежа", "type":"float64", "format":"#,##0.00"}
	}
*/
package xlsx

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Logger для использования в данном модуле
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// FieldExcel структура для колонки excel-файла
type FieldExcel struct {
	Name        string  `json:"name"`             // имя поля, имя колонки в базе данных (на ENG)
	Header      string  `json:"header,omitempty"` // заголовок колонки
	Width       float64 `json:"width,omitempty"`  // ширина колонки
	Format      string  `json:"format,omitempty"` // формат вывода
	Type        string  `json:"type,omitempty"`   // типы данных  int64, float64, date
	ParseFormat string  `json:"parse,omitempty"`  // формат для разбора входных значений
	StyleID     int     `json:"style_id"`         // код стиля в файле (служебное поле, используется для вывода)
}

// FieldsExcel структура для описания массива колонок excel-файла
type FieldsExcel struct {
	sheetName string
	fields    map[int]FieldExcel // int - используется как номер колонки (при импорте или выводе)
	log       Logger
}

// NewFieldsExcel подготавливаем окончательно структуру для работы
func NewFieldsExcel(sheetName string, fields map[int]FieldExcel, log Logger) FieldsExcel {
	for k, v := range fields {

		if v.Width == 0 { // если ширина не задана
			v.Width = float64(len([]rune(v.Header)) + 5) // берём длину заголовка + 5
			fields[k] = v
		}

		if v.Type == "date" && v.ParseFormat == "" { // если формат разбора даты не задан
			if v.Format != "" {
				tmp := strings.ToLower(v.Format)
				tmp = strings.ReplaceAll(tmp, "dd", "02")
				tmp = strings.ReplaceAll(tmp, "d", "2")
				tmp = strings.ReplaceAll(tmp, "mm", "01")
				tmp = strings.ReplaceAll(tmp, "m", "1")
				tmp = strings.ReplaceAll(tmp, "yyyy", "2006")
				tmp = strings.ReplaceAll(tmp, "yy", "06")
				v.ParseFormat = tmp
			} else {
				v.ParseFormat = "02.01.2006"
				v.Format = "dd.mm.yyyy"
			}
			fields[k] = v
		}
	}
	return FieldsExcel{sheetName: sheetName, fields: fields, log: log}
}

// CountColumn кол-во колонок
func (s FieldsExcel) CountColumn() int {
	return len(s.fields)
}

// MaxColumn максимальная колонка  (может не совпадать с кол-вом)
func (s FieldsExcel) MaxColumn() int {
	max := 0
	for key := range s.fields {
		if key > max {
			max = key
		}
	}
	return max
}

// randomRange случайное число между min и max
func randomRange(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

// NewFromJSON разбор структуры колонок
func NewFromJSON(sheetName string, text string, log Logger) (*FieldsExcel, error) {
	var err error
	var data map[int]FieldExcel

	fe := FieldsExcel{log: log, fields: data}
	if err = json.Unmarshal([]byte(text), &data); err != nil {
		return nil, fmt.Errorf("json.Unmarshal : %w", err)
	}

	fe = NewFieldsExcel(sheetName, data, log)

	return &fe, nil
}

// NewFromModelTags создание структуры описания колонок вывода Excel-файла
// из тегов модели получаем нужные данные
func NewFromModelTags(f interface{}, log Logger) (FieldsExcel, error) {
	var err error
	fe := FieldsExcel{log: log, fields: make(map[int]FieldExcel)}

	ft := reflect.TypeOf(f)
	for i := 0; i < ft.NumField(); i++ {
		curField := ft.Field(i)
		//log.Debugf("%v %v %v %v %v", ft.Name(), curField.Name, curField.Type.Name(), curField.Tag.Get("header"), curField.Tag.Get("format"))

		width := 10.0
		tmp := strings.TrimSpace(curField.Tag.Get("width"))
		if tmp != "" {
			width, err = strconv.ParseFloat(tmp, 64)
			if err != nil {
				return fe, fmt.Errorf("width=<%v> name=<%v> : %w", tmp, curField.Name, err)
			}
		}

		fe.fields[i+1] = FieldExcel{
			Name:   curField.Tag.Get("db"),
			Header: curField.Tag.Get("header"),
			Width:  width,
			Format: curField.Tag.Get("format"),
		}
	}
	return fe, nil
}

// CreateStyle Создаёт стили в файле на основе заданных форматов полей.
func (s FieldsExcel) CreateStyle(f *excelize.File) error {
	var err error
	var styles = make(map[string]int)
	for key, v := range s.fields {
		if v.Format != "" {
			style, ok := styles[v.Format]
			if ok {
				v.StyleID = style
			} else {
				switch v.Format { // https://xuri.me/excelize/ru/style.html#number_format
				case "#,##0": // для целых чисел с разделителями тысяч
					if style, err = f.NewStyle(&excelize.Style{NumFmt: 3, Lang: "ru-ru"}); err != nil {
						return err
					}
				case "#,##0.00": // для денег с разделителями тысяч
					if style, err = f.NewStyle(&excelize.Style{NumFmt: 4, Lang: "ru-ru"}); err != nil {
						return err
					}
				case "dd.mm.yyyy":
					if style, err = f.NewStyle(&excelize.Style{NumFmt: 14, Lang: "ru-ru"}); err != nil {
						return err
					}
				default:
					if style, err = f.NewStyle(&excelize.Style{CustomNumFmt: &v.Format, Lang: "ru-ru"}); err != nil {
						return err
					}
				}
				s.log.Debugf("Добавлен стиль в файл. %v=%v", style, v.Format)
				styles[v.Format] = style
				v.StyleID = styles[v.Format]
			}
			s.fields[key] = v
		}
	}
	return nil
}
