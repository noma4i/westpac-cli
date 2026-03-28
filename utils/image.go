package utils

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"time"

	"github.com/dolmen-go/kittyimg"
	"github.com/nfnt/resize"
)

var imageClient = &http.Client{Timeout: 5 * time.Second}

func RenderImageFromURL(url string, maxWidth int) (string, error) {
	resp, err := imageClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", err
	}

	// Resize to fit terminal width (each cell ~8px wide)
	cellWidth := maxWidth * 8
	if cellWidth > 400 {
		cellWidth = 400
	}
	resized := resize.Resize(uint(cellWidth), 0, img, resize.Lanczos3)

	var buf bytes.Buffer
	if err := kittyimg.Fprint(&buf, resized); err != nil {
		return "", err
	}

	return buf.String(), nil
}
