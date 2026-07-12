# Shields.io Badge Rendering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Render shields.io static badge URLs as one-line ANSI colored badges with optional Nerd Font icons.

**Architecture:** New `ansi/badge.go` file contains URL parsing, color/logo mapping, and badge rendering. `ImageElement.Render` in `ansi/image.go` intercepts shields.io URLs before mosaic rendering. A new `WithShieldsBadges` option controls the feature (default on).

**Tech Stack:** Go, `github.com/charmbracelet/x/mosaic` (existing), ANSI escape codes (existing patterns)

---

### Task 1: Options and Public API

**Files:**
- Modify: `ansi/renderer.go:28` — add `ShieldsBadges bool` field
- Modify: `glamour.go:85-90` — default `ShieldsBadges` to `true`
- Modify: `glamour.go` — add `WithShieldsBadges(enabled bool)` function
- Test: `glamour_test.go` — verify option sets correctly

- [ ] **Step 1: Add ShieldsBadges to Options struct**

```go
// ansi/renderer.go line 28, after ChromaFormatter
	ChromaFormatter  string
	MosaicMaxHeight  int
	ShieldsBadges    bool
```

- [ ] **Step 2: Default ShieldsBadges to true in NewTermRenderer**

```go
// glamour.go line 85-89
		ansiOptions: ansi.Options{
			WordWrap:      defaultWidth,
			MosaicEnabled: true,
			ShieldsBadges: true,
			Styles:        *styles.DefaultStyles["dark"],
		},
```

- [ ] **Step 3: Add WithShieldsBadges function in glamour.go**

Add after `WithMosaicMaxHeight` (line 255):

```go
// WithShieldsBadges enables or disables shields.io badge rendering.
// Badges are rendered as one-line ANSI colored segments instead of mosaic images.
func WithShieldsBadges(enabled bool) TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.ShieldsBadges = enabled
		return nil
	}
}
```

- [ ] **Step 4: Write test for option in glamour_test.go**

```go
func TestWithShieldsBadges(t *testing.T) {
	r, err := NewTermRenderer()
	if err != nil {
		t.Fatal(err)
	}
	if !r.ansiOptions.ShieldsBadges {
		t.Error("expected ShieldsBadges to default to true")
	}

	r2, err := NewTermRenderer(WithShieldsBadges(false))
	if err != nil {
		t.Fatal(err)
	}
	if r2.ansiOptions.ShieldsBadges {
		t.Error("expected ShieldsBadges to be false")
	}
}
```

- [ ] **Step 5: Run tests to verify**

Run: `go test ./... -run TestWithShieldsBadges -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add ansi/renderer.go glamour.go glamour_test.go
git commit -m "feat: add WithShieldsBadges option"
```

---

### Task 2: Badge URL Parsing

**Files:**
- Create: `ansi/badge.go` — `parseShieldsURL` function
- Test: `ansi/badge_test.go` — URL parsing tests

Algorithm for `/badge/<LABEL>-<MESSAGE>-<COLOR>`:
1. Replace `--` with sentinel `\x00` (protect literal dashes)
2. Split by `-`
3. Replace `\x00` with `-` in each part
4. Last part = color, second-to-last = message, remaining joined by `-` = label
5. Decode label and message: `__` → `_`, `_` → ` `

Logo extracted from query param `logo=NAME`.

- [ ] **Step 1: Write test for URL parsing**

Create `ansi/badge_test.go`:

```go
package ansi

import (
	"testing"
)

func TestParseShieldsURL(t *testing.T) {
	tests := []struct {
		url        string
		wantLabel  string
		wantMsg    string
		wantColor  string
		wantLogo   string
		wantOK     bool
	}{
		{
			url:       "https://img.shields.io/badge/Go-1.21-blue",
			wantLabel: "Go",
			wantMsg:   "1.21",
			wantColor: "blue",
			wantOK:    true,
		},
		{
			url:       "https://img.shields.io/badge/License-MIT-brightgreen",
			wantLabel: "License",
			wantMsg:   "MIT",
			wantColor: "brightgreen",
			wantOK:    true,
		},
		{
			url:       "https://img.shields.io/badge/Go_Releases-1.21--beta-brightgreen",
			wantLabel: "Go Releases",
			wantMsg:   "1.21-beta",
			wantColor: "brightgreen",
			wantOK:    true,
		},
		{
			url:       "https://img.shields.io/badge/hello__world-foo-ff69b4",
			wantLabel: "hello_world",
			wantMsg:   "foo",
			wantColor: "ff69b4",
			wantOK:    true,
		},
		{
			url:       "https://img.shields.io/badge/Go-1.21-blue?logo=go&style=flat",
			wantLabel: "Go",
			wantMsg:   "1.21",
			wantColor: "blue",
			wantLogo:  "go",
			wantOK:    true,
		},
		{
			url:   "https://example.com/image.png",
			wantOK: false,
		},
		{
			url:   "https://img.shields.io/endpoint?url=...",
			wantOK: false,
		},
		{
			url:   "https://img.shields.io/badge/",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		label, msg, color, logo, ok := parseShieldsURL(tt.url)
		if ok != tt.wantOK {
			t.Errorf("parseShieldsURL(%q) ok = %v, want %v", tt.url, ok, tt.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if label != tt.wantLabel {
			t.Errorf("parseShieldsURL(%q) label = %q, want %q", tt.url, label, tt.wantLabel)
		}
		if msg != tt.wantMsg {
			t.Errorf("parseShieldsURL(%q) msg = %q, want %q", tt.url, msg, tt.wantMsg)
		}
		if color != tt.wantColor {
			t.Errorf("parseShieldsURL(%q) color = %q, want %q", tt.url, color, tt.wantColor)
		}
		if logo != tt.wantLogo {
			t.Errorf("parseShieldsURL(%q) logo = %q, want %q", tt.url, logo, tt.wantLogo)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ansi/ -run TestParseShieldsURL -v`
Expected: FAIL (parseShieldsURL not defined)

- [ ] **Step 3: Implement `parseShieldsURL`**

Create `ansi/badge.go`:

```go
package ansi

import (
	"net/url"
	"strings"
)

// parseShieldsURL parses a shields.io static badge URL.
// Returns label, message, color, logo name, and whether parsing succeeded.
func parseShieldsURL(rawURL string) (label, message, color, logo string, ok bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", "", false
	}
	if u.Host != "img.shields.io" {
		return "", "", "", "", false
	}
	path := strings.TrimPrefix(u.Path, "/badge/")
	if path == u.Path || path == "" {
		return "", "", "", "", false
	}
	// Protect literal dashes (--) before splitting on single dash
	path = strings.ReplaceAll(path, "--", "\x00")
	parts := strings.Split(path, "-")
	// Restore dashes in each part
	for i := range parts {
		parts[i] = strings.ReplaceAll(parts[i], "\x00", "-")
	}
	if len(parts) < 3 {
		return "", "", "", "", false
	}
	color = parts[len(parts)-1]
	message = parts[len(parts)-2]
	label = strings.Join(parts[:len(parts)-2], "-")
	// Decode underscores in label and message
	label = decodeShieldsValue(label)
	message = decodeShieldsValue(message)
	// Extract logo from query params
	logo = u.Query().Get("logo")
	return label, message, color, logo, true
}

// decodeShieldsValue decodes a single badge component value:
//   __ → literal underscore
//   _  → space
func decodeShieldsValue(s string) string {
	s = strings.ReplaceAll(s, "__", "\x01")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "\x01", "_")
	return s
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./ansi/ -run TestParseShieldsURL -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ansi/badge.go ansi/badge_test.go
git commit -m "feat: add parseShieldsURL for badge URL parsing"
```

---

### Task 3: Color and Logo Mappings

**Files:**
- Modify: `ansi/badge.go` — add color map, hex→ANSI conversion, logo→nerdfont map
- Test: `ansi/badge_test.go` — test color and logo lookups

- [ ] **Step 1: Write tests for color and logo mappings**

Append to `ansi/badge_test.go`:

```go
func TestBadgeNamedColor(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"brightgreen", 2},
		{"blue", 32},
		{"red", 196},
		{"unknown", 240},
	}
	for _, tt := range tests {
		got := badgeNamedColor(tt.name)
		if got != tt.want {
			t.Errorf("badgeNamedColor(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestHexToANSI(t *testing.T) {
	tests := []struct {
		hex  string
		want int
	}{
		{"000000", 16},
		{"ffffff", 231},
		{"ff0000", 196},
	}
	for _, tt := range tests {
		got := hexToANSI(tt.hex)
		if got != tt.want {
			t.Errorf("hexToANSI(%q) = %d, want %d", tt.hex, got, tt.want)
		}
	}
}

func TestLogoNerdIcon(t *testing.T) {
	tests := []struct {
		logo string
		want string
	}{
		{"go", "\ue61b"},
		{"", ""},
		{"unknown", "\uf0a3"},
	}
	for _, tt := range tests {
		got := logoNerdIcon(tt.logo)
		if got != tt.want {
			t.Errorf("logoNerdIcon(%q) = %q, want %q", tt.logo, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ansi/ -run "TestBadgeNamedColor|TestHexToANSI|TestLogoNerdIcon" -v`
Expected: FAIL (functions not defined)

- [ ] **Step 3: Implement color and logo helpers**

Append to `ansi/badge.go`:

```go
// Named shields.io colors mapped to ANSI 256-color codes.
// Derived from https://shields.io/badges
var badgeNamedColors = map[string]int{
	"brightgreen": 2,   // #44CC11
	"green":       106, // #97CA00
	"yellowgreen": 142, // #A4A61D
	"yellow":      214, // #DFB317
	"orange":      208, // #FE7D37
	"red":         196, // #E05D44
	"blue":        32,  // #007EC6
	"lightgrey":   250, // #9F9F9F
	"grey":        240, // #555555
	"blueviolet":  99,  // #800080
	"pink":        205, // #E04E8C
	"cyan":        39,  // #00BFFF
	"purple":      93,  // #8A2BE2
}

func badgeNamedColor(name string) int {
	if c, ok := badgeNamedColors[name]; ok {
		return c
	}
	return 240 // default dark grey
}

func hexToANSI(hex string) int {
	if len(hex) != 6 {
		return 240
	}
	r := parseHexByte(hex[0:2])
	g := parseHexByte(hex[2:4])
	b := parseHexByte(hex[4:6])
	return closestANSI256(r, g, b)
}

func parseHexByte(s string) byte {
	b, _ := strconv.ParseUint(s, 16, 8)
	return byte(b)
}

// closestANSI256 returns the closest ANSI 256-color code to the given RGB.
func closestANSI256(r, g, b byte) int {
	// 6×6×6 color cube: 16 + 36*r + 6*g + b
	cr := int(r) * 5 / 255
	cg := int(g) * 5 / 255
	cb := int(b) * 5 / 255
	return 16 + 36*cr + 6*cg + cb
}

// logoNerdIcon maps shields.io logo names to Nerd Font Unicode codepoints.
func logoNerdIcon(logo string) string {
	if logo == "" {
		return ""
	}
	if icon, ok := badgeLogoIcons[strings.ToLower(logo)]; ok {
		return icon
	}
	return "\uf0a3" // generic certificate icon
}

// badgeLogoIcons maps logo names to Nerd Font icon strings.
// Uses Font Awesome and Devicons codepoints from the Nerd Font PUA range.
var badgeLogoIcons = map[string]string{
	"go":         "\ue61b", // nf-dev-go
	"golang":     "\ue61b", // nf-dev-go
	"rust":       "\ue7a8", // nf-dev-rust
	"python":     "\ue73c", // nf-dev-python
	"node":       "\ue718", // nf-dev-nodejs
	"nodejs":     "\ue718", // nf-dev-nodejs
	"javascript": "\ue74e", // nf-dev-javascript
	"js":         "\ue74e", // nf-dev-javascript
	"typescript": "\ue628", // nf-dev-typescript-icon
	"ts":         "\ue628", // nf-dev-typescript-icon
	"docker":     "\ue7b0", // nf-dev-docker
	"github":     "\uf09b", // nf-fa-github
	"git":        "\uf1d3", // nf-fa-git-alt
	"react":      "\ue7ba", // nf-dev-react
	"vue":        "\ue6d0", // nf-dev-vuejs
	"angular":    "\ue753", // nf-dev-angular
	"ruby":       "\ue739", // nf-dev-ruby
	"java":       "\ue738", // nf-dev-java
	"kotlin":     "\ue634", // nf-dev-kotlin
	"swift":      "\ue755", // nf-dev-swift
	"php":        "\ue73d", // nf-dev-php
	"c":          "\ue708", // nf-dev-c
	"cpp":        "\ue708", // nf-dev-c
	"c++":        "\ue708", // nf-dev-c
	"zig":        "\ue6a9", // nf-dev-zig
	"deno":       "\ue60f", // nf-dev-deno
	"discord":    "\uf392", // nf-fa-discord
	"slack":      "\uf198", // nf-fa-slack
	"nginx":      "\ue776", // nf-dev-nginx
	"redis":      "\ue76d", // nf-dev-redis
	"postgresql": "\ue76e", // nf-dev-postgresql
	"postgres":   "\ue76e", // nf-dev-postgresql
	"mysql":      "\ue704", // nf-dev-mysql
	"mongodb":    "\ue7a4", // nf-dev-mongodb
	"aws":        "\ue7ad", // nf-dev-aws
	"amazon":     "\ue7ad", // nf-dev-aws
	"linkedin":   "\uf0e1", // nf-fa-linkedin
	"twitter":    "\uf099", // nf-fa-twitter
	"x":          "\ue619", // nf-dev-x
	"youtube":    "\uf167", // nf-fa-youtube
	"npm":        "\ue71e", // nf-dev-npm
	"license":    "\uf0a3", // generic
}
```

Needs import: `"strconv"` added to badge.go imports.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./ansi/ -run "TestBadgeNamedColor|TestHexToANSI|TestLogoNerdIcon" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ansi/badge.go ansi/badge_test.go
git commit -m "feat: add badge color and logo mappings"
```

---

### Task 4: Badge Rendering Function

**Files:**
- Modify: `ansi/badge.go` — add `renderBadge` function
- Test: `ansi/badge_test.go` — test render output

- [ ] **Step 1: Write test for renderBadge**

Append to `ansi/badge_test.go`:

```go
func TestRenderBadge(t *testing.T) {
	t.Run("without icon", func(t *testing.T) {
		var buf strings.Builder
		renderBadge(&buf, "Go", "1.21", 32, "")
		out := buf.String()
		if !strings.Contains(out, "Go") || !strings.Contains(out, "1.21") {
			t.Errorf("badge output missing label/message: %q", out)
		}
		if !strings.Contains(out, "\x1b[") {
			t.Errorf("badge output missing ANSI escapes: %q", out)
		}
	})

	t.Run("with icon", func(t *testing.T) {
		var buf strings.Builder
		renderBadge(&buf, "Go", "1.21", 32, "\ue61b")
		out := buf.String()
		if !strings.Contains(out, "\ue61b") {
			t.Errorf("badge output missing icon: %q", out)
		}
	})

	t.Run("line break prefix", func(t *testing.T) {
		var buf strings.Builder
		renderBadge(&buf, "Go", "1.21", 32, "")
		out := buf.String()
		if out[0] != '\n' {
			t.Errorf("badge should start with newline, got: %q", out)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ansi/ -run TestRenderBadge -v`
Expected: FAIL (renderBadge not defined)

- [ ] **Step 3: Implement renderBadge**

Append to `ansi/badge.go`:

```go
import "fmt"

// renderBadge writes a shields.io-style badge to w.
// Format: \n[grey bg white fg] icon? LABEL [color bg white fg] MESSAGE [reset]
func renderBadge(w io.Writer, label, message string, color int, icon string) {
	labelBg := 240 // dark grey
	fg := 97       // bright white
	iconPart := icon
	if iconPart != "" {
		iconPart += " "
	}
	_, _ = fmt.Fprintf(w, "\n\033[48;5;%d;38;5;%dm %s%s \033[0m\033[48;5;%d;38;5;%dm %s \033[0m",
		labelBg, fg, iconPart, label, color, fg, message)
}
```

Add `"fmt"` to imports if not already present. Note: `badge.go` currently only imports `"net/url"` and `"strings"`. Now needs: `"fmt"`, `"io"`, `"strconv"`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./ansi/ -run TestRenderBadge -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ansi/badge.go ansi/badge_test.go
git commit -m "feat: add badge rendering function"
```

---

### Task 5: Integration into ImageElement

**Files:**
- Modify: `ansi/image.go` — add `tryRenderBadge` method, call from `Render`
- Test: `ansi/badge_test.go` — integration test

- [ ] **Step 1: Write integration test**

Append to `ansi/badge_test.go`:

```go
func TestTryRenderBadge(t *testing.T) {
	t.Run("renders shield badge", func(t *testing.T) {
		e := &ImageElement{
			URL: "https://img.shields.io/badge/Go-1.21-blue",
		}
		ctx := NewRenderContext(Options{
			ShieldsBadges: true,
			MosaicEnabled: false,
		})
		var buf strings.Builder
		result := e.tryRenderBadge(&buf, ctx)
		if !result {
			t.Fatal("tryRenderBadge returned false")
		}
		out := buf.String()
		if !strings.Contains(out, "Go") || !strings.Contains(out, "1.21") {
			t.Errorf("badge output missing content: %q", out)
		}
	})

	t.Run("skips non-shield URL", func(t *testing.T) {
		e := &ImageElement{
			URL: "https://example.com/image.png",
		}
		ctx := NewRenderContext(Options{
			ShieldsBadges: true,
		})
		var buf strings.Builder
		result := e.tryRenderBadge(&buf, ctx)
		if result {
			t.Fatal("tryRenderBadge should return false for non-shield URL")
		}
	})

	t.Run("skips when disabled", func(t *testing.T) {
		e := &ImageElement{
			URL: "https://img.shields.io/badge/Go-1.21-blue",
		}
		ctx := NewRenderContext(Options{
			ShieldsBadges: false,
		})
		var buf strings.Builder
		result := e.tryRenderBadge(&buf, ctx)
		if result {
			t.Fatal("tryRenderBadge should return false when disabled")
		}
	})

	t.Run("uses Nerd Font icon when enabled", func(t *testing.T) {
		e := &ImageElement{
			URL: "https://img.shields.io/badge/Go-1.21-blue?logo=go",
		}
		ctx := NewRenderContext(Options{
			ShieldsBadges: true,
			NerdFontIcons: true,
		})
		var buf strings.Builder
		result := e.tryRenderBadge(&buf, ctx)
		if !result {
			t.Fatal("tryRenderBadge returned false")
		}
		out := buf.String()
		if !strings.Contains(out, "\ue61b") {
			t.Errorf("expected Nerd Font icon in output: %q", out)
		}
	})

	t.Run("no icon when Nerd Font disabled", func(t *testing.T) {
		e := &ImageElement{
			URL: "https://img.shields.io/badge/Go-1.21-blue?logo=go",
		}
		ctx := NewRenderContext(Options{
			ShieldsBadges: true,
			NerdFontIcons: false,
		})
		var buf strings.Builder
		result := e.tryRenderBadge(&buf, ctx)
		if !result {
			t.Fatal("tryRenderBadge returned false")
		}
		out := buf.String()
		if strings.Contains(out, "\ue61b") {
			t.Errorf("expected no Nerd Font icon when disabled: %q", out)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ansi/ -run TestTryRenderBadge -v`
Expected: FAIL (tryRenderBadge not defined)

- [ ] **Step 3: Add tryRenderBadge to ImageElement and wire into Render**

In `ansi/image.go`, add before `tryRenderMosaic`:

```go
func (e *ImageElement) tryRenderBadge(w io.Writer, ctx RenderContext) bool {
	if !ctx.options.ShieldsBadges || e.TextOnly {
		return false
	}
	u := resolveRelativeURL(e.BaseURL, e.URL)
	label, msg, color, logo, ok := parseShieldsURL(u)
	if !ok {
		return false
	}
	ansiColor := badgeNamedColor(color)
	if len(color) == 6 && isHex(color) {
		ansiColor = hexToANSI(color)
	}
	var icon string
	if ctx.options.NerdFontIcons {
		icon = logoNerdIcon(logo)
	}
	renderBadge(w, label, msg, ansiColor, icon)
	return true
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
```

In `Render` method, add tryRenderBadge call before tryRenderMosaic:

```go
func (e *ImageElement) Render(w io.Writer, ctx RenderContext) error {
	if e.tryRenderBadge(w, ctx) {
		return nil
	}
	if e.tryRenderMosaic(w, ctx) {
		return nil
	}
	// ... rest unchanged
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./ansi/ -run TestTryRenderBadge -v`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `go test ./...`
Expected: All tests pass

- [ ] **Step 6: Commit**

```bash
git add ansi/image.go ansi/badge_test.go
git commit -m "feat: integrate badge rendering into ImageElement"
```

---

### Task 6: README Update

**Files:**
- Modify: `README.md` — document shields badge feature

- [ ] **Step 1: Add shields.io section to README**

Add after Nerd Font Icons section (~line 92):

```markdown
### Shields.io Badge Rendering

Detect `img.shields.io` badge URLs and render them as styled one-line ANSI badges:

```go
r, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithMosaic(false), // badges work independently
)

out, _ := r.Render("![Go](https://img.shields.io/badge/Go-1.21-blue)")
fmt.Print(out)
```

Output shows: `▐ Go ▐▐ 1.21 ▐` with colored backgrounds.

- Existing option `WithNerdFontIcons()` adds a logo icon when the badge URL includes a `logo` parameter (e.g., `?logo=go`)
- Disable badge rendering: `glamour.WithShieldsBadges(false)`
```

Add to Custom Renderer Options table:

```markdown
| `WithShieldsBadges(enabled)` | Enable/disable shields.io badge rendering (default: enabled) |
```

- [ ] **Step 2: Verify build**

Run: `go build ./... && go vet ./...`
Expected: no output

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: shields.io badge rendering"
```
