package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/e0m-ru/echoserver/config"
	"github.com/e0m-ru/echoserver/report"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

type chanData struct {
	name    string
	objList *[]caldav.CalendarObject
	err     error
}

func reportPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	//TODO date time range from url path
	var (
		now   = time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end   = start.AddDate(0, 1, -1)
	)

	//TODO calendar selection
	allowedCalendars := map[string]bool{
		"111":     true,
		"505":     true,
		"114":     true,
		"116":     true,
		"737":     true,
		"OTT":     true,
		"КЗ":      true,
		"ДИП":     true,
		"ОЗО":     true,
		"Фото":    true,
		"ДКУпДК":  true,
		"Особняк": true,
	}

	C := config.LoadConifg()

	c := webdav.HTTPClientWithBasicAuth(nil, C.YaAuth.YAUSER, C.YaAuth.CALPWD)
	CLNT, err := caldav.NewClient(c, C.YaAuth.YACAL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	R, err := report.NewDateRangeReport(ctx, CLNT, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}

	var wg sync.WaitGroup
	var out = make(chan chanData, len(R.Calendars))

	for _, c := range R.Calendars {
		if allowedCalendars[c.Name] {
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				calendarObjects, err := R.QueryCalendarData(c)
				//TODO return error/
				out <- chanData{c.Name, &calendarObjects, err}
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
		(*R.Reports)[v.name] = v.objList
	}

	R.PrintReport(w)
}

func RunServer() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))
	mux.HandleFunc("/", reportPage)
	fmt.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
