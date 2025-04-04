package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/e0m-ru/echoserver/report"
	"github.com/gorilla/mux"
)

type Port int

func RunServer(port Port) {
	mux := mux.NewRouter()
	mux.HandleFunc("/static/{rest:.*}", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join("static/", strings.TrimPrefix(r.URL.Path, "/static/"))
		info, err := os.Stat(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if info.IsDir() {
			http.Error(w, "Directory listing is not allowed", http.StatusForbidden)
			return
		}
		http.ServeFile(w, r, path)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		report.ReportPage(w, r)
	})
	// mux.HandleFunc("/report/{year:[0-9]{2,4}}/", func(w http.ResponseWriter, r *http.Request) {
	// 	report.ReportYear(w, r)
	// })
	// mux.HandleFunc("/report/{year:[0-9]{2,4}}/{month:[0-9]{1,2}}/", func(w http.ResponseWriter, r *http.Request) {
	// 	report.ReportMonth(w, r)
	// })
	// mux.HandleFunc("/report/{year:[0-9]{2,4}}/{month:[0-9]{1,2}}/{day:[0-9]{1,2}}", func(w http.ResponseWriter, r *http.Request) {
	// 	report.ReportDay(w, r)
	// })
	fmt.Printf("Server listening on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
