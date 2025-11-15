package badge_gen

import (
	"strings"
	"testing"
)

func TestBadgeGenerator_Simple(t *testing.T) {
	// use the embedded templates (templates/duoText.svg)
	g, err := NewBadgeGeneratorFromEmbed("duoText")
	if err != nil {
		t.Fatalf("NewBadgeGenerator error: %v", err)
	}
	b, err := g.GenerateBadge("duoText", 42)
	if err != nil {
		t.Fatalf("GenerateBadge error: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, ">42<") {
		t.Fatalf("expected 42 in output; got: %q", s[:200])
	}
}

func TestBadgeGenerator_Truncate(t *testing.T) {
	g, err := NewBadgeGeneratorFromEmbed("duoText")
	if err != nil {
		t.Fatalf("NewBadgeGenerator error: %v", err)
	}
	b, err := g.GenerateBadge("duoText", 123456)
	if err != nil {
		t.Fatalf("GenerateBadge error: %v", err)
	}
	if !strings.Contains(string(b), "123456") {
		t.Fatalf("expected exact number in output; got: %q", string(b)[:200])
	}
}

func TestBadgeGenerator_UnknownVariant(t *testing.T) {
	g, err := NewBadgeGeneratorFromEmbed("duoText")
	if err != nil {
		t.Fatalf("NewBadgeGenerator error: %v", err)
	}
	_, err = g.GenerateBadge("missing", 10)
	if err == nil {
		t.Fatalf("expected error for missing template variant")
	}
}

func TestGenerateBadgeFromSingleTemplate(t *testing.T) {
	b, err := GenerateBadgeFromSingleTemplate(3)
	if err != nil {
		t.Fatalf("GenerateBadgeFromSingleTemplate: %v", err)
	}
	if !strings.Contains(string(b), ">3<") {
		t.Fatalf("expected 3 in output")
	}
}
