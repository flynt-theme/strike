package render

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/flynt-theme/strike/internal/palette"
)

var frontmatterRe = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n`)

func parseTemplate(path string) (outputTmpl string, bodyTmpl string, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("read template: %w", err)
	}

	m := frontmatterRe.FindSubmatch(raw)
	if m == nil {
		return "", "", fmt.Errorf("%s: missing frontmatter block (--- ... ---)", path)
	}

	var outputLine string
	for _, line := range strings.Split(string(m[1]), "\n") {
		if strings.HasPrefix(line, "output:") {
			outputLine = strings.TrimSpace(strings.TrimPrefix(line, "output:"))
		}
	}
	if outputLine == "" {
		return "", "", fmt.Errorf("%s: frontmatter must include 'output:'", path)
	}

	return outputLine, string(raw[len(m[0]):]), nil
}

func parseHex(s string) (r, g, b int64, err error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color %q", "#"+s)
	}
	r, err = strconv.ParseInt(s[0:2], 16, 64)
	if err == nil {
		g, err = strconv.ParseInt(s[2:4], 16, 64)
	}
	if err == nil {
		b, err = strconv.ParseInt(s[4:6], 16, 64)
	}
	return r, g, b, err
}

func hex2rgb(hex string) (string, error) {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d,%d,%d", r, g, b), nil
}

func hex2rgbf(hex string) (string, error) {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%.4f %.4f %.4f", float64(r)/255, float64(g)/255, float64(b)/255), nil
}

func hexchan(hex, channel string) (string, error) {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return "", err
	}
	switch channel {
	case "r":
		return fmt.Sprintf("%.4f", float64(r)/255), nil
	case "g":
		return fmt.Sprintf("%.4f", float64(g)/255), nil
	case "b":
		return fmt.Sprintf("%.4f", float64(b)/255), nil
	default:
		return "", fmt.Errorf("hexchan: unknown channel %q (want r, g, or b)", channel)
	}
}

func nohash(hex string) string {
	return strings.TrimPrefix(hex, "#")
}

func shadeFunc(shades interface{}, color, level string) (string, error) {
	m, ok := shades.(map[string]map[string]string)
	if !ok {
		return "", fmt.Errorf("shade: first argument must be .shades")
	}
	colors, ok := m[color]
	if !ok {
		return "", fmt.Errorf("shade: unknown color %q", color)
	}
	v, ok := colors[level]
	if !ok {
		return "", fmt.Errorf("shade: unknown level %q for color %q", level, color)
	}
	return v, nil
}

func buildFuncs() template.FuncMap {
	return template.FuncMap{
		"lower":    strings.ToLower,
		"upper":    strings.ToUpper,
		"hex2rgb":  hex2rgb,
		"hex2rgbf": hex2rgbf,
		"hexchan":  hexchan,
		"nohash":   nohash,
		"shade":    shadeFunc,
	}
}

func buildData(ctx palette.Context) map[string]any {
	data := map[string]any{
		"variant": ctx.Variant,
		"label":   ctx.Label,
		"shades":  ctx.Shades,
	}
	for k, v := range ctx.Base {
		data[k] = v
	}
	for k, v := range ctx.Accents {
		data[k] = v
	}
	return data
}

type compiledTemplate struct {
	output *template.Template
	body   *template.Template
}

func compileTemplate(path string) (*compiledTemplate, error) {
	outputTmpl, bodyTmpl, err := parseTemplate(path)
	if err != nil {
		return nil, err
	}
	funcs := buildFuncs()
	outT, err := template.New("output").Funcs(funcs).Parse(outputTmpl)
	if err != nil {
		return nil, fmt.Errorf("parse output template: %w", err)
	}
	bodyT, err := template.New("body").Funcs(funcs).Parse(bodyTmpl)
	if err != nil {
		return nil, fmt.Errorf("parse body template: %w", err)
	}
	return &compiledTemplate{output: outT, body: bodyT}, nil
}

func BuildCombined(templatePath string, contexts []palette.Context, outDir string, check bool) error {
	tmpl, err := compileTemplate(templatePath)
	if err != nil {
		return err
	}

	data := map[string]any{}
	for _, ctx := range contexts {
		data[ctx.Variant] = buildData(ctx)
	}

	var outName bytes.Buffer
	if err := tmpl.output.Execute(&outName, data); err != nil {
		return fmt.Errorf("render output filename: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.body.Execute(&body, data); err != nil {
		return fmt.Errorf("render body: %w", err)
	}

	outPath := filepath.Join(outDir, outName.String())

	if check {
		existing, err := os.ReadFile(outPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("check failed: missing: %s", outPath)
		} else if err != nil {
			return err
		} else if !bytes.Equal(existing, body.Bytes()) {
			return fmt.Errorf("check failed: stale: %s", outPath)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, body.Bytes(), 0644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	fmt.Println("wrote", outPath)
	return nil
}

func Build(templatePath string, contexts []palette.Context, outDir string, check bool) error {
	tmpl, err := compileTemplate(templatePath)
	if err != nil {
		return err
	}

	var stale []string

	for _, ctx := range contexts {
		data := buildData(ctx)

		var outName bytes.Buffer
		if err := tmpl.output.Execute(&outName, data); err != nil {
			return fmt.Errorf("render output filename: %w", err)
		}

		var body bytes.Buffer
		if err := tmpl.body.Execute(&body, data); err != nil {
			return fmt.Errorf("render body for %s: %w", ctx.Variant, err)
		}

		outPath := filepath.Join(outDir, outName.String())

		if check {
			existing, err := os.ReadFile(outPath)
			if os.IsNotExist(err) {
				stale = append(stale, "missing: "+outPath)
			} else if err != nil {
				return err
			} else if !bytes.Equal(existing, body.Bytes()) {
				stale = append(stale, "stale:   "+outPath)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(outPath, body.Bytes(), 0644); err != nil {
				return fmt.Errorf("write %s: %w", outPath, err)
			}
			fmt.Println("wrote", outPath)
		}
	}

	if len(stale) > 0 {
		return fmt.Errorf("check failed:\n  %s", strings.Join(stale, "\n  "))
	}
	return nil
}
