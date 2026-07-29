# Pages

Standalone pages (about, guide, terms, a 404) that aren't part of any
campaign. Each `pages/<file>.md` renders to `<slug>.html`.

```text
pages/
  about.md
  guide.md
  terms.md
  404.md
```

Pages support the same `theme` override as campaigns and stages (see
[Themes](themes.md)) but carry no answer, hints, or chain position, just
title, optional slug, and a Markdown body.

## Frontmatter

`pages/<file>.md`: standalone pages (about, guide, terms, …) rendered to
`<slug>.html`.

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

| Field              | Required | Notes                                    |
|--------------------|----------|---------------------------------------------|
| `title`            | yes      |                                              |
| `slug`             | no       | Derived from the filename if omitted.        |
| `theme`            | no       | Overrides the global theme for this page.    |
| `published_at`     | no       | RFC 3339 timestamp.                          |
| `last_updated_at`  | no       | RFC 3339 timestamp.                          |
