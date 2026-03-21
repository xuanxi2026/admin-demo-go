package storage

import (
	"context"
	"io"
	"time"
)

const (
	AreaPublic  = "public"
	AreaPrivate = "private"
)

type ObjectInfo struct {
	Name      string
	Size      int64
	UpdatedAt time.Time
}

type Client interface {
	List(ctx context.Context, area string) ([]ObjectInfo, error)
	Upload(ctx context.Context, area, filename string, reader io.Reader, size int64) (ObjectInfo, error)
	Download(ctx context.Context, area, filename string) (io.ReadCloser, error)
	Delete(ctx context.Context, area, filename string) error
}
