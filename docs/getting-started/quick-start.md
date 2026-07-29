# Quick Start

The repository ships a bundled example site under `example/`, ready to build.

## Build once

```sh
./builder build -c example/config.toml
```

This reads content from `example/`, resolves themes from `themes/`, and
writes the generated site to `example/public/`. Serve it with any static
file server:

```sh
cd example/public && python3 -m http.server
```

## Or serve with live reload

Skip the manual rebuild loop with `serve`, which builds, serves, and rebuilds
on every change automatically:

```sh
./builder serve -c example/config.toml
```

Every change to the input directory (campaigns, pages, inputs/assets), the
active theme's directory, or the config file itself triggers a rebuild.
Connected browser tabs auto-reload after each successful rebuild; a failed
rebuild logs the error and keeps serving the last good build.

See the [CLI Reference](../cli-reference.md) for all flags.

## Where the example content lives

```text
example/
  config.toml            # site config
  campaigns/<slug>/       # one directory per campaign
    campaign.md
    stages/<file>.md
  pages/<file>.md         # standalone pages (about, guide, terms, …)
```

Next, read up on [Campaigns](../campaigns.md) and
[Stages](../stages.md) to understand how that content is structured.
