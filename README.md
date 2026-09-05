# markdown-limit-watcher

A live terminal dashboard that watches a Markdown file and shows how close
each section is to its declared character limit.

---

## What is it useful for?

Many writing workflows impose hard character or word limits per section — think
prompt engineering documents, CV/résumé templates, brief sheets, or spec
documents with word-count constraints. Keeping track of how much space you have
left while you type is tedious.

`markdown-limit-watcher` solves this by:

- **Watching a file in real time** — it polls the file every second and
  re-renders the dashboard whenever the file changes.
- **Showing a colour-coded progress bar** for each limited section:
  - 🟢 **Green** — under 90 % of the limit
  - 🟡 **Yellow** — between 90 % and 100 %
  - 🔴 **Red** — over the limit
- **Counting Unicode characters** (runes), so emoji and non-ASCII text are
  handled correctly.

Typical use cases:

- Writing LLM system-prompt documents where each section must stay under a
  token/character budget.
- Keeping a structured CV within per-section limits.
- Collaborative brief sheets or RFP responses with per-section word caps.

---

## Marking up your Markdown file

Add a limit declaration anywhere inside a section body using this exact syntax:

```markdown
## My Section

- Limit: 300 characters

Write your content here. The watcher counts every character from the line
after the limit declaration up to the next heading.
```

The limit line must match `- Limit: <number> characters` (case-sensitive).
Content **before** the limit line is ignored; only text that follows it
(until the next heading) counts toward the limit.

You can have as many limited sections as you like in a single file:

```markdown
# My Document

## Introduction

- Limit: 200 characters

A short intro goes here.

## Details

- Limit: 500 characters

Longer content lives here.

## Appendix

This section has no limit declaration, so it is ignored by the watcher.
```

---

## Building

Requires Go 1.21 or later.

```bash
# Using make
make

# Or directly
go build -o limit_watcher limit_watcher.go
```

---

## Usage

```
limit_watcher <markdown-file>
```

**Example:**

```bash
./limit_watcher my_document.md
```

The terminal clears and displays a live dashboard:

```
📄  Watching: my_document.md

  Introduction
  87 / 200 characters (43.5%)
  [████████████████░░░░░░░░░░░░░░░░░░░░░░░░]

  Details
  463 / 500 characters (92.6%)
  [██████████████████████████████████████░░]

  Bio
  312 / 300 characters (104.0%)
  [████████████████████████████████████████]
```

The dashboard refreshes automatically whenever you save the file. Press
`Ctrl+C` to quit.
