package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<a href="/">test</a><a href="/unknown"></a>`)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<a href="/test">test</a>`)
	})
	log.Println("listening to port :8080. press ctrl + c to cancel.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
