package report

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/e0m-ru/echoserver/caldavclient"
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Инициализация шаблона
	tmpl, err := template.ParseGlob("templates/event.html")
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	for k, v := range *r.Reports {
		fmt.Fprintf(w, "<h1>-----CALENDAR %s-------</h1>\n", k)
		for _, e := range *v {
			err = tmpl.Execute(w, e.Data.Events()[0])
			if err != nil {
				fmt.Fprint(w, err)
			}

			// p := e.Data.Events()[0].Component.Props
			// ds, err := p.Get(ical.PropDateTimeStart).DateTime(time.Now().Location())
			// if err != nil {
			// 	fmt.Fprint(w, err)
			// }
			// s, err := p.Get(ical.PropSummary).Text()
			// if err != nil {
			// 	fmt.Fprint(w, err)
			// }
			// fmt.Fprintf(w, "    %s\n    %s\n", ds, s)
		}
	}
}
