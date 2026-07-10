# Basic HTML Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add rendering of `<center>`, `<img width/height>`, and `<div align="center">` HTML elements in markdown input.

**Architecture:** Two mechanisms: (1) pre/post-processing in `glamour.go` for `<center>` blocks (which contain markdown and need recursive rendering), (2) AST-level HTML parsing in `ansi/elements.go` for `<img>` with size attributes. `<center>` must be handled at the glamour level because the ansi package cannot import glamour (circular dependency).

**Tech Stack:** Go, `golang.org/x/net/html` for HTML parsing, `golang.org/x/net` as new dependency.

---

### Task 1: Add `golang.org/x/net` dependency

**Files:**
- Modify: `go.mod`
- Run: `go mod tidy`

- [ ] **Step 1: Add the dependency**

Run:
```bash
cd /home/q/files/prjkt/glamoured && go get golang.org/x/net@latest
```

- [ ] **Step 2: Tidy the module**

Run:
```bash
cd /home/q/files/prjkt/glamoured && go mod tidy
```

Expected: clean output, no errors.

- [ ] **Step 3: Verify it compiles**

Run:
```bash
cd /home/q/files/prjkt/glamoured && go build ./...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /home/q/files/prjkt/glamoured && git add -A && git commit -m "chore: add golang.org/x/net dependency"
```

---

### Task 2: Add Width/Height fields to ImageElement for per-image size override

**Files:**
- Modify: `ansi/image.go`
- Modify: `ansi/elements.go` (also used by ImageElement creation)
- Test: `ansi/renderer_test.go` (need to add image test)

First, let me read the current ImageElement struct and its Render method to understand what to modify.

- [ ] **Step 1: Read current ImageElement code**

Read `ansi/image.go` lines 16-23 (struct) and lines 49-91 (Render mosaic path) to confirm current structure.

- [ ] **Step 2: Add Width and Height fields to ImageElement**

In `ansi/image.go`, modify the struct:

```go
type ImageElement struct {
	Text     string
	BaseURL  string
	URL      string
	Child    ElementRenderer
	TextOnly bool
	Width    int
	Height   int
}
```

- [ ] **Step 3: Update ImageElement.Render to use per-image Width/Height**

In `ansi/image.go`, modify the mosaic width calculation (lines 54-63):

```go
width := e.Width
if width <= 0 {
    width = ctx.options.MosaicWidth
}
if width <= 0 {
    width = int(ctx.blockStack.Width(ctx))
    if width <= 0 {
        width = ctx.options.WordWrap / 2
    }
    if width < 20 {
        width = 20
    }
}
```

The per-image `e.Width` takes first priority, then `ctx.options.MosaicWidth`, then the automatic fallback.

- [ ] **Step 4: Build and test**

Run:
```bash
cd /home/q/files/prjkt/glamoured && go build ./... && go test ./...
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
cd /home/q/files/prjkt/glamoured && git add -A && git commit -m "feat: add Width/Height override fields to ImageElement"
```

---

### Task 3: Create HTML parser for `<img>` with width/height in `ansi/elements.go`

**Files:**
- New: none (inline change in elements.go)
- Modify: `ansi/elements.go` (KindHTMLBlock and KindRawHTML cases)
- Test: `ansi/renderer_test.go`

- [ ] **Step 1: Read current HTML block/span handlers**

Read `ansi/elements.go` lines 414-430 to see the current KindHTMLBlock and KindRawHTML handlers.

- [ ] **Step 2: Add a helper function to parse HTML for `<img>` tags**

Add a function after line 430 in `ansi/elements.go`:

```go
// parseHTMLImage parses an HTML string looking for <img> tags.
// Returns the src, width, height if found, otherwise empty strings/0.
// For non-img HTML, returns empty values.
func parseHTMLImage(html string) (src string, width int, height int) {
	doc, err := html.Parse(strings.NewReader(html))
	if err != nil {
		return
	}
	var findImg func(*html.Node) bool
	findImg = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				switch a.Key {
				case "src":
					src = a.Val
				case "width":
					width, _ = strconv.Atoi(a.Val)
				case "height":
					height, _ = strconv.Atoi(a.Val)
				}
			}
			return true
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if findImg(c) {
				return true
			}
		}
		return false
	}
	findImg(doc)
	return
}
```

Add imports: `"golang.org/x/net/html"`, `"strconv"`.

- [ ] **Step 3: Update KindHTMLBlock handler**

Replace lines 415-422:

```go
case ast.KindHTMLBlock:
    n := node.(*ast.HTMLBlock)
    raw := string(n.Text(source))
    src, w, h := parseHTMLImage(raw)
    if src != "" {
        return Element{
            Renderer: &ImageElement{
                BaseURL:  ctx.options.BaseURL,
                URL:      src,
                TextOnly: false,
                Width:    w,
                Height:   h,
            },
        }
    }
    return Element{
        Renderer: &BaseElement{
            Token: ctx.SanitizeHTML(raw, true),
            Style: ctx.options.Styles.HTMLBlock.StylePrimitive,
        },
    }
```

- [ ] **Step 4: Update KindRawHTML handler**

Replace lines 423-430 identically (same logic for inline HTML):

```go
case ast.KindRawHTML:
    n := node.(*ast.RawHTML)
    raw := string(n.Text(source))
    src, w, h := parseHTMLImage(raw)
    if src != "" {
        return Element{
            Renderer: &ImageElement{
                BaseURL:  ctx.options.BaseURL,
                URL:      src,
                TextOnly: false,
                Width:    w,
                Height:   h,
            },
        }
    }
    return Element{
        Renderer: &BaseElement{
            Token: ctx.SanitizeHTML(raw, true),
            Style: ctx.options.Styles.HTMLSpan.StylePrimitive,
        },
    }
```

- [ ] **Step 5: Build and test**

Run:
```bash
cd /home/q/files/prjkt/glamoured && go build ./... && go test ./...
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
cd /home/q/files/prjkt/glamoured && git add -A && git commit -m "feat: parse <img> HTML tags for per-image width/height"
```

---

### Task 4: Add `<center>` pre/post-processing in `glamour.go`

**Files:**
- Modify: `glamour.go`
- Test: `glamour_test.go`

- [ ] **Step 1: Read the current Render/RenderBytes methods**

Read `glamour.go` lines 284-294 to see `Render` and `RenderBytes`.

- [ ] **Step 2: Add center block extraction and injection functions**

Add to `glamour.go` (before `RenderBytes`):

```go
import (
    "regexp"
    "strings"
)

var centerRe = regexp.MustCompile(`(?is)<(?:center|div\s+align="?center"?)\s*>([\s\S]*?)</(?:center|div)\s*>`)

// extractCenterBlocks finds <center>...</center> and <div align="center">...</div>
// blocks, replaces them with markers, and stores the inner content.
func extractCenterBlocks(input string) (string, map[string]string) {
    blocks := make(map[string]string)
    var markerIndex int
    result := centerRe.ReplaceAllStringFunc(input, func(match string) string {
        markerIndex++
        marker := fmt.Sprintf("<glamour-center-%d>", markerIndex)
        inner := centerRe.FindStringSubmatch(match)
        if len(inner) > 1 {
            // Strip leading/trailing blank lines from inner content
            content := strings.TrimSpace(inner[1])
            blocks[marker] = content
        }
        return marker
    })
    return result, blocks
}
```

- [ ] **Step 3: Add centering function**

Add after extractCenterBlocks:

```go
// centerText centers each line of text within the given width.
// It strips ANSI codes to measure visual width, then pads each line.
func centerText(text string, width int) string {
    lines := strings.Split(text, "\n")
    for i, line := range lines {
        // Measure visible width (strip ANSI codes)
        trimmed := ansi.Truncate(line, 999999, "") // strips ANSI, gets visible string
        // But Truncate also truncates — we just use it for ANSI stripping
        // Actually, we need a width function. Use a simple approach:
        visible := ansi.Truncate(line, len(line), "") // just strip ANSI
        _ = visible
        // Simpler: just count visible runes after removing ANSI codes
        visibleWidth := visibleWidth(line)
        if visibleWidth >= width {
            continue
        }
        padding := (width - visibleWidth) / 2
        if padding > 0 {
            lines[i] = strings.Repeat(" ", padding) + line
        }
    }
    return strings.Join(lines, "\n")
}

// visibleWidth returns the visible width of a string (stripping ANSI codes).
func visibleWidth(s string) int {
    var inEscape bool
    n := 0
    for _, r := range s {
        if inEscape {
            if r == 'm' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
                inEscape = false
            }
            continue
        }
        if r == '\x1b' {
            inEscape = true
            continue
        }
        n++
    }
    return n
}
```

Wait, we need `ansi.Truncate` for ANSI stripping. Let me check if it works for this. Actually, we can use the same `visibleWidth` function approach. The ANSI Truncate approach may not work right. Let me just use a simple ANSI-stripping regex.

Actually, looking at `ansi.Truncate`: it truncates a string to a given width, stripping ANSI codes. If we pass a very large limit and `""` for the tail, we get the full visible string. Let me use that.

Actually, `ansi.Truncate(s, maxWidth, tail)` — the `maxWidth` is in display cells. If we pass something larger than the string's visual width, it returns the full visible string. Let me check...

Actually, looking at the signature: `ansi.Truncate(s string, n int, tail string) string`. The `n` is the max width in cells. If the visible width exceeds `n`, it truncates and appends `tail`. If visible width <= n, it returns the original string (with ANSI? or without?).

Let me just use a simple ANSI-stripping function instead of relying on Truncate behavior.

Let me revise the centering approach to use a simple visibleWidth function defined locally:

```go
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func visibleWidth(s string) int {
    return len([]rune(ansiRe.ReplaceAllString(s, "")))
}
```

Actually, that regex doesn't handle all ANSI sequences. Let me use the charmbracelet/x/ansi package's `ansi.Truncate` with a very large limit:

```go
func visibleWidth(s string) int {
    return len([]rune(ansi.Truncate(s, 999999, "")))
}
```

The `Truncate` function with n=999999 and empty tail should just strip ANSI and return the full visible text. Let me use this approach.

- [ ] **Step 4: Modify RenderBytes to add pre/post-processing**

Replace the `RenderBytes` method:

```go
func (tr *TermRenderer) RenderBytes(in []byte) ([]byte, error) {
    // Pre-process: extract <center> blocks
    inputStr := string(in)
    processed, centerBlocks := extractCenterBlocks(inputStr)

    // Render through goldmark
    var buf bytes.Buffer
    err := tr.md.Convert([]byte(processed), &buf)
    if err != nil {
        return nil, err
    }
    result := buf.String()

    // Post-process: render and center each <center> block
    ww := tr.ansiOptions.WordWrap
    if ww <= 0 {
        ww = defaultWidth
    }
    // Block width accounts for margins (like blockStack.Width)
    blockWidth := ww
    if m := tr.ansiOptions.Styles.Document.Margin; m != nil {
        blockWidth -= int(*m) * 2
    }

    for marker, content := range centerBlocks {
        // Recursively render the inner markdown content
        inner, err := Render(content, "")
        if err != nil {
            continue // skip on error
        }
        centered := centerText(inner, blockWidth)
        result = strings.Replace(result, marker, centered, 1)
    }

    return []byte(result), nil
}
```

- [ ] **Step 5: Update the `Render` method too**

Modify `Render` method (line 43-46):

```go
func Render(in string, stylePath string) (string, error) {
    b, err := RenderBytes([]byte(in), stylePath)
    return string(b), err
}
```

No change needed here — it delegates to RenderBytes which already handles it.

Wait, actually `Render` and `RenderBytes` are standalone functions that create their own renderer. The TermRenderer method is `(*TermRenderer).RenderBytes`. Let me check:

- `Render(in, style)` → `RenderBytes(in, stylePath)` (standalone, creates new TermRenderer) → should also do pre/post processing
- `(*TermRenderer).RenderBytes(in)` → uses the TermRenderer's goldmark instance

The standalone functions create a new TermRenderer via `NewTermRenderer(WithStylePath(stylePath))`. So the pre/post-processing should be in `(*TermRenderer).RenderBytes`.

Actually wait, looking at the code more carefully:

```go
func RenderBytes(in []byte, stylePath string) ([]byte, error) {
    r, _ := NewTermRenderer(WithStylePath(stylePath))
    return r.RenderBytes(in)
}
```

So the standalone `RenderBytes` calls `r.RenderBytes(in)` which IS the method on `*TermRenderer`. So modifying the method handles both paths. Good.

- [ ] **Step 6: Verify it compiles**

Run:
```bash
cd /home/q/files/prjkt/glamoured && go build ./...
```

Expected: no errors.

- [ ] **Step 7: Write a quick manual test**

Create a temporary test file:

```go
package main

import (
    "fmt"
    "charm.land/glamour/v2"
)

func main() {
    md := `<center>
**Bold centered text**

![image](https://stuff.charm.sh/charm-badge.jpg)
</center>

Regular paragraph.`
    out, err := glamour.Render(md, "dark")
    if err != nil {
        panic(err)
    }
    fmt.Print(out)
}
```

Run:
```bash
cd /tmp/opencode && cat > center_test.go << 'GOEOF'
package main

import (
    "fmt"
    "strings"
    "charm.land/glamour/v2"
)

func main() {
    md := "# Test\n\n<center>\n**Centered**\n\nNormal text inside\n</center>\n\nRegular paragraph."

    out, err := glamour.Render(md, "dark")
    if err != nil {
        panic(err)
    }
    // Check that the output has centered-looking lines
    lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
    for i, l := range lines {
        fmt.Printf("Line %d: [%d] %s\n", i, len(l), l[:min(20, len(l))])
    }
}
GOEOF
cd /tmp/opencode && go mod init test && go mod edit -replace charm.land/glamour/v2=/home/q/files/prjkt/glamoured && go mod tidy && go run center_test.go
```

- [ ] **Step 8: Commit**

```bash
cd /home/q/files/prjkt/glamoured && git add -A && git commit -m "feat: add <center> HTML block support with markdown rendering"
```

---

### Task 5: Handle `<div align="center">` in the center pre-processor

This is already handled by the regex in Task 4 Step 2:
```go
var centerRe = regexp.MustCompile(`(?is)<(?:center|div\s+align="?center"?)\s*>([\s\S]*?)</(?:center|div)\s*>`)
```

This regex matches both `<center>` and `<div align="center">` (with or without quotes around the value).

- [ ] **Step 1: Verify the regex handles `<div align=center>` (without quotes)**

No code change needed — the regex uses `"?"` to make quotes optional.

- [ ] **Step 2: Verify the regex handles `<DIV ALIGN=CENTER>` (case insensitive)**

No code change needed — the regex uses `(?i)` flag for case insensitivity.

- [ ] **Step 3: Commit (if tests pass — this is already committed with Task 4)**

---

### Task 6: Add tests for center processing

**Files:**
- Modify: `glamour_test.go`

- [ ] **Step 1: Add a test for `<center>` rendering**

Add to `glamour_test.go`:

```go
func TestCenterBlock(t *testing.T) {
    in := "# Title\n\n<center>\n**Centered Bold**\n\n![alt](https://stuff.charm.sh/charm-badge.jpg)\n</center>\n\nFooter"
    r, err := NewTermRenderer(
        WithWordWrap(80),
        WithStandardStyle("dark"),
    )
    if err != nil {
        t.Fatal(err)
    }
    out, err := r.Render(in)
    if err != nil {
        t.Fatal(err)
    }
    if out == "" {
        t.Fatal("expected non-empty output")
    }
    // Verify the centered block marker was replaced
    if strings.Contains(out, "<glamour-center-") {
        t.Fatal("center marker was not replaced in output")
    }
    // Verify the output has content
    if !strings.Contains(out, "Centered") {
        t.Fatal("expected centered content in output")
    }
}
```

- [ ] **Step 2: Run the test**

Run:
```bash
cd /home/q/files/prjkt/glamoured && go test ./... -run TestCenterBlock -v
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/q/files/prjkt/glamoured && git add -A && git commit -m "test: add tests for <center> HTML block rendering"
```

---

### Task 7: Handle `<div>` without align (passthrough)

`<div>` without `align="center"` should be treated as unknown HTML → stripped to text by the existing SanitizeHTML path. This is the default behavior already — no code change needed.

If `<div>` contains markdown-like content, it won't be rendered as markdown (HTML blocks in goldmark are raw text). This is acceptable for initial version.

- [ ] **Step 1: Verify the regex doesn't match `<div>` without align**

The regex requires `align="?center"?` after `<div`, so a plain `<div>` won't be matched. It passes through to the existing SanitizeHTML handler, which strips the tag and keeps the text content.

No code change needed.

- [ ] **Step 2: Run full test suite**

Run:
```bash
cd /home/q/files/prjkt/glamoured && go test ./... -v 2>&1 | tail -30
```

Expected: all tests pass.
