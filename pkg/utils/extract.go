package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
)

var (
	attrs    = []string{"href", "src"}
	prefixes = []string{"http://", "https://", "javascript:", "#"}
)

// Blob parsing structure
type Blob struct {
	AppID    string
	MD5      string
	FileName string
}

// ReplaceAttr replace label properties
func ReplaceAttr(blob Blob, blobTemplate, subFilePath, dst string) (*bytes.Buffer, error) {
	// read file
	file, err := os.Open(subFilePath) // nolint: gosec
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint: errcheck, gosec

	dom, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return nil, err
	}

	for _, attr := range attrs {
		selector := fmt.Sprintf("[%s]", attr)
		dom.Find(selector).Each(func(i int, selection *goquery.Selection) {
			val, exists := selection.Attr(attr)
			if !exists || val == "" || skipPrefix(val) {
				return
			}
			// remove the file name and temporary directory
			blob.FileName = replacePath(subFilePath, val, dst)

			value := ExecuteURL(blob, blobTemplate)

			selection.SetAttr(attr, value)
		})
	}
	ret, _ := dom.Html()
	buf := bytes.NewBufferString(ret)
	return buf, nil
}

func replacePath(subFilePath, val, dst string) string {
	dir := cleanDir(subFilePath, dst)
	valArr := strings.Split(val, "/")
	for _, val := range valArr {
		index := strings.LastIndex(dir, "/")
		switch val {
		case ".", "/":
		case "..":
			dir = dir[:index]
		default:
			dir = filepath.Join(dir, val)
		}
	}
	return dir
}

func cleanDir(subFilePath, dst string) string {
	dir := filepath.Dir(subFilePath)
	dir = strings.Replace(dir, dst, "", 1)
	if dir != "" {
		dir = dir[1:]
	}
	return dir
}

func skipPrefix(val string) bool {
	flag := false
	for _, prefix := range prefixes {
		if strings.HasPrefix(val, prefix) {
			flag = true
			break
		}
	}
	return flag
}

// ExecuteURL ExecuteURL
func ExecuteURL(blob Blob, url string) string {
	var buf bytes.Buffer
	t, _ := template.New("").Parse(url)
	_ = t.Execute(&buf, blob)
	return buf.String()
}
