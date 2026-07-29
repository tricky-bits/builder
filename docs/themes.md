# Themes

A theme is a directory of HTML templates and static assets that render the
site's HTML.

```text
themes/<name>/
  templates/*.html   # required: home.html, campaign.html, stage.html, page.html, 404.html
  static/            # copied to <output_dir>/assets/<name>/
    fallbacks/       # optional campaign featured-image fallbacks
```

A theme must provide all five required entry templates or the build fails at
load time.

## Resolution order

Theme resolution follows **stage → campaign → global**, so the most specific
declared `theme` wins:

1. A stage's own `theme` field, if set.
2. Otherwise its campaign's `theme` field, if set.
3. Otherwise the global `theme` from `config.toml`.

Standalone pages follow the same rule without a campaign level: page
`theme`, else global.

## Featured image fallbacks

Any image file dropped in `themes/<name>/static/fallbacks/` becomes a
candidate fallback for campaigns that don't ship their own `featured.png`.
Builder picks one deterministically per campaign by hashing the campaign's
slug, so a given campaign always renders with the same fallback across
rebuilds.

The bundled `base` theme is the only theme shipped today.
