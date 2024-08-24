package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/adrg/frontmatter"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

func main() {
	mux := http.NewServeMux()

	postTemplate := template.Must(template.ParseFiles("post.gohtml"))

	mux.HandleFunc("GET /post/{slug}", SlugHandler(FileReader{}, postTemplate))

	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		log.Fatal(err)
	}
}

type SlugReader interface {
	Read(slug string) (string, error)
}

type FileReader struct{}

type PostData struct {
	Content template.HTML
	Title   string `toml:"title"`
	Author  Author `toml:"author"`
}

type Author struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`

}

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

func SlugHandler(slugReader SlugReader, postTemplate *template.Template) http.HandlerFunc {
	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
		),
	)

	return func(w http.ResponseWriter, r *http.Request) {
		content, err := slugReader.Read(r.PathValue("slug"))
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		var post PostData
		rest, err := frontmatter.Parse(strings.NewReader(content), &post)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		var buf bytes.Buffer
		err = mdRenderer.Convert([]byte(rest), &buf)
		if err != nil {
			panic(err)
		}

		post.Content = template.HTML(buf.String())

		postTemplate.Execute(w, post)
	}
}
