---
title: A **Downloadable** Input
author: Margaret
slug: input-file
difficulty: 2
eta_minutes: 4
start: true
next: asset-image
tags:
  - files
  - inputs
answer: "octopus"
completion_message: "You read the value straight out of the JSON file."
inputs:
  - data.json
hints:
  - wait_seconds: 0
    text: "Open `data.json` and look at the value of the `codeword` field."
---

This stage ships a file called `data.json`. It lives in the campaign's `inputs/`
directory; because it's listed under `inputs:` in the frontmatter, the builder
copies it to `campaigns/<slug>/inputs/` and shows it as a **downloadable chip**
in the stage info section.

Download it and open it. You'll find a small JSON object:

```json
{
  "campaign": "files-on-hand",
  "codeword": "octopus",
  "note": "On a real puzzle this value would not be repeated below."
}
```

The value of the `codeword` field is the answer. To keep this example
puzzle-free, here it is again in bold: the answer is **octopus**.
