package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalClient struct {
	baseDir string
}

func NewLocal(baseDir string) *LocalClient {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "storage"
	}
	return &LocalClient{baseDir: baseDir}
}

func (c *LocalClient) List(_ context.Context, area string) ([]ObjectInfo, error) {
	dir := c.areaDir(area)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	rows := make([]ObjectInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		rows = append(rows, ObjectInfo{
			Name:      e.Name(),
			Size:      info.Size(),
			UpdatedAt: info.ModTime(),
		})
	}
	return rows, nil
}

func (c *LocalClient) Upload(_ context.Context, area, filename string, reader io.Reader, _ int64) (ObjectInfo, error) {
	dir := c.areaDir(area)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ObjectInfo{}, err
	}
	name := safeName(filename)
	if name == "" {
		return ObjectInfo{}, fmt.Errorf("invalid filename")
	}
	target := filepath.Join(dir, name)
	if _, err := os.Stat(target); err == nil {
		name = fmt.Sprintf("%d_%s", time.Now().UnixMilli(), name)
		target = filepath.Join(dir, name)
	}
	f, err := os.Create(target)
	if err != nil {
		return ObjectInfo{}, err
	}
	defer f.Close()
	n, err := io.Copy(f, reader)
	if err != nil {
		return ObjectInfo{}, err
	}
	stat, _ := os.Stat(target)
	tm := time.Now()
	if stat != nil {
		tm = stat.ModTime()
	}
	return ObjectInfo{Name: name, Size: n, UpdatedAt: tm}, nil
}

func (c *LocalClient) Download(_ context.Context, area, filename string) (io.ReadCloser, error) {
	name := safeName(filename)
	if name == "" {
		return nil, os.ErrNotExist
	}
	target := filepath.Join(c.areaDir(area), name)
	return os.Open(target)
}

func (c *LocalClient) Delete(_ context.Context, area, filename string) error {
	name := safeName(filename)
	if name == "" {
		return os.ErrNotExist
	}
	target := filepath.Join(c.areaDir(area), name)
	return os.Remove(target)
}

func (c *LocalClient) areaDir(area string) string {
	if area == AreaPrivate {
		return filepath.Join(c.baseDir, AreaPrivate)
	}
	return filepath.Join(c.baseDir, AreaPublic)
}

func safeName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "..", "")
	return name
}
