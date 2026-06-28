<br />

<p align="center">
  <h1 align="center">Strike</h1>
</p>

<p align="center">Theme generator for <a href="https://github.com/flynt-theme/flynt">Flynt</a>. Write one template per app, generate all variants.</p>

<br />

## Install

```sh
go install github.com/flynt-theme/strike/cmd/strike@latest
```

Requires Go 1.21+.

## Usage

```sh
strike build <template> [flags]

Flags:
  --palette   Path or URL to palette.json  (default: flynt-theme/flynt on GitHub)
  --out       Output directory             (default: dist/ next to the template)
  --check     Exit non-zero if outputs are stale, without writing
```

**Example** - build from a local palette:

```sh
strike build theme.yaml.tmpl --palette ../flynt/palette.json
```

**Example** - CI check (verify committed files match the template):

```sh
strike build theme.yaml.tmpl --check
```

## Template format

Templates are plain text files with a YAML frontmatter block:

```
---
output: flynt-{{ .variant }}.yaml
---
# file body — any format, rendered as a Go template
accent: '{{ .amber }}'
background: '{{ .bg }}'
foreground: '{{ .tx }}'
```

Strike renders one output file per palette variant (dark, light). The `output` line is itself a template - use `.variant` to produce separate filenames.

## Template variables

### Variant

| Variable | Example values |
|----------|----------------|
| `.variant` | `dark`, `light` |
| `.label` | `Flynt Dark`, `Flynt Light` |

### Base tokens

| Variable | Token |
|----------|-------|
| `.bg` | `bg` |
| `.bg2` - `.bg5` | `bg-2` through `bg-5` |
| `.tx` | `tx` |
| `.tx2` - `.tx5` | `tx-2` through `tx-5` |

All values are hex strings with a `#` prefix (e.g. `#100E0C`).

### Accent primaries (shade 500)

`.ember` `.clay` `.amber` `.moss` `.fern` `.teal` `.delft` `.iris` `.plum` `.rose`

### Full shade range

```
{{ index (index .shades "amber") "400" }}
```

Available shades: `50` `100` `150` `200` `300` `400` `500` `600` `700` `800` `850` `900` `950`

### Template functions

| Function | Example |
|----------|---------|
| `lower` | `{{ lower .variant }}` |
| `upper` | `{{ upper .label }}` |

## Ports

| App | Repo |
|-----|------|
| Warp | [flynt-theme/flynt-warp](https://github.com/flynt-theme/flynt-warp) |
| Zed | [flynt-theme/flynt-zed](https://github.com/flynt-theme/flynt-zed) |

## License

MIT - [Flynt Theme](https://github.com/flynt-theme)
