package main

import "github.com/e0m-ru/caldavreport/server"

// "github.com/e0m-ru/caldavreport/server"

func main() {
	// excel.Excelize()
	server.RunServer(8888)
	// 	e := caldavclient.NewEvent("SSSSSSS", "Test event", "Test", time.Now(), time.Now().Add(time.Hour))
	// 	c := caldavclient.NewCalendar(e)
	// 	// R, err := report.NewDateRangeReport()
	// 	// if err != nil {
	// 	// 	log.Fatal(err)
	// 	// }
	// 	client, err := caldavclient.NewClient()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	o, err := client.PutCalendarObject(context.Background(), "/calendars/e0m.ru@ya.ru/events-29358211/1", c)
	// 	if err != nil {
	// 		log.Fatal("\nЧТО\n", err)
	// 	}
	// 	fmt.Printf("%#+v\n", *o)
}
