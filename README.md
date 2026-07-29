# builder

A static site generator for [tricky-bits](https://github.com/tricky-bits) puzzle
sites. It turns Markdown files with YAML frontmatter into a fully static HTML
website made of **campaigns** (puzzle series), **stages** (individual puzzles
within a campaign), and standalone **pages** (about, guide, terms, …).

Full documentation, including installation, getting started, configuration,
and the campaign/stage/page frontmatter reference, lives at
**[docs.trickybits.dev](https://docs.trickybits.dev/)**.

## Quick build

```sh
go build -o builder .
./builder build -c example/tbb.toml
cd example/public && python3 -m http.server
```

See the docs site for `serve` (live-reload dev server), full configuration
reference, and the concepts behind campaigns, stages, pages, themes, and
answer obfuscation.
