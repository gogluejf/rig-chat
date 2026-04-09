package chat

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ImageMIME returns the MIME type for an image file based on extension
func ImageMIME(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

// EncodeImageDataURL reads an image file and returns a base64 data URL
func EncodeImageDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read image %s: %w", path, err)
	}
	mime := ImageMIME(path)
	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, b64), nil
}

// BuildMultimodalContent creates the OpenAI multimodal content array for a message with an image
func BuildMultimodalContent(text, imagePath string) ([]ContentPart, error) {
	dataURL, err := EncodeImageDataURL(imagePath)
	if err != nil {
		return nil, err
	}
	return []ContentPart{
		{Type: "image_url", ImageURL: &ImageURL{URL: dataURL}},
		{Type: "text", Text: text},
	}, nil
}
