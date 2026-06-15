---
title: "**Files** on Hand"
category: Mechanics
slug: files-on-hand
difficulty: 2
order: 22
featured: false
published_at: 2026-05-15T10:00:00Z
abstract: |
  Shows how stages ship extra files: a downloadable input (data.json) that
  appears as a chip in the stage info, and an image asset embedded in the body.
completion_message: |
  ### Files delivered!
  You downloaded an input and viewed an embedded asset. That's the difference
  **Files on Hand** set out to show.
---

Sometimes a stage needs to hand the solver a file. TBB supports two kinds:

- **Inputs** — files listed under `inputs:`. They are copied next to the stage
  and surfaced as a **download chip** in the stage's info section. Use these for
  data the solver needs to open or process.
- **Assets** — files listed under `assets:`. They are also copied next to the
  stage but are **not** listed as downloads; reference them from the body (e.g.
  an image). Use these for supporting media.

Stage 1 demonstrates an input; stage 2 demonstrates an asset.
