package ansi

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // required for JPEG decoding
	_ "image/png"  // required for PNG decoding
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/x/mosaic"
)

// pixelsPerCell is the approximate width of a terminal character cell in CSS pixels.
// Used to convert HTML <img width> attribute values to character cells.
const pixelsPerCell = 10

// An ImageElement is used to render images elements.
type ImageElement struct {
	Text     string
	BaseURL  string
	URL      string
	Child    ElementRenderer
	TextOnly bool
	Width    int
	Height   int
}

func loadImage(url string) (image.Image, error) {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("glamour: error creating request: %w", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("glamour: error fetching image: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, http.ErrMissingFile
		}
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("glamour: error decoding image: %w", err)
		}
		return img, nil
	}
	f, err := os.Open(url)
	if err != nil {
		return nil, fmt.Errorf("glamour: error opening image file: %w", err)
	}
	defer func() { _ = f.Close() }()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("glamour: error decoding image: %w", err)
	}
	return img, nil
}

func (e *ImageElement) tryRenderMosaic(w io.Writer, ctx RenderContext) bool {
	if !ctx.options.MosaicEnabled || e.TextOnly {
		return false
	}
	u := resolveRelativeURL(e.BaseURL, e.URL)
	img, err := loadImage(u)
	if err != nil {
		return false
	}
	width := e.Width
	if width > 0 {
		width /= pixelsPerCell
		if width < 1 {
			width = 1
		}
		if ctx.options.MosaicWidth > 0 && width > ctx.options.MosaicWidth {
			width = ctx.options.MosaicWidth
		}
	}
	if width <= 0 {
		width = ctx.options.MosaicWidth
	}
	if width <= 0 {
		width = ctx.options.WordWrap / 2
		if width < 20 {
			width = 20
		}
	}
	maxH := ctx.options.MaxImageHeight
	art := renderMosaic(img, width, maxH)
	el := &BaseElement{
		Token: art,
		Style: ctx.options.Styles.Image,
	}
	return el.Render(w, ctx) == nil
}

func renderMosaic(img image.Image, widthCells, maxHeight int) string {
	m := mosaic.New()
	m = m.Width(widthCells * 2)
	art := m.Render(img)
	if maxHeight > 0 {
		lines := strings.Count(art, "\n")
		if lines > maxHeight {
			newWidth := widthCells * maxHeight / lines
			if newWidth < 1 {
				newWidth = 1
			}
			m = mosaic.New()
			m = m.Width(newWidth * 2)
			art = m.Render(img)
		}
	}
	return art
}

// Render renders an ImageElement.
func (e *ImageElement) Render(w io.Writer, ctx RenderContext) error {
	if e.tryRenderMosaic(w, ctx) {
		return nil
	}

	// Make OSC 8 hyperlink token.
	hyperlink, resetHyperlink, _ := makeHyperlink(e.URL)

	style := ctx.options.Styles.ImageText
	if e.TextOnly {
		style.Format = strings.TrimSuffix(style.Format, " →")
	}

	if len(e.Text) > 0 {
		token := hyperlink + e.Text + resetHyperlink
		el := &BaseElement{
			Token: token,
			Style: style,
		}
		err := el.Render(w, ctx)
		if err != nil {
			return err
		}
	}

	if e.TextOnly {
		return nil
	}

	if len(e.URL) > 0 {
		token := hyperlink + resolveRelativeURL(e.BaseURL, e.URL) + resetHyperlink
		el := &BaseElement{
			Token:  token,
			Prefix: " ",
			Style:  ctx.options.Styles.Image,
		}
		err := el.Render(w, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
