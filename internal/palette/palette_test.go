package palette

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func writePaletteFile(t *testing.T, dir string, p raw) string {
	t.Helper()
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "palette.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func minimalPalette() raw {
	tok := func(name, hex string, hsl [3]int) token {
		return token{Token: name, Hex: hex, HSL: hsl}
	}
	dark := []token{
		tok("bg", "100E0C", [3]int{30, 14, 5}),
		tok("bg-2", "1C1916", [3]int{30, 12, 10}),
		tok("tx", "F5EDD8", [3]int{40, 60, 90}),
		tok("tx-2", "D8CEBC", [3]int{40, 25, 80}),
		tok("amber", "C69F2A", [3]int{43, 67, 47}),
		tok("ember", "C6372A", [3]int{3, 68, 47}),
	}
	light := []token{
		tok("bg", "FFFCEF", [3]int{50, 100, 97}),
		tok("bg-2", "F2EDDE", [3]int{43, 53, 91}),
		tok("tx", "1A1512", [3]int{30, 16, 9}),
		tok("tx-2", "3D3228", [3]int{30, 21, 20}),
		tok("amber", "C69F2A", [3]int{43, 67, 47}),
		tok("ember", "C6372A", [3]int{3, 68, 47}),
	}
	return raw{
		Info:  map[string]string{"name": "Flynt", "version": "0.1.0"},
		Dark:  dark,
		Light: light,
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := writePaletteFile(t, dir, minimalPalette())

	contexts, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(contexts) != 2 {
		t.Fatalf("len(contexts) = %d, want 2", len(contexts))
	}
	if contexts[0].Variant != "dark" {
		t.Errorf("contexts[0].Variant = %q, want %q", contexts[0].Variant, "dark")
	}
	if contexts[1].Variant != "light" {
		t.Errorf("contexts[1].Variant = %q, want %q", contexts[1].Variant, "light")
	}
}

func TestLoad_FromURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(minimalPalette())
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer srv.Close()

	contexts, err := Load(srv.URL + "/palette.json")
	if err != nil {
		t.Fatalf("Load from URL: %v", err)
	}
	if len(contexts) != 2 {
		t.Fatalf("len(contexts) = %d, want 2", len(contexts))
	}
}

func TestLoad_URLReturnsNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := Load(srv.URL + "/palette.json"); err == nil {
		t.Error("expected error for non-200 response, got nil")
	}
}

func TestLoad_MissingFileReturnsError(t *testing.T) {
	if _, err := Load("/nonexistent/palette.json"); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "palette.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestBuildContext_BaseTokens(t *testing.T) {
	tokens := minimalPalette().Dark
	ctx := buildContext("dark", tokens)

	if got := ctx.Base["bg"]; got != "#100E0C" {
		t.Errorf("Base[bg] = %q, want %q", got, "#100E0C")
	}
	if got := ctx.Base["bg2"]; got != "#1C1916" {
		t.Errorf("Base[bg2] = %q, want %q", got, "#1C1916")
	}
	// Dashes stripped: tx-2 -> tx2
	if got := ctx.Base["tx2"]; got != "#D8CEBC" {
		t.Errorf("Base[tx2] = %q, want %q", got, "#D8CEBC")
	}
}

func TestBuildContext_Accents(t *testing.T) {
	ctx := buildContext("dark", minimalPalette().Dark)
	if got := ctx.Accents["amber"]; got != "#C69F2A" {
		t.Errorf("Accents[amber] = %q, want %q", got, "#C69F2A")
	}
}

func TestBuildContext_ShadesComputed(t *testing.T) {
	ctx := buildContext("dark", minimalPalette().Dark)
	if _, ok := ctx.Shades["amber"]; !ok {
		t.Fatal("Shades[amber] not present")
	}
	if _, ok := ctx.Shades["amber"]["500"]; !ok {
		t.Error("Shades[amber][500] not present")
	}
	if _, ok := ctx.Shades["amber"]["300"]; !ok {
		t.Error("Shades[amber][300] not present")
	}
}

func TestBuildContext_KnownVariantLabel(t *testing.T) {
	ctx := buildContext("dark", minimalPalette().Dark)
	if ctx.Label != "Flynt Dark" {
		t.Errorf("Label = %q, want %q", ctx.Label, "Flynt Dark")
	}
}

func TestBuildContext_UnknownVariantLabel(t *testing.T) {
	ctx := buildContext("custom", minimalPalette().Dark)
	if ctx.Label != "Flynt Custom" {
		t.Errorf("Label = %q, want %q", ctx.Label, "Flynt Custom")
	}
}
