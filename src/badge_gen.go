// This file is a disabled backup of the badge generator implementation and is
// intentionally left out of the build by the `broken` build tag. It contains
// a previous iteration of the generator; `badge_gen.go` is the authoritative
// implementation. Feel free to delete this file when ready.

//go:build broken

package badge_gen

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

// BadgeGenerator manages multiple templates and renders them efficiently.
type BadgeGenerator struct {
	templates   map[string]*template.Template
	defaultName string
	bufPool     sync.Pool
}

// NewBadgeGenerator builds a generator from a map of name->SVG template strings.
func NewBadgeGenerator(templates map[string]string, defaultName string) (*BadgeGenerator, error) {
	if len(templates) == 0 {
		return nil, errors.New("templates map is empty")
	}
	if _, ok := templates[defaultName]; !ok {
		return nil, fmt.Errorf("default template %q not provided", defaultName)
	}
	parsed := make(map[string]*template.Template, len(templates))
	for k, v := range templates {
		t, err := template.New(k).Parse(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %q: %w", k, err)
		}
		parsed[k] = t
	}

	g := &BadgeGenerator{
		templates:   parsed,
		defaultName: defaultName,
		bufPool:     sync.Pool{New: func() any { return new(bytes.Buffer) }},
	}
	return g, nil
}

// GenerateBadge renders the template named by `variant` (default if empty) with the given streak.
func (g *BadgeGenerator) GenerateBadge(variant string, streak int) ([]byte, error) {
	if g == nil {
		return nil, errors.New("nil generator")
	}
	if variant == "" {
		variant = g.defaultName
	}
	tmpl, ok := g.templates[variant]
	if !ok {
		return nil, fmt.Errorf("template variant %q not found", variant)
	}

	if streak < 0 {
		streak = 0
	}

	s := strconv.Itoa(streak)

	data := BadgeData{Streak: s}

	buf := g.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer g.bufPool.Put(buf)

	if err := tmpl.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %q: %w", variant, err)
	}

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

// GenerateBadgeFromSingleTemplate is a convenience helper for the single-template case.
// This prefers embedded templates and does not fall back to inline assets.
func GenerateBadgeFromSingleTemplate(streak int) ([]byte, error) {
	g, err := NewBadgeGeneratorFromEmbed("duoText")
	if err != nil {
		return nil, err
	}
	return g.GenerateBadge("", streak)
}

//go:embed templates/*.svg
var templatesFS embed.FS

// NewBadgeGeneratorFromEmbed parses SVGs embedded in the binary and returns a generator.
// Each template is keyed by its filename (without extension), e.g. `duoText`.
func NewBadgeGeneratorFromEmbed(defaultName string) (*BadgeGenerator, error) {
	entries, err := fs.ReadDir(templatesFS, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded templates: %w", err)
	}

	templates := map[string]string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		b, err := templatesFS.ReadFile(filepath.Join("templates", name))
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded template %q: %w", name, err)
		}
		key := strings.TrimSuffix(name, filepath.Ext(name))
		templates[key] = string(b)
	}

	if defaultName == "" {
		for k := range templates {
			defaultName = k
			break
		}
	}
	return NewBadgeGenerator(templates, defaultName)
}
