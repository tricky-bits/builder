---
title: Snippet Three
author: Ada
slug: snippet-sql
difficulty: 2
eta_minutes: 2
next: ""
tags:
  - sql
  - data
answer: "sql"
completion_message: "Correct — `SELECT ... FROM ... GROUP BY` is the giveaway."
hints:
  - wait_seconds: 0
    text: "It's a query language for relational databases."
---

Name the language used in the snippet below.

```sql
SELECT category, COUNT(*) AS total
FROM campaigns
WHERE featured = TRUE
GROUP BY category
ORDER BY total DESC;
```

### Clues

- `SELECT ... FROM ... WHERE` clauses.
- Aggregation with `COUNT(*)` and `GROUP BY`.
- Set-based, declarative querying of tables.

The language is **sql**.

Type the answer: `sql` (lowercase).
