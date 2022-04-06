package decompress

import (
	"fmt"
	"mime/multipart"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// Decompress Decompress.
type Decompress interface {
	Name() string
	SetValue(d *Decompressor)
	Walk(file multipart.File) (string, error)
	Unarchive(file multipart.File, dst string) error
}

var decompressors = []Decompress{
	&ZipDecompress{},
}

// Decompressor Decompressor.
type Decompressor struct {
	decompress map[string]Decompress
}

// NewDecompressor NewDecompressor.
func NewDecompressor() *Decompressor {
	dp := &Decompressor{
		decompress: map[string]Decompress{},
	}

	for _, decompressor := range decompressors {
		dp.decompress[decompressor.Name()] = decompressor
	}

	return dp
}

// GetDecompress GetDecompress.
func (d *Decompressor) GetDecompress(fileExt string) (Decompress, error) {
	if dp, ok := d.decompress[fileExt]; ok {
		dp.SetValue(d)
		return dp, nil
	}
	return nil, fmt.Errorf("unsupported file extension: %s", fileExt)
}

const (
	indexHTML = "index.html"
	indexHTM  = "index.htm"
)

const (
	fileMode = 0o755
)

func filterMacImpact(fileName string) bool {
	res := strings.Contains(fileName, "__MACOSX") ||
		strings.Contains(fileName, ".DS_Store")
	return res
}

func decoding(str string) (string, error) {
	data := []byte(str)
	switch {
	case isUtf8(data):
		return str, nil
	case isGBK(data):
		utf8Data, err := simplifiedchinese.GBK.NewDecoder().Bytes(data)
		return string(utf8Data), err
	default:
		return str, nil
	}
}

func isGBK(data []byte) bool {
	length, i := len(data), 0
	for i < length {
		if data[i] <= 0xff {
			i++
			continue
		}
		// Use double byte encoding for greater than 127
		if data[i] >= 0x81 && data[i] <= 0xfe && data[i+1] >= 0x40 &&
			data[i+1] <= 0xfe && data[i+1] != 0xf7 {
			i += 2
			continue
		}
		return false
	}
	return true
}

func isUtf8(b []byte) bool {
	return utf8.Valid(b)
}
