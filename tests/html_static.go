package main

import (
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./tests/static")))
	_ = http.ListenAndServe(":8085", nil)
}
