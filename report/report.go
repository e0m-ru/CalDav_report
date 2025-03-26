package report

import (
	"context"
	"fmt"
	"html/template"
	"log"
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Инициализация шаблона
	tmpl, err := template.New("CalDav").ParseGlob("templates/*")
	if err != nil {
		fmt.Print(err)
		return
	}
	err = tmpl.ExecuteTemplate(
		w,
		"index.html",
		"",
	)
	for k, v := range *r.Reports {
		for _, events := range *v {
			for _, e := range events.Data.Events() {
				err = tmpl.ExecuteTemplate(
					w,
					"event.html",
					dataForTemplate{
						Name:    k,
						Event:   e,
						Tz:      time.Now().Location(),
						GetText: GetText,
						Works: []string{
							"фото",
							"видео",
							"звук",
							"синхрон",
							"трансляция",
							"экран",
						},
					})
				if err != nil {
					fmt.Fprint(w, err)
				}
			}
		}
	}
	_, err = fmt.Fprint(w, `</table>
	        <script src='static/js/tablesort.js'></script>
</body>
</html>`)
	if err != nil {
		log.Fatal(err)
	}

}

type dataForTemplate struct {
	Name    string
	Works   []string
	Event   ical.Event
	Tz      *time.Location
	GetText func(ical.Prop) string
}

func GetText(p ical.Prop) string {
	text, err := p.Text()
	if err != nil {
		s := strings.Replace(p.Value, "\\", "", -1)
		return s
	}
	return text
}
