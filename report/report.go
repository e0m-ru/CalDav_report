package report

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/e0m-ru/caldavreport/caldavclient"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

type dataForTemplate struct {
	Reports  map[string]*[]caldav.CalendarObject
	Name     string
	Works    map[string]bool
	Event    ical.Event
	Tz       *time.Location
	GetText  func(ical.Prop) template.HTML
	getWorks func(ical.Prop) map[string]bool
}

// TimeRange представляет временной диапазон
type TimeRange struct {
	start, end, Now time.Time
}

type calendarData struct {
	name    string
	objList *[]caldav.CalendarObject
	err     error
}

// DateRangeReport представляет отчет по диапазону дат
type DateRangeReport struct {
	calDavPrincipal   *caldavclient.CalDavPrincipal
	Calendars         []caldav.Calendar
	TimeRange         TimeRange
	Reports           map[string]*[]caldav.CalendarObject
	Request           *http.Request
	SelectedCalendars map[string]bool
}

// NewDateRangeReport создает новый отчет по диапазону дат
func NewDateRangeReport(start, end time.Time) (r DateRangeReport, err error) {
	n := time.Now()
	client, err := caldavclient.NewClient()
	if err != nil {
		return DateRangeReport{}, fmt.Errorf("Ошибка webDav создания клиента: %e", err)
	}
	report := DateRangeReport{
		calDavPrincipal: &caldavclient.CalDavPrincipal{
			Ctx:    context.Background(),
			Client: *client,
		},
		TimeRange: TimeRange{
			start: start,
			end:   end,
			Now:   n,
		},
		Reports:           make(map[string]*[]caldav.CalendarObject),
		SelectedCalendars: make(map[string]bool),
	}
	report.calDavPrincipal.Query = caldavclient.BuildDateRangeQuery(report.TimeRange.start, report.TimeRange.end)
	if err := report.getCalendars(); err != nil {
		return report, err
	}
	return report, nil
}

// getCalendars получает список календарей
func (r *DateRangeReport) getCalendars() error {
	principal := r.calDavPrincipal
	lst, err := caldavclient.GetCalendars(principal.Ctx, principal.Client)
	if err != nil {
		return err
	}
	r.Calendars = lst
	for _, c := range lst {
		r.SelectedCalendars[c.Name] = true
	}
	return nil
}

// QueryCalendarData выполняет запрос данных календаря
func (r *DateRangeReport) QueryCalendarData(calendar caldav.Calendar) (lst []caldav.CalendarObject, err error) {
	cdp := r.calDavPrincipal
	return cdp.Client.QueryCalendar(cdp.Ctx, calendar.Path, &cdp.Query)
}

// dict создает карту для передачи данных в шаблон
func Dict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("invalid dict call: uneven number of arguments")
	}
	d := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		d[key] = values[i+1]
	}
	return d, nil
}

func GetText(p ical.Prop) template.HTML {
	s := strings.Replace(p.Value, "\\n", "<br/>", -1)
	s = strings.Replace(s, "\\", "", -1)
	return template.HTML(s)
}

func (r *DateRangeReport) PrintReport(w http.ResponseWriter) {
	baseT, err := ParseBaseTemplate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
	}

	reportT, err := baseT.ParseFiles("static/templates/report.html", "static/templates/events.html", "static/templates/event.html")
	if err != nil {
		log.Fatal("Parse report", err)
	}

	d := dataForTemplate{
		Reports: r.Reports,
		Tz:      time.Now().Location(),
		GetText: GetText,
		Works: map[string]bool{
			"Ф": true,
			"В": true,
			"З": true,
			"С": true,
			"Т": true,
			"К": true,
			"Э": true,
		},
	}

	err = reportT.ExecuteTemplate(w, "base", d)
	if err != nil {
		log.Fatal("Execute", err)
	}
}

// парсит базовый шаблон
func ParseBaseTemplate() (*template.Template, error) {
	funcMap := template.FuncMap{
		"dict":    Dict, // Регистрация функции dict
		"getText": GetText,
	}
	baseT, err := template.New("").Funcs(funcMap).ParseGlob("static/templates/base/*")
	if err != nil {
		return baseT, err
	}
	return baseT, nil
}

// Изменяем
func (r *DateRangeReport) ParseWorks() {
	for _, report := range r.Reports {
		for _, r := range *report {
			props := r.Data.Events()[0].Component.Props

			var cab string
			if len(props["CATEGORIES"]) > 0 {
				cab, _ = props["CATEGORIES"][0].Text()
			}

			if sc(cab, "111", "114", "505") {
				ss(&props, "TV")
			}

			var sum string
			if len(props["SUMMARY"]) > 0 {
				sum, _ = props["SUMMARY"][0].Text()
			}
			var desc string
			if len(props["DESCRIPTION"]) > 0 {
				desc, _ = props["DESCRIPTION"][0].Text()
			}
			s := strings.ToLower(sum + " " + desc)

			if sc(s, "вкс") {
				ss(&props, "TV", "VKS", "SOUND", "VIDEO")
			}
			if sc(s, "видео", "теле") {
				ss(&props, "VIDEO")
			}
			if sc(s, "суфл") {
				ss(&props, "VIDEO", "TV", "SOUND")
			}
			if sc(s, "экран", "телевизор", "проектор", "презентац") {
				ss(&props, "TV")
			}
			if sc(s, "аудио", "звук") {
				ss(&props, "SOUND")
			}
			if sc(s, "синх", "перев", "анг", "фра") {
				ss(&props, "SOUND", "SYNCH")
			}
			if sc(s, "трансл") {
				ss(&props, "VIDEO", "TRANS", "SOUND")
			}
			if sc(s, "фото") {
				ss(&props, "PHOTO")
			}

			// fmt.Printf("%v\n", props["TV"][0])
		}
	}
}

func ss(p *ical.Props, s ...string) {
	for _, v := range s {
		p.Add(&ical.Prop{
			Name:  v,
			Value: "true",
		})
	}
}

func sc(S string, s ...string) bool {
	for _, v := range s {
		if strings.Contains(S, v) {
			return true
		}
	}
	return false
}
