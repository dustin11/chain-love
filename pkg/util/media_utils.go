package util

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"path/filepath"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/disintegration/imaging"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	// DefaultJPEGMime 默认 JPEG 媒体类型。
	DefaultJPEGMime = "image/jpeg"
	// DefaultUTF8TextMime 默认 UTF-8 文本媒体类型。
	DefaultUTF8TextMime = "text/plain; charset=utf-8"
)

// ImageProcessOptions 定义图片通用处理参数。
type ImageProcessOptions struct {
	// MaxBytes 限制原始文件最大字节数；小于等于 0 表示不限制。
	MaxBytes int64
	// MaxDimension 限制原图最大边；小于等于 0 表示不缩放。
	MaxDimension int
	// ThumbMaxDimension 限制缩略图最大边；小于等于 0 表示不生成缩略图。
	ThumbMaxDimension int
	// JPEGQuality 控制原图 JPEG 质量；小于等于 0 使用默认值。
	JPEGQuality int
	// ThumbJPEGQuality 控制缩略图 JPEG 质量；小于等于 0 使用默认值。
	ThumbJPEGQuality int
}

// ProcessedImage 保存通用图片处理结果。
type ProcessedImage struct {
	// Original 处理后的原图字节。
	Original []byte
	// Thumb 处理后的缩略图字节。
	Thumb []byte
	// Mime 处理后的媒体类型。
	Mime string
	// Ext 推荐保存的文件后缀。
	Ext string
	// Width 原图宽度。
	Width int
	// Height 原图高度。
	Height int
}

// NormalizeTextFileBytes 把文本类文件尽量归一化为 UTF-8。
func NormalizeTextFileBytes(data []byte, filename string, mime string) ([]byte, string, bool) {
	if !shouldNormalizeTextFile(data, filename, mime) {
		return nil, "", false
	}
	decodedText, ok := decodeTextFileBytes(data)
	if !ok {
		return nil, "", false
	}
	return []byte(decodedText), DefaultUTF8TextMime, true
}

// ProcessImageBytes 统一处理图片字节，完成校验、缩放和编码。
func ProcessImageBytes(data []byte, options ImageProcessOptions) (*ProcessedImage, error) {
	if len(data) == 0 {
		return nil, errors.New("上传文件为空")
	}
	if options.MaxBytes > 0 && int64(len(data)) > options.MaxBytes {
		return nil, errors.New("图片超过最大限制")
	}
	contentType := http.DetectContentType(data[:minInt(len(data), 512)])
	if !strings.HasPrefix(contentType, "image/") {
		return nil, errors.New("不是图片文件")
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if config.Width > 10000 || config.Height > 10000 {
		return nil, errors.New("图片分辨率太大")
	}

	sourceImage, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if options.MaxDimension > 0 &&
		(sourceImage.Bounds().Dx() > options.MaxDimension ||
			sourceImage.Bounds().Dy() > options.MaxDimension) {
		sourceImage = imaging.Resize(
			sourceImage,
			options.MaxDimension,
			0,
			imaging.CatmullRom,
		)
	}

	jpegQuality := options.JPEGQuality
	if jpegQuality <= 0 {
		jpegQuality = 82
	}
	thumbJPEGQuality := options.ThumbJPEGQuality
	if thumbJPEGQuality <= 0 {
		thumbJPEGQuality = 78
	}

	mime := DefaultJPEGMime
	ext := ".jpg"
	var original bytes.Buffer
	if HasImageAlpha(sourceImage) {
		mime = "image/png"
		ext = ".png"
		if err := png.Encode(&original, sourceImage); err != nil {
			return nil, err
		}
	} else if err := jpeg.Encode(&original, sourceImage, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return nil, err
	}

	var thumbBytes []byte
	if options.ThumbMaxDimension > 0 {
		thumbImage := imaging.Thumbnail(
			sourceImage,
			options.ThumbMaxDimension,
			options.ThumbMaxDimension,
			imaging.CatmullRom,
		)
		var thumb bytes.Buffer
		if mime == "image/png" {
			if err := png.Encode(&thumb, thumbImage); err != nil {
				return nil, err
			}
		} else if err := jpeg.Encode(&thumb, thumbImage, &jpeg.Options{Quality: thumbJPEGQuality}); err != nil {
			return nil, err
		}
		thumbBytes = thumb.Bytes()
	}

	return &ProcessedImage{
		Original: original.Bytes(),
		Thumb:    thumbBytes,
		Mime:     mime,
		Ext:      ext,
		Width:    sourceImage.Bounds().Dx(),
		Height:   sourceImage.Bounds().Dy(),
	}, nil
}

// HasImageAlpha 检测图片是否包含透明通道。
func HasImageAlpha(img image.Image) bool {
	switch typedImage := img.(type) {
	case *image.NRGBA:
		for index := 3; index < len(typedImage.Pix); index += 4 {
			if typedImage.Pix[index] != 0xFF {
				return true
			}
		}
		return false
	case *image.NRGBA64:
		for index := 7; index < len(typedImage.Pix); index += 8 {
			if typedImage.Pix[index-1] != 0xFF || typedImage.Pix[index] != 0xFF {
				return true
			}
		}
		return false
	default:
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y += maxInt(1, bounds.Dy()/10) {
			for x := bounds.Min.X; x < bounds.Max.X; x += maxInt(1, bounds.Dx()/10) {
				_, _, _, alpha := img.At(x, y).RGBA()
				if alpha != 0xFFFF {
					return true
				}
			}
		}
		return false
	}
}

func shouldNormalizeTextFile(data []byte, filename string, mime string) bool {
	if len(data) == 0 {
		return false
	}
	normalizedMime := strings.ToLower(strings.TrimSpace(strings.Split(mime, ";")[0]))
	if strings.HasPrefix(normalizedMime, "text/") {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(filepath.Ext(filename))) {
	case ".lrc", ".txt", ".md", ".json", ".csv", ".xml", ".yaml", ".yml", ".ini", ".log":
		return true
	default:
		return false
	}
}

func decodeTextFileBytes(data []byte) (string, bool) {
	trimmedBom := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(trimmedBom) {
		return string(trimmedBom), true
	}
	if decoded, ok := decodeUTF16TextFileBytes(data); ok {
		return decoded, true
	}
	if decoded, ok := decodeGB18030TextFileBytes(data); ok {
		return decoded, true
	}
	return "", false
}

func decodeUTF16TextFileBytes(data []byte) (string, bool) {
	if len(data) < 2 {
		return "", false
	}
	var words []uint16
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}):
		payload := data[2:]
		if len(payload)%2 != 0 {
			return "", false
		}
		words = make([]uint16, 0, len(payload)/2)
		for index := 0; index < len(payload); index += 2 {
			words = append(words, uint16(payload[index])|uint16(payload[index+1])<<8)
		}
	case bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		payload := data[2:]
		if len(payload)%2 != 0 {
			return "", false
		}
		words = make([]uint16, 0, len(payload)/2)
		for index := 0; index < len(payload); index += 2 {
			words = append(words, uint16(payload[index])<<8|uint16(payload[index+1]))
		}
	default:
		return "", false
	}
	return string(utf16.Decode(words)), true
}

func decodeGB18030TextFileBytes(data []byte) (string, bool) {
	decoded, _, err := transform.String(
		simplifiedchinese.GB18030.NewDecoder(),
		string(data),
	)
	if err != nil {
		return "", false
	}
	if !utf8.ValidString(decoded) {
		return "", false
	}
	return decoded, true
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
