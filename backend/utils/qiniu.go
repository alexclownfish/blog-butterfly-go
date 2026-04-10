package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"gopkg.in/ini.v1"
	"mime/multipart"
	"path/filepath"
	"sync"
	"time"
)

var (
	AccessKey   string
	SecretKey   string
	Bucket      string
	QiniuServer string

	qiniuConfigOnce sync.Once
	qiniuConfigErr  error
)

func loadQiniuConfig() error {
	qiniuConfigOnce.Do(func() {
		cfg, err := ini.Load("config.ini")
		if err != nil {
			qiniuConfigErr = fmt.Errorf("配置文件加载失败: %w", err)
			return
		}

		AccessKey = cfg.Section("qiniu").Key("AccessKey").String()
		SecretKey = cfg.Section("qiniu").Key("SecretKey").String()
		Bucket = cfg.Section("qiniu").Key("Bucket").String()
		QiniuServer = cfg.Section("qiniu").Key("QiniuServer").String()

		if AccessKey == "" || SecretKey == "" || Bucket == "" || QiniuServer == "" {
			qiniuConfigErr = errors.New("七牛配置不完整")
		}
	})

	return qiniuConfigErr
}

func UploadToQiniu(file multipart.File, filename string) (string, error) {
	if err := loadQiniuConfig(); err != nil {
		return "", err
	}

	mac := qbox.NewMac(AccessKey, SecretKey)
	putPolicy := storage.PutPolicy{Scope: Bucket}
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{UseHTTPS: false}
	formUploader := storage.NewFormUploader(&cfg)

	key := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(filename))
	ret := storage.PutRet{}

	err := formUploader.Put(context.Background(), &ret, upToken, key, file, -1, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", QiniuServer, key), nil
}

func ListQiniuImages() ([]map[string]interface{}, error) {
	if err := loadQiniuConfig(); err != nil {
		return nil, err
	}

	mac := qbox.NewMac(AccessKey, SecretKey)
	cfg := storage.Config{UseHTTPS: false}
	bucketManager := storage.NewBucketManager(mac, &cfg)

	limit := 100
	prefix := ""
	delimiter := ""
	marker := ""

	entries, _, _, _, err := bucketManager.ListFiles(Bucket, prefix, delimiter, marker, limit)
	if err != nil {
		return nil, err
	}

	var images []map[string]interface{}
	for _, entry := range entries {
		images = append(images, map[string]interface{}{
			"url":  fmt.Sprintf("%s%s", QiniuServer, entry.Key),
			"key":  entry.Key,
			"size": entry.Fsize,
			"time": entry.PutTime,
		})
	}
	return images, nil
}

func DeleteQiniuImage(key string) error {
	if err := loadQiniuConfig(); err != nil {
		return err
	}

	mac := qbox.NewMac(AccessKey, SecretKey)
	cfg := storage.Config{UseHTTPS: false}
	bucketManager := storage.NewBucketManager(mac, &cfg)
	return bucketManager.Delete(Bucket, key)
}
