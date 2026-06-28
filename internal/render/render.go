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

func hex2rgb(hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "0,0,0"
	}
	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)
	return fmt.Sprintf("%d,%d,%d", r, g, b)
}

func shadeFunc(shades interface{}, color, level string) string {
	m, ok := shades.(map[string]map[string]string)
	if !ok {
		return "#000000"
	}
	if colors, ok := m[color]; ok {
		if v, ok := colors[level]; ok {
			return v
		}
	}
	return "#000000"
}

func buildFuncs() template.FuncMap {
	return template.FuncMap{
		"lower":   strings.ToLower,
		"upper":   strings.ToUpper,
		"hex2rgb": hex2rgb,
		"shade":   shadeFunc,
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

func BuildCombined(templatePath string, contexts []palette.Context, outDir string, check bool) error {
	outputTmpl, bodyTmpl, err := parseTemplate(templatePath)
	if err != nil {
		return err
	}

	funcs := buildFuncs()

	outT, err := template.New("output").Funcs(funcs).Parse(outputTmpl)
	if err != nil {
		return fmt.Errorf("parse output template: %w", err)
	}
	bodyT, err := template.New("body").Funcs(funcs).Parse(bodyTmpl)
	if err != nil {
		return fmt.Errorf("parse body template: %w", err)
	}

	data := map[string]any{}
	for _, ctx := range contexts {
		data[ctx.Variant] = buildData(ctx)
	}

	var outName bytes.Buffer
	if err := outT.Execute(&outName, data); err != nil {
		return fmt.Errorf("render output filename: %w", err)
	}

	var body bytes.Buffer
	if err := bodyT.Execute(&body, data); err != nil {
		return fmt.Errorf("render body: %w", err)
	}

	outPath := filepath.Join(outDir, outName.String())

	if check {
		existing, err := os.ReadFile(outPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("check failed: missing: %s", outPath)
		} else if err != nil {
			return err
		} else if string(existing) != body.String() {
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
	outputTmpl, bodyTmpl, err := parseTemplate(templatePath)
	if err != nil {
		return err
	}

	funcs := buildFuncs()

	outT, err := template.New("output").Funcs(funcs).Parse(outputTmpl)
	if err != nil {
		return fmt.Errorf("parse output template: %w", err)
	}
	bodyT, err := template.New("body").Funcs(funcs).Parse(bodyTmpl)
	if err != nil {
		return fmt.Errorf("parse body template: %w", err)
	}

	var stale []string

	for _, ctx := range contexts {
		data := buildData(ctx)

		var outName bytes.Buffer
		if err := outT.Execute(&outName, data); err != nil {
			return fmt.Errorf("render output filename: %w", err)
		}

		var body bytes.Buffer
		if err := bodyT.Execute(&body, data); err != nil {
			return fmt.Errorf("render body for %s: %w", ctx.Variant, err)
		}

		outPath := filepath.Join(outDir, outName.String())

		if check {
			existing, err := os.ReadFile(outPath)
			if os.IsNotExist(err) {
				stale = append(stale, "missing: "+outPath)
			} else if err != nil {
				return err
			} else if string(existing) != body.String() {
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
