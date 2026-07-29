# CLI Reference

`--config`/`-c` is a persistent flag shared by every subcommand:

| Flag           | Default       | Description         |
|----------------|---------------|-----------------------|
| `-c, --config` | `config.toml` | Path to config file  |

Running `builder` with no subcommand behaves like `build`.

## `build`

Builds the static site from config and Markdown files, once.

```sh
builder build -c config.toml
```

## `serve`

Builds once, then serves the site locally while watching for changes.

```sh
builder serve -c config.toml
```

- Every change to the input directory (campaigns, pages, inputs/assets), the
  active theme's directory, or the config file itself triggers a rebuild.
- Each rebuild renders into a fresh scratch directory (never the configured
  `output_dir`) and atomically swaps it in once it succeeds, so in-flight
  requests never see a half-written site.
- Connected browser tabs auto-reload after each successful rebuild via an
  injected livereload script (websocket-based, like Hugo's dev server). A
  rebuild that fails logs the error and keeps serving the last good build
  instead of crashing.

`serve` flags (in addition to `-c, --config`):

| Flag     | Default     | Description                    |
|----------|-------------|-----------------------------------|
| `--bind` | `127.0.0.1` | Address to bind the dev server.  |
| `--port` | `1313`      | Port to serve on.                 |
