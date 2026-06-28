package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flynt-theme/strike/internal/palette"
)

// testContexts returns a minimal dark+light pair for use in tests.
func testContexts() []palette.Context {
	base := map[string]string{
		"bg":  "#100E0C",
		"bg2": "#1C1916",
		"tx":  "#F5EDD8",
		"tx2": "#D8CEBC",
	}
	accents := map[string]string{
		"amber": "#C69F2A",
		"ember": "#C6372A",
	}
	shades := map[string]map[string]string{
		"amber": {"300": "#D4A84B", "500": "#C69F2A"},
		"ember": {"300": "#D45A4B", "500": "#C6372A"},
	}
	dark := palette.Context{
		Variant: "dark",
		Label:   "Flynt Dark",
		Base:    base,
		Accents: accents,
		Shades:  shades,
	}
	light := palette.Context{
		Variant: "light",
		Label:   "Flynt Light",
		Base: map[string]string{
			"bg":  "#FFFCEF",
			"bg2": "#F2EDDE",
			"tx":  "#1A1512",
			"tx2": "#3D3228",
		},
		Accents: accents,
		Shades:  shades,
	}
	return []palette.Context{dark, light}
}

func writeTemplate(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ── Build ─────────────────────────────────────────────────────────────────────

func TestBuild_WritesOutputPerVariant(t *testing.T) {
	dir := t.TempDir()
	tmpl := writeTemplate(t, dir, "theme.tmpl", `---
output: theme-{{ .variant }}.txt
---
bg={{ .bg }} accent={{ .amber }}
`)
	if err := Build(tmpl, testContexts(), dir, false); err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, variant := range []string{"dark", "light"} {
		path := filepath.Join(dir, "theme-"+variant+".txt")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected output file %s, got: %v", path, err)
		}
	}
}

func TestBuild_CheckPassesWhenUpToDate(t *testing.T) {
	dir := t.TempDir()
	tmpl := writeTemplate(t, dir, "theme.tmpl", `---
output: out-{{ .variant }}.txt
---
{{ .bg }}
`)
	// First build to create the files.
	if err := Build(tmpl, testContexts(), dir, false); err != nil {
		t.Fatalf("Build: %v", err)
	}
	// Check should pass - files match.
	if err := Build(tmpl, testContexts(), dir, true); err != nil {
		t.Errorf("check on up-to-date files: %v", err)
	}
}

func TestBuild_CheckFailsWhenStale(t *testing.T) {
	dir := t.TempDir()
	tmpl := writeTemplate(t, dir, "theme.tmpl", `---
output: out-{{ .variant }}.txt
---
{{ .bg }}
`)
	// Write wrong content.
	for _, v := range []string{"dark", "light"} {
		if err := os.WriteFile(filepath.Join(dir, "out-"+v+".txt"), []byte("stale"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := Build(tmpl, testContexts(), dir, true); err == nil {
		t.Error("expected check to fail on stale files, got nil")
	}
}

func TestBuild_CheckFailsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	tmpl := writeTemplate(t, dir, "theme.tmpl", `---
output: out-{{ .variant }}.txt
---
{{ .bg }}
`)
	if err := Build(tmpl, testContexts(), dir, true); err == nil {
		t.Error("expected check to fail on missing files, got nil")
	}
}

// ── BuildCombined ─────────────────────────────────────────────────────────────

func TestBuildCombined_WritesSingleFile(t *testing.T) {
	dir := t.TempDir()
	tmpl := writeTemplate(t, dir, "theme.tmpl", `---
output: theme.txt
---
dark={{ .dark.bg }} light={{ .light.bg }}
`)
	if err := BuildCombined(tmpl, testContexts(), dir, false); err != nil {
		t.Fatalf("BuildCombined: %v", err)
	}
	out, err := os.ReadFile(filepath.Join(dir, "theme.txt"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	got := string(out)
	if want := "dark=#100E0C light=#FFFCEF\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

// ── Template functions ─────────────────────────────────────────────────────────

func TestHex2RGB(t *testing.T) {
	cases := []struct{ in, want string }{
		{"#C69F2A", "198,159,42"},
		{"C69F2A", "198,159,42"},
	}
	for _, c := range cases {
		got, err := hex2rgb(c.in)
		if err != nil {
			t.Errorf("hex2rgb(%q): %v", c.in, err)
		} else if got != c.want {
			t.Errorf("hex2rgb(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHex2RGB_InvalidReturnsError(t *testing.T) {
	if _, err := hex2rgb("#ZZZ"); err == nil {
		t.Error("expected error for invalid hex, got nil")
	}
	if _, err := hex2rgb("#ABC"); err == nil {
		t.Error("expected error for short hex, got nil")
	}
}

func TestHex2RGBF(t *testing.T) {
	got, err := hex2rgbf("#FFFFFF")
	if err != nil {
		t.Fatalf("hex2rgbf: %v", err)
	}
	if want := "1.0000 1.0000 1.0000"; got != want {
		t.Errorf("hex2rgbf = %q, want %q", got, want)
	}
}

func TestHexchan(t *testing.T) {
	cases := []struct {
		hex, ch, want string
	}{
		{"#FF0000", "r", "1.0000"},
		{"#00FF00", "g", "1.0000"},
		{"#0000FF", "b", "1.0000"},
		{"#FF0000", "g", "0.0000"},
	}
	for _, c := range cases {
		got, err := hexchan(c.hex, c.ch)
		if err != nil {
			t.Errorf("hexchan(%q, %q): %v", c.hex, c.ch, err)
		} else if got != c.want {
			t.Errorf("hexchan(%q, %q) = %q, want %q", c.hex, c.ch, got, c.want)
		}
	}
}

func TestHexchan_InvalidChannelReturnsError(t *testing.T) {
	if _, err := hexchan("#FF0000", "x"); err == nil {
		t.Error("expected error for unknown channel, got nil")
	}
}

func TestShadeFunc(t *testing.T) {
	shades := map[string]map[string]string{
		"amber": {"300": "#D4A84B", "500": "#C69F2A"},
	}
	got, err := shadeFunc(shades, "amber", "300")
	if err != nil {
		t.Fatalf("shadeFunc: %v", err)
	}
	if want := "#D4A84B"; got != want {
		t.Errorf("shadeFunc = %q, want %q", got, want)
	}
}

func TestShadeFunc_UnknownColorReturnsError(t *testing.T) {
	shades := map[string]map[string]string{}
	if _, err := shadeFunc(shades, "nope", "500"); err == nil {
		t.Error("expected error for unknown color, got nil")
	}
}

func TestShadeFunc_UnknownLevelReturnsError(t *testing.T) {
	shades := map[string]map[string]string{"amber": {"500": "#C69F2A"}}
	if _, err := shadeFunc(shades, "amber", "999"); err == nil {
		t.Error("expected error for unknown level, got nil")
	}
}

func TestShadeFunc_BadTypeReturnsError(t *testing.T) {
	if _, err := shadeFunc("not a map", "amber", "500"); err == nil {
		t.Error("expected error for wrong shades type, got nil")
	}
}

// ── parseTemplate ─────────────────────────────────────────────────────────────

func TestParseTemplate_MissingFrontmatterReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := writeTemplate(t, dir, "bad.tmpl", "no frontmatter here")
	if _, _, err := parseTemplate(path); err == nil {
		t.Error("expected error for missing frontmatter, got nil")
	}
}

func TestParseTemplate_MissingOutputReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := writeTemplate(t, dir, "bad.tmpl", "---\nfoo: bar\n---\nbody")
	if _, _, err := parseTemplate(path); err == nil {
		t.Error("expected error for missing output key, got nil")
	}
}
