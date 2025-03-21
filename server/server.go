package server

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"encoding/json"
)

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
		panic(err)
	}

	// caldavclient.PrintDetails(&w, *r)
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//TODO
}

func RunServer() {

	mux := http.NewServeMux()

	mux.HandleFunc("/echo", echo)
	mux.HandleFunc("/", mainPage)

	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
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
