package code

import (
	error2 "github.com/quanxiang-cloud/cabin/error"
)

func init() {
	error2.CodeTable = CodeTable
}

// error code.
const (
	InvalidStorage       = 100015000001
	ErrFileLimit         = 100014020002
	InvalidDelFile       = 100014020003
	ErrUploadFile        = 100014020004
	ErrDownload          = 100014020005
	InvalidIndex         = 100014020006
	InvalidCompress      = 100014020007
	InvalidExist         = 100014020008
	ErrThumbnail         = 100014020009
	ErrUnarchive         = 100014020010
	ErrSinger            = 100014020011
	ErrListMultiPart     = 100014020012
	ErrCompleteMultiPart = 100014020013
)

// CodeTable code table.
var CodeTable = map[int64]string{
	InvalidStorage:       "无效的存储配置",
	ErrFileLimit:         "文件大小超出限制",
	InvalidDelFile:       "文件已被删除",
	ErrUploadFile:        "文件上传失败",
	ErrDownload:          "文件下载失败",
	InvalidIndex:         "根目录下没有index页面",
	InvalidCompress:      "请上传合法的压缩包",
	InvalidExist:         "文件不存在",
	ErrThumbnail:         "图片裁剪失败",
	ErrUnarchive:         "解压失败",
	ErrSinger:            "签名失败",
	ErrListMultiPart:     "查找分块失败",
	ErrCompleteMultiPart: "合并分块失败",
}
