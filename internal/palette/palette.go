package palette

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
)

const DefaultURL = "https://raw.githubusercontent.com/flynt-theme/flynt/main/palette.json"

// lightness values per shade step (Flynt spec)
var lightness = map[string]float64{
	"50": 91, "100": 87, "150": 81, "200": 76,
	"300": 64, "400": 53, "500": 47, "600": 42,
	"700": 36, "800": 27, "850": 21, "900": 16, "950": 11,
}

var shadeOrder = []string{"50", "100", "150", "200", "300", "400", "500", "600", "700", "800", "850", "900", "950"}

type token struct {
	Token string    `json:"token"`
	Hex   string    `json:"hex"`
	HSL   [3]int    `json:"hsl"`
}

type raw struct {
	Info  map[string]string `json:"_info"`
	Dark  []token           `json:"dark"`
	Light []token           `json:"light"`
}

// Context is the template context for one variant.
type Context struct {
	Variant string
	Label   string
	// Base tokens (dashes replaced with underscores for template access)
	Base map[string]string
	// Accent primary shades (shade 500)
	Accents map[string]string
	// All shades per accent: Shades["amber"]["300"]
	Shades map[string]map[string]string
}

func hslToHex(h, s, l float64) string {
	s /= 100
	l /= 100
	a := s * math.Min(l, 1-l)
	f := func(n float64) int {
		k := math.Mod(n+h/30, 12)
		val := l - a*math.Max(math.Min(k-3, math.Min(9-k, 1)), -1)
		return int(math.Round(val * 255))
	}
	return fmt.Sprintf("#%02X%02X%02X", f(0), f(8), f(4))
}

func buildContext(variant string, tokens []token) Context {
	labels := map[string]string{"dark": "Flynt Dark", "light": "Flynt Light"}
	label, ok := labels[variant]
	if !ok {
		label = "Flynt " + strings.Title(variant)
	}

	base := map[string]string{}
	accents := map[string]string{}
	accentHues := map[string]float64{}

	for _, t := range tokens {
		name := t.Token
		hex := "#" + t.Hex
		if strings.HasPrefix(name, "bg") || strings.HasPrefix(name, "tx") {
			key := strings.ReplaceAll(name, "-", "")
			base[key] = hex
		} else {
			accents[name] = hex
			accentHues[name] = float64(t.HSL[0])
		}
	}

	shades := map[string]map[string]string{}
	for name, hue := range accentHues {
		shades[name] = map[string]string{}
		for _, sh := range shadeOrder {
			shades[name][sh] = hslToHex(hue, 50, lightness[sh])
		}
	}

	return Context{
		Variant: variant,
		Label:   label,
		Base:    base,
		Accents: accents,
		Shades:  shades,
	}
}

func Load(src string) ([]Context, error) {
	var data []byte
	var err error

	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		resp, err := http.Get(src)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", src, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("fetch %s: status %d", src, resp.StatusCode)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = os.ReadFile(src)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", src, err)
		}
	}

	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse palette: %w", err)
	}

	return []Context{
		buildContext("dark", r.Dark),
		buildContext("light", r.Light),
	}, nil
}
