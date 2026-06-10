---
title: "About **Tricky Bits**"
slug: about
published_at: 2026-06-01T09:00:00Z
---

# About this example

Welcome to the **Tricky Bits Builder** example site. It is a small, deliberately
varied project whose only job is to **demonstrate what the builder can do**.

If you are evaluating TBB, browse around: every campaign on the home page was
chosen to exercise a different part of the system.

## What this example demonstrates

- **Pages** rendered to the site root (`about.html`, `guide.html`, `terms.html`).
- **Navigation** — pages wired into the navbar and footer from `tbb.toml`.
- **Featured vs. ordinary campaigns** — featured ones appear in the home-page strip.
- **Rich markdown** — headings, lists, tables, task lists, blockquotes, code
  blocks, links, images, and emphasis.
- **Stage mechanics** — linear stage chains, timed hints, completion messages,
  difficulty and time estimates, downloadable inputs, and image assets.
- **Featured images** — a real `featured.png` on one campaign; automatic theme
  fallbacks on the rest.

## How it works

TBB reads markdown files with YAML frontmatter and renders them into a
self-contained static site. Answers are validated client-side with SHA-256, so
no server is required. The source for everything you see lives in the
[`example/` directory on GitHub](https://github.com/tricky-bits/builder).

> This is open source. If the builder is useful to you, consider leaving a star.

For a tour of the file layout and frontmatter fields, see the
[Author's Guide](/guide.html).
