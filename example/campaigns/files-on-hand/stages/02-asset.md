---
title: An **Embedded** Asset
author: Margaret
slug: asset-image
difficulty: 2
eta_minutes: 3
next: ""
tags:
  - files
  - assets
answer: "diagramé"
completion_message: "Asset embedded and displayed — that's a wrap."
assets:
  - diagram.png
hints:
  - wait_seconds: 0
    text: "The answer is the kind of image embedded below."
---

This stage lists `diagram.png` under `assets:`. The file lives in the campaign's
`assets/` directory and is copied once to `campaigns/<slug>/assets/`. Assets are
**not** shown as downloads — you reference them from the body. Here it is,
embedded with a relative path:

![A supporting diagram](assets/diagram.png)

Because the stage page sits at `campaigns/<slug>/asset-image.html`, a relative
`assets/diagram.png` path resolves to the shared assets directory.

The kind of image shown above is a **diagram**.

The answer is **diagram**.
