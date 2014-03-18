package main

import (
	"html/template"
	"net/http"
	"os"
)

func hello(w http.ResponseWriter, req *http.Request) {
	// Set HTTP headers.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Parse template.
	t := template.Must(template.New("index").Parse(`
  <html>
    <title> Toy Server </title>
    <body>
      Lady Gagagaagagaga...
    </body>
  </html>
  `))

	if e := t.Execute(w, nil); e != nil {
		panic("template render issue")
	}
}

func main() {
	http.HandleFunc("/", hello)
	http.Handle("/static", http.FileServer(http.Dir("./static")))

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		panic(err)
	}
}
