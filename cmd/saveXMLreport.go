package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/e0m-ru/caldavreport/report"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
	"github.com/xuri/excelize/v2"
)

const timeFormat = "20060102T150405"

type calendarData struct {
	name    string
	objList *[]caldav.CalendarObject
	err     error
}

type Category struct {
	Tag  string `json:"tag"`
	Name string `json:"name"`
}

var (
	categories = []Category{
		{"SOUND", "звук"},
		{"VIDEO", "видео"},
		{"PHOTO", "фото"},
		{"TRANS", "эфир"},
		{"VKS", "ВКС"},
		{"TV", "ТВ"},
		{"SYNCH", "синхрон"},
	}
	NOW     = time.Now()
	logFile *os.File
	appID   = "CALDAVREPORT"
)

func main() {

	month := flag.Int("m", int(NOW.Month()), "Укажите месяц")
	year := flag.Int("y", NOW.Year(), "Укажите год")
	logfileName := flag.String("f", "STDOUT", "Файл логирования")
	flag.Parse()

	logger, err := LogInit(*logfileName)
	if err != nil {
		log.Fatal(err)
	}
	defer LogClose()

	start := time.Date(*year, time.Month(*month), 1, 0, 0, 0, 0, NOW.Location())
	end := start.AddDate(0, 1, 0)

	R, err := report.NewDateRangeReport(start, end)
	if err != nil {
		logger.Fatal(err)
	}

	var wg sync.WaitGroup
	var out = make(chan calendarData, len(R.Calendars))

	for _, c := range R.Calendars {
		if c.Name == "Отпуска/Больничные" {
			continue
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			calendarObjects, err := R.QueryCalendarData(c)
			out <- calendarData{c.Name, &calendarObjects, err}
			wg.Done()
		}(&wg)
	}

	wg.Wait()
	close(out)

	for v := range out {
		if v.err != nil {
			logger.Fatal(v.err)
		}
		R.Reports[v.name] = v.objList
	}

	R.ParseWorks()

	err = saveExcel(R)
	if err != nil {
		logger.Fatal(err)
	}
}

func saveExcel(R report.DateRangeReport) error {
	f := excelize.NewFile()

	h1, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 24,
		},
	})
	border, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 14,
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",           // Важно указать "pattern" вместо "fill"
			Color:   []string{"#dcdcdc"}, // HEX цвет с #
			Pattern: 1,                   // 1 - solid fill
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 3},
			{Type: "top", Color: "000000", Style: 3},
			{Type: "bottom", Color: "000000", Style: 3},
			{Type: "right", Color: "000000", Style: 3},
		},
	})
	wrapStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText: true,  // Включаем перенос текста
			Vertical: "top", // Выравнивание по верхнему краю
		},
	})
	wrapBorder, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText: true,  // Включаем перенос текста
			Vertical: "top", // Выравнивание по верхнему краю
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	f.SetColWidth("Sheet1", "C", "C", 50)

	for letter := 'A'; letter <= 'J'; letter++ {
		f.SetColStyle("Sheet1", string(letter), wrapStyle)
	}

	// Set value on Sheet1
	f.SetCellValue("Sheet1", "A1", "Отчёт Пресс-центра "+R.TimeRange.Start.Format("01.2006"))
	f.SetCellStyle("Sheet1", "A1", "A1", h1)

	f.SetCellValue("Sheet1", "A2", "Дата")
	f.SetCellValue("Sheet1", "B2", "Место")
	f.SetCellValue("Sheet1", "C2", "Описание")
	f.SetCellValue("Sheet1", "D2", "звук")
	f.SetCellValue("Sheet1", "E2", "видео")
	f.SetCellValue("Sheet1", "F2", "фото")
	f.SetCellValue("Sheet1", "G2", "трансляция")
	f.SetCellValue("Sheet1", "H2", "ВКС")
	f.SetCellValue("Sheet1", "I2", "ТВ")
	f.SetCellValue("Sheet1", "J2", "Синхрон")

	f.SetCellStyle("Sheet1", "A2", "J2", border)

	rows := ParseReport(R)

	cellName := ""
	for i, row := range rows {
		for j, cellValue := range row {
			cellName, _ = excelize.CoordinatesToCellName(j+1, i+3)
			if cellValue == "" {
				continue
			}
			f.SetCellValue("Sheet1", cellName, cellValue)
		}
	}

	_, y, _ := excelize.CellNameToCoordinates(cellName)
	f.SetCellStyle("Sheet1", "A3", "J"+fmt.Sprint(y), wrapBorder)
	y += 1
	f.SetCellFormula("Sheet1", "D"+fmt.Sprint(y), fmt.Sprintf("=COUNTIFS(D3:D%v, \"<>\")", y-1))
	f.SetCellFormula("Sheet1", "E"+fmt.Sprint(y), fmt.Sprintf("=COUNTIFS(E3:E%v, \"<>\")", y-1))
	f.SetCellFormula("Sheet1", "F"+fmt.Sprint(y), fmt.Sprintf("=COUNTIFS(F3:F%v, \"<>\")", y-1))
	f.SetCellFormula("Sheet1", "G"+fmt.Sprint(y), fmt.Sprintf("=COUNTIFS(G3:G%v, \"<>\")", y-1))
	f.SetCellFormula("Sheet1", "H"+fmt.Sprint(y), fmt.Sprintf("=COUNTIFS(H3:H%v, \"<>\")", y-1))
	f.SetCellFormula("Sheet1", "I"+fmt.Sprint(y), fmt.Sprintf("=COUNTIFS(I3:I%v, \"<>\")", y-1))
	f.SetCellFormula("Sheet1", "J"+fmt.Sprint(y), fmt.Sprintf("=COUNTIFS(J3:J%v, \"<>\")", y-1))
	f.SetCellStyle("Sheet1", "D"+fmt.Sprint(y), "J"+fmt.Sprint(y), border)

	// char
	if err := f.AddChart("Sheet1", "K2", &excelize.Chart{
		Type: excelize.Pie,
		Series: []excelize.ChartSeries{
			{
				Name:       "Sheet1!$D$2",
				Categories: "Sheet1!$D$2:$J$2", //Sheet1!$A$1:$C$1
				Values:     "Sheet1!$D$" + fmt.Sprint(y) + ":$J$" + fmt.Sprint(y),
				DataLabel:  excelize.ChartDataLabel{},
			}},
		Title: []excelize.RichTextRun{
			{
				Text: "Доля по видам работ",
			},
		},
		PlotArea: excelize.ChartPlotArea{
			ShowPercent: true,
			ShowCatName: true,
		},
	}); err != nil {
		return err
	}

	// Save spreadsheet
	if err := f.SaveAs(fmt.Sprintf("./report_%02d.xlsx", R.TimeRange.Start.Month())); err != nil {
		return err
	}
	return nil
}

// Logger initiation. Default os.Stdout
func LogInit(logFileName string) (logger *log.Logger, err error) {
	var out io.Writer
	if logFileName != "STDOUT" {
		logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return &log.Logger{}, err
		}
		out = logFile
	} else {
		out = os.Stdout
	}
	logger = log.New(out, appID+": ", log.Ldate|log.Ltime|log.Lshortfile)
	return
}

// Close the logger
func LogClose() {
	if logFile != nil {
		err := logFile.Close()
		if err != nil {
			log.Println("Ошибка при закрытии файла логов:", err)
		}
	}
}

// Parse report to table
func ParseReport(R report.DateRangeReport) [][]string {
	rows := make([][]string, 0)
	for name, calendars := range R.Reports {
		for _, c := range *calendars {
			for _, event := range c.Data.Events() {
				row := make([]string, 0)
				t, _ := time.Parse(timeFormat, event.Props.Get(ical.PropDateTimeStart).Value)
				row = append(row, t.Format("02.01.06"))
				loc := event.Props.Get(ical.PropLocation)
				if loc == nil {
					loc = &ical.Prop{
						Name:  ical.PropLocation,
						Value: name,
					}
				}
				row = append(row, loc.Value)
				text, err := event.Props.Get(ical.PropSummary).Text()
				if err != nil {
					text = event.Props.Get(ical.PropSummary).Value
				}

				row = append(row, text)

				for _, w := range categories {
					if event.Props.Get(w.Tag) != nil {
						row = append(row, w.Name) //✔
					} else {
						row = append(row, "")
					}
				}
				rows = append(rows, row)
			}
		}
	}

	slices.SortFunc(rows, func(a []string, b []string) int {
		if a[0] >= b[0] {
			return 1
		} else {
			return -1
		}
	})
	return rows
}
