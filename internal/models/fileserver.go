package models

import (
	"gorm.io/gorm"
)

// update number
const (
	ReduceNumber    = -1
	IncrementNumber = 1
)

// FileServer corresponding structure of fileserver file service
type FileServer struct {
	ID       string `gorm:"column:id"`
	Path     string `gorm:"column:path"`
	CreateAt int64  `gorm:"column:create_at"`
	UpdateAt int64  `gorm:"column:update_at"`
}

// FileServerRepo file service logical interface
type FileServerRepo interface {
	GetByPath(db *gorm.DB, path string) (*FileServer, error)
	Create(db *gorm.DB, fileserver *FileServer) error
	Delete(db *gorm.DB, id string) error
}
