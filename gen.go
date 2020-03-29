// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates writers/templates.go. It can be invoked by running
// go generate
package main

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"
)

type genTemplate struct {
	Path    string
	Ident   string
	Content string
	IsHtml  bool
}

func main() {
	files := map[string]string{
		"templates/index.html.tmpl":   "indexTempl",
		"templates/partial.html.tmpl": "partialTempl",
		"templates/start.go.tmpl":     "startTempl",
		"templates/mux.go.tmpl":       "muxTempl",
		"templates/routes.go.tmpl":    "routesTempl",
		"templates/handler.go.tmpl":   "handlerTempl",
	}
	content := make([]genTemplate, 0, 4)

	var fileNames []string
	for k := range files {
		fileNames = append(fileNames, k)
	}
	sort.Strings(fileNames)

	for _, file := range fileNames {
		bites, err := ioutil.ReadFile(file)
		dieIf(err)
		ident := files[file]
		content = append(content, genTemplate{
			Path:    file,
			Ident:   ident,
			Content: string(bites),
			IsHtml:  strings.HasSuffix(file, ".html.tmpl"),
		})
	}

	f, err := os.Create("writers/templates.go")
	dieIf(err)
	defer f.Close()

	packageTemplate.Execute(f, struct {
		Timestamp time.Time
		Templates []genTemplate
	}{
		Timestamp: time.Now(),
		Templates: content,
	})
}

func dieIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var packageTemplate = template.Must(template.New("writer/templates.go").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// {{ .Timestamp }}
package writers

import (
	html "html/template"
	"log"
	text "text/template"
)

var (
	{{ range $index, $template := .Templates -}}
	{{ $template.Ident }}ate {{ if $template.IsHtml -}}
	*html.Template
	{{- else -}}
	*text.Template
	{{- end }}
	{{ end }}
)

func init() {
	var err error
	{{ range $index, $template := .Templates -}}
	{{ $template.Ident }}ate, err = {{ if $template.IsHtml -}}
	html.New("{{ $template.Path }}").Delims("[[", "]]").Parse({{ $template.Ident }})
	{{- else -}}
	text.New("{{ $template.Path }}").Parse({{ $template.Ident }})
	{{- end }}
	if err != nil {
		log.Fatal(err)
	}
	{{ end }}
}


{{- range $index, $template := .Templates }}

// {{ $template.Path }}
var {{ $template.Ident }} = ` + "`{{ $template.Content }}`" + `


{{- end }}
`))
