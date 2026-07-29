# Site Config (`config.toml`)

TOML, with two sections: `[build]` (pipeline settings, never exposed to
templates) and `[site]` (global settings exposed to templates).

```toml
[build]
input_dir  = "example"         # root of source content (campaigns/, pages/)
output_dir = "example/public"  # where rendered HTML is written
themes_dir = "themes"          # directory containing themes
theme      = "base"            # active theme name
obfuscation_key = "tricky-bits" # XOR key; "" disables obfuscation (dev mode)

[site]
name           = "Tricky Bits"
base_path      = "/"           # URL prefix the site is served under
document_title = "TBB Example"
index_title    = "Build puzzles. **Ship them as static HTML.**" # **bold** = gradient accent
index_subtitle = "A tiny example site…"
footer_slogan  = "Built to show off the builder."
copyright      = "© 2026 Tricky Bits"

[[site.nav_items]]              # navbar entries, in declared order
label = "Guide"
href  = "/guide.html"

[[site.footer_items]]          # footer entries, in declared order
label    = "GitHub"
href     = "https://github.com/tricky-bits/builder"
external = true                # emits target="_blank"; skips base_path resolution
```

## `[build]` fields

| Field              | Description                                                    |
|--------------------|------------------------------------------------------------------|
| `input_dir`        | Root of source content (`campaigns/`, `pages/`).                |
| `output_dir`       | Where rendered HTML is written.                                  |
| `themes_dir`       | Directory containing themes.                                     |
| `theme`            | Active theme name, resolved from `themes_dir`.                   |
| `obfuscation_key`  | XOR key for stage obfuscation; empty string disables it.         |

## `[site]` fields

| Field             | Description                                                   |
|-------------------|-----------------------------------------------------------------|
| `name`            | Site name, exposed to templates.                                |
| `base_path`       | URL prefix the site is served under.                             |
| `document_title`  | `<title>` tag content.                                           |
| `index_title`     | Home page hero title; `**bold**` renders as a gradient accent.   |
| `index_subtitle`  | Home page hero subtitle.                                         |
| `footer_slogan`   | Footer tagline.                                                  |
| `copyright`       | Footer copyright line.                                           |

## Nav / footer items

Both `[[site.nav_items]]` and `[[site.footer_items]]` accept `label`, `href`,
optional `icon`, and optional `external`. Internal `href`s are resolved
against `base_path`; external ones (`external = true`) are emitted as-is
with `target="_blank"`.
