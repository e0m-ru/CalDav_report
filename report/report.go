package report

import (
	"context"
	"html/template"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/e0m-ru/echoserver/caldavclient"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

// CalDavPrincipal представляет клиента CalDAV и контекст для запросов
type CalDavPrincipal struct {
	ctx    context.Context
	client caldav.Client
	query  caldav.CalendarQuery
}

// DateRangeReport представляет отчет по диапазону дат
type DateRangeReport struct {
	calDavPrincipal *CalDavPrincipal
	Calendars       []caldav.Calendar
	timeRange       timeRange
	Reports         *map[string]*[]caldav.CalendarObject
}

// timeRange представляет временной диапазон
type timeRange struct {
	start, end time.Time
}

// NewDateRangeReport создает новый отчет по диапазону дат
func NewDateRangeReport(
	ctx context.Context,
	client caldav.Client,
	start, end time.Time) (r DateRangeReport, err error) {
	m := make(map[string]*[]caldav.CalendarObject)
	report := DateRangeReport{
		calDavPrincipal: &CalDavPrincipal{
			ctx:    ctx,
			client: client,
		},
		timeRange: timeRange{
			start: start,
			end:   end,
		},
		Reports: &m,
	}
	report.calDavPrincipal.query = caldavclient.BuildDateRangeQuery(report.timeRange.start, report.timeRange.end)
	if err := report.getCalendars(); err != nil {
		return report, err
	}
	return report, nil
}

// getCalendars получает список календарей
func (r *DateRangeReport) getCalendars() error {
	principal := r.calDavPrincipal
	lst, err := caldavclient.GetCalendars(principal.ctx, principal.client)
	if err != nil {
		return err
	}
	r.Calendars = lst
	return nil
}

// QueryCalendarData выполняет запрос данных календаря
func (r *DateRangeReport) QueryCalendarData(calendar caldav.Calendar) (lst []caldav.CalendarObject, err error) {
	cdp := r.calDavPrincipal
	return cdp.client.QueryCalendar(cdp.ctx, calendar.Path, &cdp.query)
}

func (r *DateRangeReport) PrintReport(w http.ResponseWriter) {
	mime.AddExtensionType(".css", "text/css") // Ensure correct MIME type for CSS files

	baseT, err := template.ParseGlob("templates/base/*")
	if err != nil {
		log.Fatal("Parse:", err)
	}

	reportT, err := baseT.ParseFiles("templates/report.html", "templates/events.html", "templates/event.html")
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
	// for k, v := range *r.Reports {
	// 	for _, events := range *v {
	// 		for _, e := range events.Data.Events() {
	// 			err = tmpl.ExecuteTemplate(
	// 				w,
	// 				"event.html",
	// 				dataForTemplate{
	// 					Name:    k,
	// 					Event:   e,
	// 					Tz:      time.Now().Location(),
	// 					GetText: GetText,
	// 					Works: map[string]bool{
	// 						"Ф": true,
	// 						"В": true,
	// 						"З": true,
	// 						"С": true,
	// 						"Т": true,
	// 						"К": true,
	// 						"Э": true,
	// 					},
	// 				})
	// 			if err != nil {
	// 				fmt.Fprint(w, err)
	// 			}
	// 		}
	// 	}
	// }
}

type dataForTemplate struct {
	Reports  *map[string]*[]caldav.CalendarObject
	Name     string
	Works    map[string]bool
	Event    ical.Event
	Tz       *time.Location
	GetText  func(ical.Prop) string
	getWorks func(ical.Prop) map[string]bool
}

func GetText(p ical.Prop) string {
	text, err := p.Text()
	if err != nil {
		s := strings.Replace(p.Value, "\\", "", -1)
		return s
	}
	return text
}
