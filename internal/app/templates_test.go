package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

const testBaseTemplate = `{{ define "base" -}}
{{ block "main" . -}}
base template
{{- end }}
{{- end }}`

const testPageTemplate = `{{ define "main" -}}testPage{{- end }}`

func TestTemplates_Render(t *testing.T) {
	root := t.TempDir()
	tmpl := NewTemplates(root)

	t.Run("template rendering without any files", func(t *testing.T) {
		templateName := "not_existing.tmpl"
		err := tmpl.RenderPage(io.Discard, templateName, nil)
		if err == nil {
			t.Fatalf("expected error, got none")
		}

		var templateErr TemplateRenderingError
		if !errors.As(err, &templateErr) {
			t.Fatalf("expected TemplateRenderingError, got: %#v", err)
		}

		if templateErr.TemplateName != templateName {
			t.Fatalf("template names differ, got: %s, want: %s", templateErr.TemplateName, templateName)
		}
	})

	if err := os.WriteFile(filepath.Join(root, "base.tmpl"), []byte(testBaseTemplate), 0600); err != nil {
		t.Fatalf("writing test base template failed: %v", err)
	}
	pagesPath := filepath.Join(root, "pages")
	if err := os.Mkdir(pagesPath, 0775); err != nil {
		t.Fatalf("creating temporary pages directory failed: %v", err)
	}
	const testPageName = "test.tmpl"
	if err := os.WriteFile(filepath.Join(pagesPath, testPageName), []byte(testPageTemplate), 0600); err != nil {
		t.Fatalf("writing test page failed: %v", err)
	}

	const expectedContent = "testPage"
	t.Run("test page is rendered", func(t *testing.T) {
		var b bytes.Buffer
		err := tmpl.RenderPage(&b, testPageName, nil)
		if err != nil {
			t.Fatalf("page rendering failed: %s", err)
		}

		if b.String() != expectedContent {
			t.Fatalf("content differs, want: %s, got: %s", expectedContent, b.String())
		}
	})

	t.Run("cache can be generated", func(t *testing.T) {
		err := tmpl.BuildCache()
		if err != nil {
			t.Fatalf("got error: %v", err)
		}
	})

	t.Run("test page is rendered from cache", func(t *testing.T) {
		// delete the files to ensure cache usage
		if err := os.RemoveAll(pagesPath); err != nil {
			t.Fatalf("deleting temporary files failed: %v", err)
		}

		var b bytes.Buffer
		err := tmpl.RenderPage(&b, testPageName, nil)
		if err != nil {
			t.Fatalf("page rendering failed: %s", err)
		}

		if b.String() != expectedContent {
			t.Fatalf("content differs, want: %s, got: %s", expectedContent, b.String())
		}
	})

}
