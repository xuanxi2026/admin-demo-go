package admin

import (
	"context"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/storage"

	"github.com/gin-gonic/gin"
)

type fileDeleteRequest struct {
	Name string `json:"name"`
}

func (m *Module) FilePublicList(c *gin.Context) {
	rows, err := m.storage.List(context.Background(), storage.AreaPublic)
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "查询公开文件失败", "request_id": c.GetString("request_id")})
		return
	}
	data := make([]gin.H, 0, len(rows))
	for _, item := range rows {
		data = append(data, gin.H{
			"name":        item.Name,
			"size":        item.Size,
			"updatedAt":   item.UpdatedAt.Format("2006-01-02 15:04:05"),
			"downloadUrl": m.publicDownloadURL(item.Name),
		})
	}
	c.JSON(200, gin.H{"code": 200, "msg": "success", "data": data, "request_id": c.GetString("request_id")})
}

func (m *Module) FilePrivateList(c *gin.Context) {
	rows, err := m.storage.List(context.Background(), storage.AreaPrivate)
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "查询私有文件失败", "request_id": c.GetString("request_id")})
		return
	}
	data := make([]gin.H, 0, len(rows))
	for _, item := range rows {
		data = append(data, gin.H{
			"name":      item.Name,
			"size":      item.Size,
			"updatedAt": item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(200, gin.H{"code": 200, "msg": "success", "data": data, "request_id": c.GetString("request_id")})
}

func (m *Module) FilePublicUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "请选择上传文件", "request_id": c.GetString("request_id")})
		return
	}
	defer file.Close()
	info, err := m.storage.Upload(context.Background(), storage.AreaPublic, header.Filename, file, header.Size)
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": err.Error(), "request_id": c.GetString("request_id")})
		return
	}
	m.pub.Publish("file_public_upload", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  info.Name,
	})
	row := gin.H{
		"name":        info.Name,
		"size":        info.Size,
		"updatedAt":   info.UpdatedAt.Format("2006-01-02 15:04:05"),
		"downloadUrl": m.publicDownloadURL(info.Name),
	}
	c.JSON(200, gin.H{"code": 200, "msg": "上传成功", "data": row, "request_id": c.GetString("request_id")})
}

func (m *Module) FilePrivateUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "请选择上传文件", "request_id": c.GetString("request_id")})
		return
	}
	defer file.Close()
	info, err := m.storage.Upload(context.Background(), storage.AreaPrivate, header.Filename, file, header.Size)
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": err.Error(), "request_id": c.GetString("request_id")})
		return
	}
	m.pub.Publish("file_private_upload", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  info.Name,
	})
	row := gin.H{
		"name":      info.Name,
		"size":      info.Size,
		"updatedAt": info.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	c.JSON(200, gin.H{"code": 200, "msg": "上传成功", "data": row, "request_id": c.GetString("request_id")})
}

func (m *Module) FilePublicDownload(c *gin.Context) {
	name := sanitizeFileName(c.Param("name"))
	if name == "" {
		c.String(404, "file not found")
		return
	}
	reader, err := m.storage.Download(context.Background(), storage.AreaPublic, name)
	if err != nil {
		c.String(404, "file not found")
		return
	}
	defer reader.Close()
	c.Header("Content-Disposition", "attachment; filename=\""+name+"\"")
	c.Stream(func(w io.Writer) bool {
		_, err = io.Copy(w, reader)
		return false
	})
}

func (m *Module) FilePrivateDownload(c *gin.Context) {
	raw := c.Param("name")
	decoded, _ := url.PathUnescape(raw)
	name := sanitizeFileName(decoded)
	if name == "" {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "文件名不合法", "request_id": c.GetString("request_id")})
		return
	}
	reader, err := m.storage.Download(context.Background(), storage.AreaPrivate, name)
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "文件不存在", "request_id": c.GetString("request_id")})
		return
	}
	defer reader.Close()
	m.pub.Publish("file_private_download", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  name,
	})
	c.Header("Content-Disposition", "attachment; filename=\""+name+"\"")
	c.Stream(func(w io.Writer) bool {
		_, err = io.Copy(w, reader)
		return false
	})
}

func (m *Module) FilePublicDelete(c *gin.Context) {
	var req fileDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "参数不合法", "request_id": c.GetString("request_id")})
		return
	}
	name := sanitizeFileName(req.Name)
	if name == "" {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "文件名不合法", "request_id": c.GetString("request_id")})
		return
	}
	if err := m.storage.Delete(context.Background(), storage.AreaPublic, name); err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "删除公开文件失败", "request_id": c.GetString("request_id")})
		return
	}
	m.pub.Publish("file_public_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  name,
	})
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功", "request_id": c.GetString("request_id")})
}

func (m *Module) FilePrivateDelete(c *gin.Context) {
	var req fileDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "参数不合法", "request_id": c.GetString("request_id")})
		return
	}
	name := sanitizeFileName(req.Name)
	if name == "" {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "文件名不合法", "request_id": c.GetString("request_id")})
		return
	}
	if err := m.storage.Delete(context.Background(), storage.AreaPrivate, name); err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "删除私有文件失败", "request_id": c.GetString("request_id")})
		return
	}
	m.pub.Publish("file_private_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  name,
	})
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功", "request_id": c.GetString("request_id")})
}

func sanitizeFileName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "..", "")
	return name
}

func (m *Module) publicDownloadURL(name string) string {
	path := "/files/public/" + url.PathEscape(name)
	if strings.TrimSpace(m.publicBaseURL) == "" {
		return path
	}
	base := strings.TrimRight(strings.TrimSpace(m.publicBaseURL), "/")
	return base + path
}
