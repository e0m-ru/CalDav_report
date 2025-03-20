package caldavclient

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	"github.com/e0m-ru/yacaldav"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

type MonthReport struct {
	calName string
	claList []caldav.CalendarObject
	html    string
}

func Report(client *caldav.Client) (MonthReport, error) {
	ctx := context.Background()
	now := time.Now()
	year := now.Year()
	date := time.Date(year, time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	lst, err := yacaldav.GetCalendarsList(client)
	if err != nil {
		return MonthReport{}, err
	}
	var report MonthReport
	var s strings.Builder
	for _, calendar := range lst {
		report.calName = calendar.Name
		l, err := client.QueryCalendar(ctx, calendar.Path, yacaldav.BuildMonthRangeQuery(date))
		if err != nil {
			report.claList = l
			return report, err
		}
		PrintAllCalendarsData(&s, l)
	}
	report.html = s.String()
	return report, nil
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
	startTime, _ := cal.DateTimeStart(time.Local)
	endTime, _ := cal.DateTimeEnd(time.Local)

	data := struct {
		StartDate   string
		StartTime   string
		EndDate     string
		EndTime     string
		UID         string
		Title       string
		Location    string
		Description string
	}{
		StartDate:   startTime.Format("2006-01-02"),
		StartTime:   startTime.Format("15:04"),
		EndDate:     "",
		EndTime:     endTime.Format("15:04"),
		UID:         getPropText(cal, ical.PropUID),
		Title:       getPropText(cal, ical.PropSummary),
		Location:    getPropText(cal, ical.PropLocation),
		Description: getPropText(cal, ical.PropDescription),
	}

	sy, sm, sd := startTime.Date()
	ey, em, ed := endTime.Date()
	if sy != ey || sm != em || sd != ed {
		data.EndDate = endTime.Format("2006-01-02")
	}

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
