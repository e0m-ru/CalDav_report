package server

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

	"encoding/json"

	"github.com/e0m-ru/echoserver/caldavclient"
	"github.com/e0m-ru/echoserver/config"
	"github.com/e0m-ru/echoserver/report"
	"github.com/emersion/go-webdav/caldav"
)

type assa struct {
	name    string
	objList *[]caldav.CalendarObject
	err     error
}

func echo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	jsw := json.NewEncoder(w)
	err := jsw.Encode(map[string]interface{}{
		"method": r.Method,
		"url":    r.URL.String(),
		"header": r.Header,
		"body":   r.Body,
		"query":  r.URL.Query(),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	client, err := caldavclient.NewCalDavClient(
		C.YaAuth.YAUSER,
		C.YaAuth.CALPWD,
		C.YaAuth.YACAL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	R, err := report.NewDateRangeReport(ctx, client, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}

	var wg sync.WaitGroup
	var out = make(chan assa, len(R.Calendars))

	for _, c := range R.Calendars {
		if allowedCalendars[c.Name] {
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				calendarObjects, err := R.QueryCalendarData(c)
				//TODO return error/
				out <- assa{c.Name, &calendarObjects, err}
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

func mainPage(w http.ResponseWriter, r *http.Request) {
	// Инициализация шаблона
	tmpl, err := template.New("Main").ParseGlob("templates/*")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.ExecuteTemplate(
		w,
		"base.html",
		"",
	)
	if err != nil {
		log.Fatal(err)
	}

}

func RunServer() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static/"))
	mux.HandleFunc("/echo", echo)
	mux.HandleFunc("/report", reportPage)
	mux.HandleFunc("/", mainPage)
	mux.Handle("/static/", http.StripPrefix("/static", fs))

	fmt.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
	// log.Fatal(http.ListenAndServeTLS(":8080", "go-server.crt", "go-server.key", mux))
}

func PrintDetails(w *http.ResponseWriter, v ...interface{}) {
	for _, el := range v {
		elType := reflect.TypeOf(el)
		elValue := reflect.ValueOf(el)
		if elType.Kind() == reflect.Struct {
			for i := range elType.NumField() {
				fmt.Fprintf(*w, "%v: %#+v\n", elType.Field(i).Name, elValue.Field(i))
			}
		} else {
			fmt.Fprintf(*w, "%#+v: %#+v\n", elType.Name(), elValue)
		}
	}
}
