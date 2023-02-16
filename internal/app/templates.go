package app

import (
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
		tmpl, err := renderPage(tc.dir, page)
		if err != nil {
			return TemplateRenderingError{err: err, TemplateName: page}
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			return TemplateRenderingError{err: err, TemplateName: page}
		}
		return nil
	}

	// If we have a cache, all templates must be rendered in advance, so we can fetch it.
	ts, ok := tc.cache[page]
	if !ok {
		return fmt.Errorf("the template %s does not exist", page)
	}

	if err := ts.ExecuteTemplate(w, "base", data); err != nil {
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

		rendered, err := renderPage(tc.dir, name)
		if err != nil {
			return fmt.Errorf("rendering template %s failed: %w", page, err)
		}
		cache[name] = rendered
	}

	// Return the map.
	tc.cache = cache
	return nil
}

// renderPage renders all necessary files for a single page.
func renderPage(templateDir, page string) (*template.Template, error) {
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
