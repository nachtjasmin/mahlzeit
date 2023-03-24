//go:build generators

package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

var wd, _ = os.Getwd()

type cssFile string

func (me cssFile) Integrity() string {
	f, err := os.Open(filepath.Join(wd, "./dist/", string(me)))
	if err != nil {
		log.Fatalf("opening file %s failed: %v", me, err)
	}
	defer f.Close()

	h := sha512.New384()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatalf("copying file into hash failed: %v", err)
	}

	return fmt.Sprintf("sha384-%s", base64.StdEncoding.EncodeToString(h.Sum(nil)))
}

type manifestEntry struct {
	File           string    `json:"file"`
	Src            string    `json:"src"`
	IsEntry        bool      `json:"isEntry"`
	DynamicImports []string  `json:"dynamicImports"`
	Css            []cssFile `json:"css"`
	IsDynamicEntry bool      `json:"isDynamicEntry"`
}

func (me manifestEntry) Integrity() string {
	f, err := os.Open(filepath.Join(wd, "./dist/", me.File))
	if err != nil {
		log.Fatalf("opening file %s failed: %v", me.File, err)
	}
	defer f.Close()

	h := sha512.New384()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatalf("copying file into hash failed: %v", err)
	}

	return fmt.Sprintf("sha384-%s", base64.StdEncoding.EncodeToString(h.Sum(nil)))
}

type manifest map[string]manifestEntry

const templateStr = `<!-- Code generated using "templates/manifest_gen.go"; DO NOT EDIT. -->

{{"{{ define \"manifest\" }}"}}
{{ range . -}}
{{- range .Css -}}
<link rel="stylesheet" href="/{{ . }}" integrity="{{.Integrity}}">
{{- end }}

{{- if .IsEntry }}
<script src="/{{ .File }}" defer integrity="{{.Integrity}}"></script>
{{ end -}}
{{- end -}}

{{"{{ end }}"}}
`

var tmpl = template.Must(template.New("manifest").Parse(templateStr))

func main() {
	file := os.Args[1]
	if file == "" {
		log.Fatalf("no file path provided")
	}

	file = filepath.Clean(file)
	manifestJson, err := os.Open(file)
	if err != nil {
		log.Fatalf("opening manifest at path %q failed: %v", file, err)
	}

	var m manifest
	if err := json.NewDecoder(manifestJson).Decode(&m); err != nil {
		log.Fatalf("reading manifest failed: %v", err)
	}

	out, err := os.OpenFile(filepath.Clean(os.Args[2]), os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("creating manifest partial failed: %v", err)
	}

	if err := tmpl.Execute(out, m); err != nil {
		log.Fatalf("template execution failed: %v", err)
	}
}
