package caldavclient

import (
	"context"
	"time"

	"github.com/e0m-ru/echoserver/config"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

var (
	dateFormatString = "2006-01-02"
)

// CalDavPrincipal представляет клиента CalDAV и контекст для запросов
type CalDavPrincipal struct {
	Ctx    context.Context
	Client caldav.Client
	Query  caldav.CalendarQuery
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
		// Props: []caldav.PropFilter{
		// 	{Name: "getetag"},
		// 	{Name: "getcontenttype"},
		// },
		Comps: []caldav.CompFilter{{
			Name:  "VEVENT",
			Start: start,
			End:   end,
		}},
	}
	query := caldav.CalendarQuery{
		CompFilter: compFilter,
	}
	return query
}

func NewClient() (client caldav.Client, err error) {
	C := config.LoadConifg()
	c := webdav.HTTPClientWithBasicAuth(nil, C.YaAuth.YAUSER, C.YaAuth.CALPWD)
	client, err = caldav.NewClient(c, C.YaAuth.YACAL)
	if err != nil {
		return caldav.Client{}, err
	}
	return
}

func NewEvent(title, desc, loc string, st, et time.Time) *ical.Event {
	event := ical.NewEvent()
	event.Props.SetText(ical.PropUID, "ASSA") //uuid.New().String()
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now())
	event.Props.SetText(ical.PropSummary, title)
	event.Props.SetText(ical.PropDescription, desc)
	event.Props.SetText(ical.PropLocation, loc)
	event.Props.SetDateTime(ical.PropDateTimeStart, st)
	event.Props.SetDateTime(ical.PropDateTimeEnd, et)
	//ATTENDEE;PARTSTAT=NEEDS-ACTION;CN=i;ROLE=REQ-PARTICIPANT:mailto:i@e0m.ru
	event.Props.Set(&ical.Prop{
		Name:  "ATTENDEE",
		Value: "mailto:i@e0m.ru",
		Params: ical.Params{
			"PARTSTAT": []string{"NEEDS-ACTION"},
			"CN":       []string{"ASSA"},
			"ROLE":     []string{"REQ-PARTICIPANT"},
		},
	})
	return event
}

func NewCalendar(event *ical.Event) *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "ittsc 2025")
	cal.Children = append(cal.Children, event.Component)
	return cal
}
