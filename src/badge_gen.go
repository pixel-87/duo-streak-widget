package badge_gen

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// defaultTemplate is a simple placeholder so the package compiles and tests have
// something to assert against. Replace it with your actual SVG whenever you're
// ready to work on visuals.
const defaultTemplate = `<svg xmlns="http://www.w3.org/2000/svg" width="88" height="31"><text>%STREAK%</text></svg>`

// BadgeGenerator now just swaps streak numbers into string templates. Keeping it
// tiny makes it easy for you to extend later without digging through helpers.

	templates   map[string]string
	defaultName string
}

// NewBadgeGenerator wires a template map with minimal validation.
func NewBadgeGenerator(templates map[string]string, defaultName string) (*BadgeGenerator, error) {
	if len(templates) == 0 {
		return nil, errors.New("no templates provided")
	}
	if defaultName == "" {
		for name := range templates {
			defaultName = name
			break
		}
	}
	if _, ok := templates[defaultName]; !ok {
		return nil, fmt.Errorf("unknown default template %q", defaultName)
	}
	return &BadgeGenerator{templates: templates, defaultName: defaultName}, nil
}

// GenerateBadge replaces the %STREAK% marker with the number provided. There is
// intentionally no pooling or template parsing so you can decide how fancy you
// want this function to become.
func (g *BadgeGenerator) GenerateBadge(variant string, streak int) ([]byte, error) {
	if g == nil {
		return nil, errors.New("badge generator is nil")
	}
	if variant == "" {
		variant = g.defaultName
	}
	tmpl, ok := g.templates[variant]
	if !ok {
		return nil, fmt.Errorf("template %q not found", variant)
	}
	if streak < 0 {
		streak = 0
	}
	svg := strings.ReplaceAll(tmpl, "%STREAK%", strconv.Itoa(streak))
	return []byte(svg), nil
}

// GenerateBadgeFromSingleTemplate keeps the previous helper API around.
func GenerateBadgeFromSingleTemplate(streak int) ([]byte, error) {
	g, err := NewBadgeGeneratorFromEmbed("duoText")
	if err != nil {
		return nil, err
	}
	return g.GenerateBadge("duoText", streak)
}

// NewBadgeGeneratorFromEmbed now just returns the default template. Swap this
// out for real embedded assets whenever you're ready.
func NewBadgeGeneratorFromEmbed(defaultName string) (*BadgeGenerator, error) {
	if defaultName == "" {
		defaultName = "duoText"
	}
	templates := map[string]string{
		"duoText": defaultTemplate,
	}
	return NewBadgeGenerator(templates, defaultName)
}
	if err := tmpl.Execute(buf, data); err != nil {
