package mysql

import (
	"github.com/quanxiang-cloud/fileserver/internal/models"

	"gorm.io/gorm"
)

type fileserver struct{}

// NewFileServerRepo new FileServerRepo
func NewFileServerRepo() models.FileServerRepo {
	return &fileserver{}
}

func (f *fileserver) TableName() string {
	return "fileserver"
}

func (f *fileserver) GetByPath(db *gorm.DB, path string) (*models.FileServer, error) {
	fileInfo := new(models.FileServer)

	err := db.Table(f.TableName()).
		Where("path = ?", path).
		Find(&fileInfo).
		Error
	if err != nil {
		return nil, err
	}

	if fileInfo.ID == "" {
		return nil, nil
	}

	return fileInfo, nil
}

func (f *fileserver) Create(db *gorm.DB, fileserver *models.FileServer) error {
	return db.Table(f.TableName()).
		Create(fileserver).
		Error
}

func (f *fileserver) Delete(db *gorm.DB, id string) error {
	return db.Table(f.TableName()).
		Where("id = ?", id).
		Delete(&models.FileServer{}).
		Error
}

func (f *fileserver) UpdateNumber(db *gorm.DB, id string, number int) error {
	return db.Table(f.TableName()).
		Where("id = ?", id).
		Update("number", gorm.Expr("number+ ?", 1)).
		Error
}
