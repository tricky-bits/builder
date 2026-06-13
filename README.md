# builder

A static site generator for [tricky-bits](https://github.com/tricky-bits) puzzle
sites. It turns Markdown files with YAML frontmatter into a fully static HTML
website made of **campaigns** (puzzle series), **stages** (individual puzzles
within a campaign), and standalone **pages** (about, guide, terms, …).

Stage metadata, hints, and completion messages can be XOR-obfuscated in the
rendered HTML so future puzzles aren't trivially readable in page source, and
each campaign gets a featured image (real or a deterministic theme fallback).

## Build & run the example

The binary is built from the repository root:

```sh
go build -o builder .
```

Then build the bundled example site (run from the repo root):

```sh
./builder build -c example/tbb.toml
```

This reads content from `example/`, resolves themes from `themes/`, and writes
the generated site to `example/public/`. Serve it with any static file server:

```sh
cd example/public && python3 -m http.server
```

`build` flags:

| Flag           | Default     | Description          |
|----------------|-------------|----------------------|
| `-c, --config` | `tbb.toml`  | Path to config file  |

## Configuration

### Site config (`tbb.toml`)

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

Nav/footer items: `label`, `href`, optional `icon`, optional `external`.
Internal `href`s are resolved against `base_path`; external ones are emitted as-is.

### Campaign frontmatter (`campaigns/<dir>/campaign.md`)

```yaml
---
title: Getting Started          # required
category: Tutorial              # required
slug: getting-started           # optional, derived from directory name
difficulty: 2                   # optional
order: 1                        # optional, lower sorts first (default: by slug)
featured: true                  # optional, pins to home-page Featured strip
theme: dark                     # optional override (stage > campaign > global)
abstract: Short summary shown on the index instead of the full body
completion_message: Nice work!  # shown when the campaign is completed
assets: [diagram.png]           # files copied alongside the campaign
published_at: 2026-01-01T00:00:00Z
last_update_at: 2026-02-01T00:00:00Z
---
Markdown body of the campaign intro…
```

A campaign is a directory holding `campaign.md`, a `stages/` subdirectory, and
optional `inputs/` and `assets/` subdirectories for stage files:

```
campaigns/<dir>/
  campaign.md
  featured.png        # optional campaign featured image
  stages/<file>.md
  inputs/             # downloadable files referenced by stages
  assets/             # embedded files (images, …) referenced by stages
```

Drop a `featured.png` next to `campaign.md` to set its featured image; otherwise
a deterministic theme fallback is chosen. Stages must form a single linear chain:
exactly one stage marked `start: true`, linked by `next`, with no cycles or
orphans.

Stage input/asset files live in the campaign's `inputs/` and `assets/`
directories (siblings of `stages/`), are copied once per campaign to
`campaigns/<slug>/{inputs,assets}/`, and are deduplicated when shared across
stages. Stage pages are written flat as `campaigns/<slug>/<stage>.html`, so a
stage body references them with the relative paths `inputs/<file>` and
`assets/<file>` — e.g. `![](assets/diagram.png)`. Every such reference must be
declared in the matching `inputs:`/`assets:` list (a list entry need not be
referenced, allowing files the player is meant to discover).

### Stage frontmatter (`campaigns/<dir>/stages/<file>.md`)

```yaml
---
title: Welcome                  # required
author: alice                   # required
difficulty: 1                   # required, 1–5
slug: welcome                   # optional, derived from filename
tags: [intro, warmup]           # optional
eta_minutes: 5                  # optional time estimate
theme: dark                     # optional override (stage > campaign > global)
start: true                     # marks the campaign entry point (exactly one)
next: answering                 # slug of the next stage (omit on last)
answer: 42                      # expected solution; shipped as a SHA-256 hash
answer_sha256: 9c56...          # precomputed answer hash (ship without plaintext)
completion_message: Solved!     # shown on stage completion
inputs: [data.json]             # from inputs/; download chips + ref as inputs/<file>
assets: [diagram.png]           # from assets/; not listed; ref as assets/<file>
hints:                          # ordered, timed click-to-reveal hints
  - wait_seconds: 30
    text: Look at the headers.
  - wait_seconds: 120
    text: It's base64.
published_at: 2026-01-01T00:00:00Z
last_update_at: 2026-02-01T00:00:00Z
---
Markdown body of the puzzle…
```

The `answer` is never shipped in clear text — only its trimmed, lowercased
SHA-256 hash reaches the client for in-browser checking.

To publish a campaign's source without revealing solutions, drop `answer` and
supply `answer_sha256` instead: the SHA-256 hex digest of the trimmed, lowercased
answer (the same value `answer` would produce). A stage may carry either field,
both, or neither — neither is valid for informational stages that only link
`next`. When both are present, the plaintext `answer` wins and a stale
`answer_sha256` is logged as a build warning. A malformed `answer_sha256` (not 64
lowercase hex chars) fails the build so a stage never ships silently unsolvable.

### Page frontmatter (`pages/<file>.md`)

Standalone pages (about, guide, terms, …) rendered to `<slug>.html`.

```yaml
---
title: About                    # required
slug: about                     # optional, derived from filename
theme: dark                     # optional override of the global theme
published_at: 2026-01-01T00:00:00Z
last_updated_at: 2026-02-01T00:00:00Z
---
Markdown body of the page…
```

Theme resolution everywhere follows: **stage → campaign → global**, so the most
specific declared `theme` wins.
