---
title: Links, Images & Strikethrough
author: Grace
slug: inline
difficulty: 2
eta_minutes: 3
next: ""
tags:
  - markdown
  - links
answer: "linkify"
completion_message: "Links, images, and strikethrough all rendered."
hints:
  - wait_seconds: 0
    text: "GFM auto-converts bare URLs into links — that feature's name is the answer."
---

## Links

A normal [inline link to the guide](/guide.html), a link with a
[title attribute](https://github.com/tricky-bits/builder "TBB on GitHub"), and a
bare URL that GFM turns into a link automatically: https://github.com/tricky-bits/builder

## Images

Markdown images work too. Raw HTML is enabled, so here's an inline SVG that needs
no external file:

<svg width="120" height="40" role="img" aria-label="TBB badge" xmlns="http://www.w3.org/2000/svg">
  <rect width="120" height="40" rx="6" fill="#222"/>
  <text x="60" y="25" fill="#7df" font-family="monospace" font-size="14" text-anchor="middle">TBB</text>
</svg>

## Strikethrough

GFM supports ~~struck-through text~~ with double tildes.

## Your answer

The GFM feature that automatically converts a bare URL into a clickable link is
called **linkify**.

The answer is **linkify**.
