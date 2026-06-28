package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flynt-theme/strike/internal/palette"
	"github.com/flynt-theme/strike/internal/render"
)

const usage = `Strike — Flynt theme generator

Usage:
  strike build <template> [flags]

Flags:
  --palette   Path or URL to palette.json (default: fetch from flynt-theme/flynt)
  --out       Output directory (default: dist/)
  --check     Exit non-zero if outputs are stale, without writing
  --combined  Render once with both dark and light contexts (for single-file themes)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		runBuild(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}

// partitionArgs separates flag args from positional args so that
// "strike build template.tmpl --out dir" and "strike build --out dir template.tmpl"
// both work. Go's flag package stops at the first non-flag arg otherwise.
func partitionArgs(args []string, boolFlags map[string]bool) (flags, positional []string) {
	i := 0
	for i < len(args) {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			i++
			continue
		}
		name := strings.TrimLeft(arg, "-")
		if idx := strings.Index(name, "="); idx >= 0 {
			name = name[:idx]
		}
		flags = append(flags, arg)
		i++
		if !boolFlags[name] && !strings.Contains(arg, "=") && i < len(args) && !strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			i++
		}
	}
	return
}

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	paletteSrc := fs.String("palette", palette.DefaultURL, "path or URL to palette.json")
	outDir := fs.String("out", "dist", "output directory")
	check := fs.Bool("check", false, "check outputs are up to date without writing")
	combined := fs.Bool("combined", false, "render once with both dark and light contexts")

	flagArgs, posArgs := partitionArgs(args, map[string]bool{"check": true, "combined": true})
	fs.Parse(flagArgs)

	if len(posArgs) == 0 {
		fmt.Fprintln(os.Stderr, "error: template file required")
		os.Exit(1)
	}

	templatePath := posArgs[0]

	if !filepath.IsAbs(*outDir) {
		*outDir = filepath.Join(filepath.Dir(templatePath), *outDir)
	}

	contexts, err := palette.Load(*paletteSrc)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	var buildErr error
	if *combined {
		buildErr = render.BuildCombined(templatePath, contexts, *outDir, *check)
	} else {
		buildErr = render.Build(templatePath, contexts, *outDir, *check)
	}
	if buildErr != nil {
		fmt.Fprintln(os.Stderr, "error:", buildErr)
		os.Exit(1)
	}
}
