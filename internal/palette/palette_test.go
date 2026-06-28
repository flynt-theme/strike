package palette

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func minimalRaw() raw {
	darkTokens := []json.RawMessage{
		mustMarshal(baseToken{Token: "bg", Hex: "100E0C"}),
		mustMarshal(baseToken{Token: "bg-2", Hex: "1C1916"}),
		mustMarshal(baseToken{Token: "tx", Hex: "F5EDD8"}),
		mustMarshal(baseToken{Token: "tx-2", Hex: "D8CEBC"}),
		mustMarshal(accentToken{Token: "amber", Hex: "C69F2A"}),
		mustMarshal(accentToken{Token: "ember", Hex: "C6372A"}),
	}
	lightTokens := []json.RawMessage{
		mustMarshal(baseToken{Token: "bg", Hex: "FFFCEF"}),
		mustMarshal(baseToken{Token: "bg-2", Hex: "F2EDDE"}),
		mustMarshal(baseToken{Token: "tx", Hex: "1A1512"}),
		mustMarshal(baseToken{Token: "tx-2", Hex: "3D3228"}),
		mustMarshal(accentToken{Token: "amber", Hex: "C69F2A"}),
		mustMarshal(accentToken{Token: "ember", Hex: "C6372A"}),
	}
	shades := map[string]map[string]string{
		"amber": {"300": "#D4A84B", "500": "#C69F2A"},
		"ember": {"300": "#D45A4B", "500": "#C6372A"},
	}
	return raw{
		Info:   map[string]string{"name": "Flynt", "version": "0.1.0"},
		Shades: shades,
		Dark:   darkTokens,
		Light:  lightTokens,
	}
}

func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

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

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := writePaletteFile(t, dir, minimalRaw())

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
		data, _ := json.Marshal(minimalRaw())
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
	p := minimalRaw()
	ctx, err := buildContext("dark", p.Dark, p.Shades)
	if err != nil {
		t.Fatalf("buildContext: %v", err)
	}
	if got := ctx.Base["bg"]; got != "#100E0C" {
		t.Errorf("Base[bg] = %q, want %q", got, "#100E0C")
	}
	// Dashes stripped: bg-2 -> bg2
	if got := ctx.Base["bg2"]; got != "#1C1916" {
		t.Errorf("Base[bg2] = %q, want %q", got, "#1C1916")
	}
	// tx-2 -> tx2
	if got := ctx.Base["tx2"]; got != "#D8CEBC" {
		t.Errorf("Base[tx2] = %q, want %q", got, "#D8CEBC")
	}
}

func TestBuildContext_Accents(t *testing.T) {
	p := minimalRaw()
	ctx, err := buildContext("dark", p.Dark, p.Shades)
	if err != nil {
		t.Fatalf("buildContext: %v", err)
	}
	if got := ctx.Accents["amber"]; got != "#C69F2A" {
		t.Errorf("Accents[amber] = %q, want %q", got, "#C69F2A")
	}
}

func TestBuildContext_ShadesFromJSON(t *testing.T) {
	p := minimalRaw()
	ctx, err := buildContext("dark", p.Dark, p.Shades)
	if err != nil {
		t.Fatalf("buildContext: %v", err)
	}
	if got := ctx.Shades["amber"]["300"]; got != "#D4A84B" {
		t.Errorf("Shades[amber][300] = %q, want %q", got, "#D4A84B")
	}
	if got := ctx.Shades["amber"]["500"]; got != "#C69F2A" {
		t.Errorf("Shades[amber][500] = %q, want %q", got, "#C69F2A")
	}
}

func TestBuildContext_KnownVariantLabel(t *testing.T) {
	p := minimalRaw()
	ctx, err := buildContext("dark", p.Dark, p.Shades)
	if err != nil {
		t.Fatalf("buildContext: %v", err)
	}
	if ctx.Label != "Flynt Dark" {
		t.Errorf("Label = %q, want %q", ctx.Label, "Flynt Dark")
	}
}

func TestBuildContext_UnknownVariantLabel(t *testing.T) {
	p := minimalRaw()
	ctx, err := buildContext("custom", p.Dark, p.Shades)
	if err != nil {
		t.Fatalf("buildContext: %v", err)
	}
	if ctx.Label != "Flynt Custom" {
		t.Errorf("Label = %q, want %q", ctx.Label, "Flynt Custom")
	}
}
