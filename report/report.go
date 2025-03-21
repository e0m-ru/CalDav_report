package report

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/e0m-ru/echoserver/caldavclient"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

type timeRange struct {
	start, end time.Time
}

type CalDavPrincipal struct {
	ctx    context.Context
	client caldav.Client
	query  caldav.CalendarQuery
}

type DateRangeReport struct {
	calDavPrincipal *CalDavPrincipal
	Сalendars       *[]caldav.Calendar
	timeRange       timeRange
	Reports         map[string][]caldav.CalendarObject
	err             error
}

func NewDateRangeReport(
	ctx context.Context,
	client caldav.Client,
	start, end time.Time) (r DateRangeReport, err error) {
	m := make(map[string][]caldav.CalendarObject)
	report := DateRangeReport{
		calDavPrincipal: &CalDavPrincipal{
			ctx:    ctx,
			client: client,
		},
		timeRange: timeRange{
			start: start,
			end:   end,
		},
		Reports: m,
	}
	report.calDavPrincipal.query = caldavclient.BuildDateRangeQuery(report.timeRange.start, report.timeRange.end)
	err = report.getCalendars()
	if err != nil {
		return report, nil
	}
	return report, nil
}

func (r *DateRangeReport) getCalendars() error {
	principal := r.calDavPrincipal
	lst, err := caldavclient.GetCalendars(principal.ctx, principal.client)
	if err != nil {
		return err
	}
	r.Сalendars = &lst
	return nil
}

func (r DateRangeReport) QueryCalendarData(calendar caldav.Calendar) (err error) {
	cdp := r.calDavPrincipal

	lst, err := cdp.client.QueryCalendar(cdp.ctx, calendar.Path, &cdp.query)

	if err != nil {
		return
	}
	r.Reports[calendar.Name] = lst
	return
}

func PrintAllCalendarsData(w io.Writer, calendarList []caldav.CalendarObject) {
	for _, c := range calendarList {
		tmpl, err := template.ParseGlob("templates/event.html")
		if err != nil {
			panic(err)
		}
		for _, e := range c.Data.Events() {
			fmt.Fprintf(w, "<div>%s\n</div>", printEvent(e, tmpl))
		}
	}
}

func printEvent(cal ical.Event, template *template.Template) string {
	var tpl bytes.Buffer
	if err := template.Execute(&tpl, cal); err != nil {
		return fmt.Sprintf("Error executing template: %v", err)
	}

	return tpl.String()
}

func getPropText(cal ical.Event, propName string) string {
	prop := cal.Props.Get(propName)
	if prop != nil {
		text, _ := prop.Text()
		return text
	}
	return ""
}
