package main

import (
	"html/template"
	"net/http"
	"os"
)

func hello(w http.ResponseWriter, req *http.Request) {
	t := template.Must(template.New("index").Parse(`
    <title> Toy Server
    <body>
      Lady Gagagaagagaga..
  `))

	t.Execute(w, nil)
}

func main() {
	http.HandleFunc("/", hello)
	http.Handle("/static", http.FileServer(http.Dir("./static")))

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		panic(err)
	}
}
