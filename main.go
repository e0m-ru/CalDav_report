package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/e0m-ru/echoserver/caldavclient"
	"github.com/e0m-ru/echoserver/config"
	"github.com/e0m-ru/echoserver/report"
)

var (
	start, end time.Time
)

func init() {
	now := time.Now()
	start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end = start.AddDate(0, 1, -1)
}

func main() {
	C := config.LoadConifg()
	L := log.New(os.Stdout, C.AppName+": ", log.LUTC)
	ctx := context.Background()

	client, err := caldavclient.NewCalDavClient(
		C.YaAuth.YAUSER,
		C.YaAuth.CALPWD,
		C.YaAuth.YACAL)
	if err != nil {
		L.Fatal(err)
	}

	R, err := report.NewDateRangeReport(ctx, client, start, end)
	if err != nil {
		log.Fatal(err)
	}
	allowedCalendars := map[string]bool{
		"111": true,
		"505": true,
		"114": true,
	}

	for _, c := range *R.Сalendars {
		if allowedCalendars[c.Name] {
			R.QueryCalendarData(c)
		}
	}
	s := R.Reports
	L.Printf("%v\n", s)
	// server.RunServer()
}
