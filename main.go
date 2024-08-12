package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /post/{slug}", SlugHandler(FileReader{}))

	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		log.Fatal(err)
	}
}

type SlugReader interface {
	Read(slug string) (string, error)
}

type FileReader struct{}

func (fr FileReader) Read(slug string) (string, error) {
	f, err := os.Open("posts/" + slug + ".md")
	if err != nil {
		return "", err
	}

	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func SlugHandler(slugReader SlugReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := slugReader.Read(r.PathValue("slug"))
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		fmt.Fprint(w, content)
	}
}
