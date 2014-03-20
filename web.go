package main

import (
	martini "github.com/codegangsta/martini"
	"net/http"
)

func hello(w http.ResponseWriter, req *http.Request) string {
	// Set HTTP headers.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Parse template.
	return `
    <title> Toy Server </title>
    <body>
      Lady Gagagaagagaga...
  `
}

func main() {
	m := martini.Classic()
	m.Get("/", hello)

	m.Run()
}
