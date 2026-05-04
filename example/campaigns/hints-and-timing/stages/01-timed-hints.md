---
title: Three Timed Hints
author: Alan
slug: timed-hints
difficulty: 2
eta_minutes: 4
start: true
next: difficulty
tags:
  - hints
  - mechanics
answer: "patience"
completion_message: "You didn't even need the 45-second hint."
hints:
  - wait_seconds: 0
    text: "This hint is available **immediately** (`wait_seconds: 0`)."
  - wait_seconds: 15
    text: "This one unlocks after **15 seconds** on the stage."
  - wait_seconds: 45
    text: "And this one only after **45 seconds**. The answer is the word for waiting calmly."
---

Each stage can declare a list of hints. Every hint has a `wait_seconds` value
that controls **when** it becomes available after you open the stage.

This stage has three hints:

1. The first is available right away.
2. The second unlocks after 15 seconds.
3. The third unlocks after 45 seconds.

Open the hints panel and watch them become clickable over time.

The answer — fittingly — is **patience**.
