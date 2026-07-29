# Campaigns

Campaigns are the main component of a tricky bits site: the website's home
page presents a catalogue of campaigns for players to pick from.

A campaign is a puzzle series: an ordered chain of [stages](stages.md) plus
one `campaign.md` file carrying the series' metadata and intro text.

## Directory layout

A campaign is a directory holding `campaign.md`, a `stages/` subdirectory,
and optional `inputs/` and `assets/` subdirectories for files stages
reference:

```text
campaigns/<dir>/
  campaign.md
  featured.png        # optional campaign featured image
  stages/<file>.md
  inputs/             # downloadable files referenced by stages
  assets/             # embedded files (images, …) referenced by stages
```

## Featured images

Drop a `featured.png` (or `.avif`/`.webp`/`.jpg`/`.jpeg`/`.gif`) next to
`campaign.md` to set its featured image. Otherwise builder picks a
deterministic fallback from the active theme, chosen by hashing the
campaign's slug: the same campaign always gets the same fallback image.

## Inputs and assets

Stage input/asset files live in the campaign's `inputs/` and `assets/`
directories (siblings of `stages/`). They're copied once per campaign and
deduplicated when shared across stages.

The two differ in intent: an **input** is something a player needs to solve
the puzzle, so it's clearly presented on the generated stage page (a download
link). An **asset** just helps serve the content (an image embedded in the
body, a font, …), it's copied to the right place with no special
presentation, campaign makers are expected to know how to reference it
themselves.

Every reference must be declared in the matching `inputs:`/`assets:` list on
the stage (see [Stage Frontmatter](stages.md#frontmatter)). A list entry
doesn't have to be referenced in the body, which lets a puzzle ship a file
the player is meant to discover rather than one you link to directly.

## Generated output layout

Each campaign is rendered flat under `campaigns/<slug>/` in the output
directory:

```text
<output_dir>/campaigns/<slug>/
  index.html          # campaign landing page
  featured.png         # featured image, real or fallback
  <stage>.html         # one file per stage, flat (no stages/ subdirectory)
  inputs/              # copied from the campaign's inputs/
  assets/              # copied from the campaign's assets/
```

Because stage pages are written flat as `campaigns/<slug>/<stage>.html`, a
stage body references its files with relative paths, e.g.
`![](assets/diagram.png)`.

## Frontmatter

`campaigns/<dir>/campaign.md`:

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

| Field                 | Required | Notes                                                          |
|-----------------------|----------|-----------------------------------------------------------------|
| `title`               | yes      |                                                                    |
| `category`            | yes      |                                                                    |
| `slug`                | no       | Derived from the directory name if omitted.                      |
| `difficulty`          | no       | Unconstrained integer indicator (unlike stage `difficulty`).      |
| `order`               | no       | Lower sorts first; campaigns without it sort by slug.             |
| `featured`            | no       | Pins the campaign to the home page's Featured strip.              |
| `theme`               | no       | Overrides the global theme for this campaign and its stages.      |
| `abstract`            | no       | Shown on the index instead of the full body.                      |
| `completion_message`  | no       | Shown when the player finishes the campaign's last stage.          |
| `assets`              | no       | List of files copied alongside the campaign.                      |
| `published_at`        | no       | RFC 3339 timestamp.                                               |
| `last_update_at`      | no       | RFC 3339 timestamp.                                               |
