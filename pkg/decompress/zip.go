package decompress

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/quanxiang-cloud/fileserver/pkg/utils"
)

// ZipDecompress Decompress.
type ZipDecompress struct {
	*Decompressor
}

// Name Name.
func (z *ZipDecompress) Name() string {
	return "zip"
}

// SetValue SetValue.
func (z *ZipDecompress) SetValue(d *Decompressor) {
	z.Decompressor = d
}

// Walk Walk.
func (z *ZipDecompress) Walk(file multipart.File) (string, error) {
	zipReader, err := z.zipReader(file)
	if err != nil {
		return "", err
	}

	indexHTML, indexHTM := z.genIndex("")
	indexPathArr := make([]string, 0, 4)
	indexPathArr = append(indexPathArr, indexHTML, indexHTM)
	flag := false
	for _, zipFile := range zipReader.File {
		// Prevent GBK garbled code
		fileName, err := decoding(zipFile.Name)
		if err != nil {
			return fileName, err
		}

		// Filter hidden files in MacOS
		if filterMacImpact(fileName) {
			continue
		}

		if zipFile.FileInfo().IsDir() {
			if !flag {
				flag = true
				indexHTML1, indexHTM2 := z.genIndex(fileName)
				indexPathArr = append(indexPathArr, indexHTML1, indexHTM2)
			}
			continue
		}

		if !flag {
			flag = true
		}

		if z.checkIndex(fileName, indexPathArr) {
			return fileName, nil
		}
	}

	return "", nil
}

// Unarchive Unarchive.
func (z *ZipDecompress) Unarchive(file multipart.File, dst string) error {
	// Create save directory
	err := os.MkdirAll(dst, fileMode)
	if err != nil {
		return err
	}

	// Open zip file
	zipReader, err := z.zipReader(file)
	if err != nil {
		return err
	}

	childDst := dst
	flag := false
	for _, zipFile := range zipReader.File {
		// Prevent GBK garbled code
		fileName, err := decoding(zipFile.Name)
		if err != nil {
			return err
		}

		fileName = filepath.Join(dst, fileName)

		// Filter hidden files in MacOS
		if filterMacImpact(fileName) {
			continue
		}

		rc, err := zipFile.Open()
		if err != nil {
			return err
		}
		defer rc.Close() // nolint: errcheck, gosec

		switch zipFile.FileInfo().IsDir() {
		case true:
			childDst, flag, err = z.mkdir(flag, fileName, childDst)
			if err != nil {
				return err
			}
		case false:
			flag, err = z.createFile(rc, flag, fileName)
			if err != nil {
				return err
			}
			// Deep decompression
			err = z.diffUnarchive(fileName, childDst)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (z *ZipDecompress) zipReader(file multipart.File) (*zip.Reader, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
}

func (z *ZipDecompress) mkdir(flag bool, fileName, childDst string) (string, bool, error) {
	err := os.MkdirAll(fileName, fileMode)
	if err != nil {
		return "", false, err
	}

	if !flag {
		flag = true
		childDst = fileName
	}

	return childDst, flag, nil
}

func (z *ZipDecompress) createFile(rc io.ReadCloser, flag bool, fileName string) (bool, error) {
	if !flag {
		flag = true
	}

	w, err := os.Create(fileName)
	if err != nil {
		return flag, err
	}
	defer w.Close() // nolint: errcheck, gosec

	_, err = io.Copy(w, rc)

	return flag, err
}

func (z *ZipDecompress) diffUnarchive(fileName, childDst string) error {
	ext := utils.GetExt(fileName)
	dp, err := z.GetDecompress(ext)
	if err != nil {
		return nil
	}

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close() // nolint: errcheck, gosec

	err = dp.Unarchive(f, childDst)
	if err != nil {
		return err
	}

	// 删除解压的文件
	return os.RemoveAll(fileName)
}

// genIndex genIndex.
func (z *ZipDecompress) genIndex(dir string) (string, string) {
	if dir == "" {
		return indexHTML, indexHTM
	}

	return fmt.Sprintf("%s%s", dir, indexHTML), fmt.Sprintf("%s%s", dir, indexHTM)
}

// checkIndex checkIndex.
func (z *ZipDecompress) checkIndex(fileName string, indexPathArr []string) bool {
	for _, indexPath := range indexPathArr {
		if fileName == indexPath {
			return true
		}
	}
	return false
}
