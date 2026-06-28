# Strike

Theme generator for [Flynt](https://github.com/flynt-theme/flynt). Write one template per app, generate all variants.

## Install

```sh
brew tap flynt-theme/tap
brew install strike
```

Or with Go:

```sh
go install github.com/flynt-theme/strike/cmd/strike@latest
```

## Usage

```sh
strike build <template> [flags]

Flags:
  --palette   Path or URL to palette.json  (default: flynt-theme/flynt on GitHub)
  --out       Output directory             (default: dist/ next to the template)
  --combined  Render once with both dark and light contexts (for single-file themes)
  --check     Exit non-zero if outputs are stale, without writing
```

**Example** - build from a local palette:

```sh
strike build theme.yaml.tmpl --palette ../flynt/palette.json
```

**Example** - single-file theme (e.g. Obsidian):

```sh
strike build theme.css.tmpl --combined
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

| Function | Example | Output |
|----------|---------|--------|
| `hex2rgb` | `{{ hex2rgb .amber }}` | `198,159,42` |
| `hex2rgbf` | `{{ hex2rgbf .amber }}` | `0.7765 0.6235 0.1647` |
| `hexchan` | `{{ hexchan .amber "r" }}` | `0.7765` |
| `nohash` | `{{ nohash .amber }}` | `C69F2A` |
| `shade` | `{{ shade .dark.shades "amber" "400" }}` | hex string |
| `lower` | `{{ lower .variant }}` | `dark` |
| `upper` | `{{ upper .label }}` | `FLYNT DARK` |

`hexchan` accepts `"r"`, `"g"`, or `"b"` as the channel argument. Useful for formats like iTerm2 that need each component on a separate line.

## Ports

| App | Repo |
|-----|------|
| Obsidian | [flynt-theme/flynt-obsidian](https://github.com/flynt-theme/flynt-obsidian) |
| Warp | [flynt-theme/flynt-warp](https://github.com/flynt-theme/flynt-warp) |
| Zed | [flynt-theme/flynt-zed](https://github.com/flynt-theme/flynt-zed) |
| Ghostty | [flynt-theme/flynt-ghostty](https://github.com/flynt-theme/flynt-ghostty) |
| iTerm2 | [flynt-theme/flynt-iterm2](https://github.com/flynt-theme/flynt-iterm2) |

## License

MIT - [Flynt Theme](https://github.com/flynt-theme)
