package report

import (
	"context"
	"net/http"
	"sync"
	"time"
)

var (
	ctx = context.Background()
	now = time.Now()
)

func ReportPage(w http.ResponseWriter, r *http.Request) {
	var selectedMonth string
	var selectedCalendars = make(map[string]bool)
	var err error
	var start, end time.Time

	// Обработка данных формы
	if r.Method == http.MethodPost {
		// Получение выбранного месяца
		selectedMonth = r.FormValue("month")
		if selectedMonth == "" {
			http.Error(w, "Не указан месяц", http.StatusBadRequest)
			return
		}
		start, err = time.Parse("2006-01", selectedMonth)
		if err != nil {
			http.Error(w, "Не получилось распознать дату", http.StatusBadRequest)
		}
		end = start.AddDate(0, 1, 0)
	} else {
		// Установить значения по умолчанию для GET-запроса
		selectedMonth = now.Format("2006-01")
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)
	}

	// Создание объекта отчёта
	R, err := NewDateRangeReport(start, end)
	R.Request = r
	if err != nil {
		http.Error(w, "Ошибка создания объекта отчёта: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получение выбранных календарей
	for _, c := range r.Form["calendars"] {
		selectedCalendars[c] = true
	}

	// Получение данных
	if r.Method == http.MethodPost {
		var wg sync.WaitGroup
		var out = make(chan calendarData, len(R.Calendars))

		for _, c := range R.Calendars {
			if selectedCalendars[c.Name] {
				wg.Add(1)
				go func(wg *sync.WaitGroup) {
					calendarObjects, err := R.QueryCalendarData(c)
					//TODO return error/
					out <- calendarData{c.Name, &calendarObjects, err}
					wg.Done()
				}(&wg)
			}
		}

		wg.Wait()
		close(out)

		for v := range out {
			if v.err != nil {
				http.Error(w, v.err.Error(), http.StatusInternalServerError)
			}
			R.Reports[v.name] = v.objList
		}
	}

	// TODO ?? это кажется лишним.
	R.TimeRange.Now, _ = time.Parse("2006-01", selectedMonth)

	if r.Method == http.MethodPost {
		R.SelectedCalendars = selectedCalendars
	}

	// Парсим шаблон
	baseT, err := ParseBaseTemplate()
	if err != nil {
		http.Error(w, "Ошибка парсинга базового шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	reportT, err := baseT.ParseFiles("static/templates/reportRequestForm.html")
	if err != nil {
		http.Error(w, "Ошибка парсинга шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = reportT.ExecuteTemplate(w, "base", R)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
	R.PrintReport(w)
}
