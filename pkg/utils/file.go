package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

// GetMD5ByMultipart get file digest value through file stream
// TODO: change it to sha256
func GetMD5ByMultipart(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint: errcheck

	md5h := md5.New() // nolint:gosec
	_, err = io.Copy(md5h, f)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(md5h.Sum(nil)), nil
}

// Scale Scale
func Scale(in io.Reader, out io.Writer, width, height, quality int) error {
	origin, fm, err := image.Decode(in)
	if err != nil {
		return err
	}

	canvas := imaging.Resize(origin, width, height, imaging.Lanczos)

	switch fm {
	case "jpeg":
		return jpeg.Encode(out, canvas, &jpeg.Options{Quality: quality})
	case "png":
		return png.Encode(out, canvas)
	case "gif":
		return gif.Encode(out, canvas, &gif.Options{})
	case "bmp":
		return bmp.Encode(out, canvas)
	case "tiff":
		return tiff.Encode(out, canvas, &tiff.Options{})
	default:
		return errors.New("ERROR FORMAT")
	}
}

// ExistBucket ExistBucket
func ExistBucket(buckets map[string]string, target string) bool {
	for _, bucket := range buckets {
		if bucket == target {
			return true
		}
	}

	return false
}

// GetExt get extension of the file
func GetExt(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	return strings.TrimPrefix(ext, ".")
}

// URLQueryEscape escapes the original string.
func URLQueryEscape(origin string) string {
	escaped := url.QueryEscape(origin)
	escaped = strings.Replace(escaped, "%2F", "/", -1)
	escaped = strings.Replace(escaped, "%3D", "=", -1)
	escaped = strings.Replace(escaped, "+", "%20", -1)
	return escaped
}

// Split Split
func Split(str string, sep string) (string, string) {
	arr := strings.SplitN(str, sep, 2)
	return arr[0], arr[1]
}
