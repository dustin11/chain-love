package util

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestNormalizeTextFileBytesConvertsGBKToUTF8(t *testing.T) {
	sourceText := "[00:00.50]给我一个理由忘记\n"
	data, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte(sourceText))
	if err != nil {
		t.Fatalf("encode gb18030 text: %v", err)
	}

	normalizedData, normalizedMime, normalized := NormalizeTextFileBytes(
		data,
		"lyrics.lrc",
		"text/plain",
	)
	if !normalized {
		t.Fatalf("expected text file to be normalized")
	}
	if normalizedMime != DefaultUTF8TextMime {
		t.Fatalf("unexpected mime: %q", normalizedMime)
	}
	if string(normalizedData) != sourceText {
		t.Fatalf("expected utf-8 normalized text, got %q", string(normalizedData))
	}
}

func TestProcessImageBytesBuildsThumbAndKeepsTransparency(t *testing.T) {
	sourceImage := image.NewNRGBA(image.Rect(0, 0, 24, 24))
	sourceImage.Set(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 128})

	var pngBytes bytes.Buffer
	if err := png.Encode(&pngBytes, sourceImage); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	processedImage, err := ProcessImageBytes(pngBytes.Bytes(), ImageProcessOptions{
		MaxBytes:          int64(len(pngBytes.Bytes()) + 1),
		MaxDimension:      1600,
		ThumbMaxDimension: 12,
	})
	if err != nil {
		t.Fatalf("process image bytes: %v", err)
	}
	if processedImage.Mime != "image/png" {
		t.Fatalf("unexpected mime: %q", processedImage.Mime)
	}
	if processedImage.Ext != ".png" {
		t.Fatalf("unexpected ext: %q", processedImage.Ext)
	}
	if len(processedImage.Original) == 0 {
		t.Fatalf("expected original bytes")
	}
	if len(processedImage.Thumb) == 0 {
		t.Fatalf("expected thumb bytes")
	}
}
