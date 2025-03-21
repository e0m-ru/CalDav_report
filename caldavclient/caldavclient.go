package caldavclient

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/google/uuid"
)

var (
	dateFormatString = "2006-01-02"
)

func NewCalDavClient(user, pwd, url string) (caldav.Client, error) {
	c := webdav.HTTPClientWithBasicAuth(nil, user, pwd)
	client, err := caldav.NewClient(c, url)
	return client, err
}

func GetCalendars(ctx context.Context, client caldav.Client) (calendars []caldav.Calendar, err error) {

	principal, err := client.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return calendars, err
	}
	homeset, err := client.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return calendars, err
	}
	calendars, err = client.FindCalendars(ctx, homeset)
	if err != nil {
		return calendars, err
	}
	return calendars, err
}

func BuildDateRangeQuery(start, end time.Time) caldav.CalendarQuery {
	compFilter := caldav.CompFilter{
		Name: "VCALENDAR",
		Props: []caldav.PropFilter{
			{Name: "Name"},
		},
		Comps: []caldav.CompFilter{{
			Name:  "VEVENT",
			Start: start,
			End:   end,
			// Props: []caldav.PropFilter{{
			// 	Name: "SUMMARY",
			// 	TextMatch: &caldav.TextMatch{
			// 		Text: "ОТТ",
			// 	},
			// }},
		}},
	}
	query := caldav.CalendarQuery{
		CompFilter: compFilter,
	}
	return query
}

func NewEvent(title, desc, loc string, st, et time.Time) *ical.Event {
	event := ical.NewEvent()
	uid := uuid.New().String()
	event.Props.SetText(ical.PropUID, uid)
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now())
	event.Props.SetText(ical.PropSummary, title)
	event.Props.SetText(ical.PropDescription, desc)
	event.Props.SetText(ical.PropLocation, loc)
	event.Props.SetDateTime(ical.PropDateTimeStart, st)
	event.Props.SetDateTime(ical.PropDateTimeEnd, et)
	return event
}

func NewCalendar(event *ical.Event) *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "ittsc 2025")
	cal.Children = append(cal.Children, event.Component)
	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(cal); err != nil {
		log.Fatal(err)
	}
	return cal
}
