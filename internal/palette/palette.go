package palette

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const DefaultURL = "https://flynt-theme.github.io/flynt/palette.json"

type baseToken struct {
	Token string `json:"token"`
	Hex   string `json:"hex"`
}

type accentToken struct {
	Token  string            `json:"token"`
	Hex    string            `json:"hex"`
	Shades map[string]string `json:"shades"`
}

type raw struct {
	Info  map[string]string            `json:"_info"`
	Shades map[string]map[string]string `json:"shades"`
	Dark  []json.RawMessage            `json:"dark"`
	Light []json.RawMessage            `json:"light"`
}

// Context is the template context for one variant.
type Context struct {
	Variant string
	Label   string
	// Base tokens (dashes removed for template access, e.g. tx-2 -> tx2).
	Base map[string]string
	// Accent primaries keyed by name.
	Accents map[string]string
	// All shades per accent: Shades["amber"]["300"].
	Shades map[string]map[string]string
}

func buildContext(variant string, tokens []json.RawMessage, shades map[string]map[string]string) (Context, error) {
	labels := map[string]string{"dark": "Flynt Dark", "light": "Flynt Light"}
	label, ok := labels[variant]
	if !ok && len(variant) > 0 {
		label = "Flynt " + strings.ToUpper(variant[:1]) + variant[1:]
	} else if !ok {
		label = "Flynt"
	}

	base := map[string]string{}
	accents := map[string]string{}

	for _, raw := range tokens {
		// Try base token first (no shades field).
		var bt baseToken
		if err := json.Unmarshal(raw, &bt); err != nil {
			return Context{}, fmt.Errorf("parse token: %w", err)
		}
		name := bt.Token
		hex := "#" + strings.ToUpper(bt.Hex)

		if strings.HasPrefix(name, "bg") || strings.HasPrefix(name, "tx") {
			key := strings.ReplaceAll(name, "-", "")
			base[key] = hex
		} else {
			accents[name] = hex
		}
	}

	return Context{
		Variant: variant,
		Label:   label,
		Base:    base,
		Accents: accents,
		Shades:  shades,
	}, nil
}

func fetch(src string) ([]byte, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(src)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", src, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch %s: status %d", src, resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", src, err)
	}
	return data, nil
}

func Load(src string) ([]Context, error) {
	data, err := fetch(src)
	if err != nil {
		return nil, err
	}

	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse palette: %w", err)
	}

	dark, err := buildContext("dark", r.Dark, r.Shades)
	if err != nil {
		return nil, fmt.Errorf("build dark context: %w", err)
	}
	light, err := buildContext("light", r.Light, r.Shades)
	if err != nil {
		return nil, fmt.Errorf("build light context: %w", err)
	}

	return []Context{dark, light}, nil
}
