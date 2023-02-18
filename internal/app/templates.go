package app

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
)

type TemplateRenderingError struct {
	TemplateName string
	err          error
}

func (e TemplateRenderingError) Error() string {
	return fmt.Sprintf("rendering template %s failed: %s", e.TemplateName, e.err.Error())
}
func (e TemplateRenderingError) Unwrap() error {
	return e.err
}

type Templates struct {
	dir   string
	cache map[string]*template.Template
}

func NewTemplates(templateDir string) Templates {
	return Templates{dir: templateDir}
}

// Render renders a template by the name. If the internal cache is built, e.g. during startup,
// it's used. Otherwise, templates are rendered with the given data on demand. Any errors
// are wrapped inside a [TemplateRenderingError] for further analysis.
func (tc *Templates) Render(w io.Writer, page string, data any) error {
	// If we don't have the templates cached, build them on demand.
	if tc.cache == nil {
		tmpl, err := buildPageTemplate(tc.dir, page)
		if err != nil {
			return TemplateRenderingError{err: err, TemplateName: page}
		}

		if err := safeRenderTemplate(tmpl, w, data); err != nil {
			return TemplateRenderingError{err: err, TemplateName: page}
		}
		return nil
	}

	// If we have a cache, all templates must be rendered in advance, so we can fetch it.
	ts, ok := tc.cache[page]
	if !ok {
		return fmt.Errorf("the template %s does not exist", page)
	}

	if err := safeRenderTemplate(ts, w, data); err != nil {
		return TemplateRenderingError{err: err, TemplateName: page}
	}
	return nil
}

func (tc *Templates) BuildCache() error {
	cache := map[string]*template.Template{}

	pagesSubDir := filepath.Join(tc.dir, "pages")
	pages, err := filepath.Glob(filepath.Join(tc.dir, "./**/*.tmpl"))
	if err != nil {
		return fmt.Errorf("error during template search: %w", err)
	}

	for _, page := range pages {
		// Let's make the relative path the template name.
		// The [1:] is used to remove the trailing slash.
		name := strings.TrimPrefix(page, pagesSubDir)[1:]

		rendered, err := buildPageTemplate(tc.dir, name)
		if err != nil {
			return fmt.Errorf("rendering template %s failed: %w", page, err)
		}
		cache[name] = rendered
	}

	// Return the map.
	tc.cache = cache
	return nil
}

// buildPageTemplate renders all necessary files for a single page.
func buildPageTemplate(templateDir, page string) (*template.Template, error) {
	files := []string{
		filepath.Join(templateDir, "base.tmpl"),
		filepath.Join(templateDir, "pages", page),
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		return nil, fmt.Errorf("parsing templates failed: %w", err)
	}

	return ts, nil
}

// safeRenderTemplate renders a page with the given data to the passed writer, only if the rendering process
// is successful. Any runtime errors are wrapped inside a TemplateRenderingError.
func safeRenderTemplate(t *template.Template, w io.Writer, data any) error {
	// Because any rendering error that occurs on runtime can result in incomplete outputs,
	// we write the template to a buffer first. If that's successful, we can copy it to the writer.
	var buf bytes.Buffer

	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		return fmt.Errorf("runtime error on template execution: %w", err)
	}

	// Template rendering succeeded, let's copy it to the actual output.
	// Usually this is the http.ResponseWriter.
	_, _ = buf.WriteTo(w)

	return nil
}
