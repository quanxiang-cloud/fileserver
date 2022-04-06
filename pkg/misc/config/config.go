package config

import (
	"os"
	"time"

	"github.com/quanxiang-cloud/cabin/logger"
	mysql2 "github.com/quanxiang-cloud/cabin/tailormade/db/mysql"
	redis2 "github.com/quanxiang-cloud/cabin/tailormade/db/redis"

	"gopkg.in/yaml.v2"
)

// Conf configuration file.
var Conf *Config

// DefaultPath default configuration path.
var DefaultPath = "./configs/config.yml"

// Config configuration file.
type Config struct {
	Port    string            `yaml:"port"`
	Model   string            `yaml:"model"`
	MaxSize int64             `yaml:"maxSize"`
	Log     logger.Config     `yaml:"log"`
	Mysql   mysql2.Config     `yaml:"mysql"`
	Redis   redis2.Config     `yaml:"redis"`
	Storage Storage           `yaml:"storage"`
	Blob    Blob              `yaml:"blob"`
	Buckets map[string]string `yaml:"buckets"`
}

// Storage Storage.
type Storage struct {
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	Region          string
	URLExpire       time.Duration
	PartExpire      time.Duration
}

// Blob decompression configuration.
type Blob struct {
	Template string `yaml:"template"`
	TempPath string `yaml:"tempPath"`
}

// NewConfig get configuration.
func NewConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &Conf)
	if err != nil {
		return nil, err
	}

	return Conf, nil
}
