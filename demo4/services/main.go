package main

import "net/http"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Greeting from Hello Service"))
	})
	http.ListenAndServe(":8080", nil)
}
