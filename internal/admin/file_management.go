package admin

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"
	"admin-demo-go/internal/storage"

	"github.com/gin-gonic/gin"
)

type fileDeleteRequest struct {
	Name string `json:"name"`
}

func (m *Module) FilePublicList(c *gin.Context) {
	rows, err := m.storage.List(context.Background(), storage.AreaPublic)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询公开文件失败")
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
	response.OK(c, "success", data)
}

func (m *Module) FilePrivateList(c *gin.Context) {
	rows, err := m.storage.List(context.Background(), storage.AreaPrivate)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询私有文件失败")
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
	response.OK(c, "success", data)
}

func (m *Module) FilePublicUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, ecode.InvalidParams, "请选择上传文件")
		return
	}
	defer file.Close()
	info, err := m.storage.Upload(context.Background(), storage.AreaPublic, header.Filename, file, header.Size)
	if err != nil {
		response.Fail(c, ecode.InvalidParams, err.Error())
		return
	}
	m.pub.Publish("file_public_upload", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  info.Name,
	})
	m.recordOperation(operationContext{
		Module:    "file",
		Action:    "upload_public",
		Operator:  c.GetString("username"),
		Target:    info.Name,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    buildDetail("area=public", fmt.Sprintf("size=%d", info.Size)),
	})
	row := gin.H{
		"name":        info.Name,
		"size":        info.Size,
		"updatedAt":   info.UpdatedAt.Format("2006-01-02 15:04:05"),
		"downloadUrl": m.publicDownloadURL(info.Name),
	}
	response.OK(c, "上传成功", row)
}

func (m *Module) FilePrivateUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, ecode.InvalidParams, "请选择上传文件")
		return
	}
	defer file.Close()
	info, err := m.storage.Upload(context.Background(), storage.AreaPrivate, header.Filename, file, header.Size)
	if err != nil {
		response.Fail(c, ecode.InvalidParams, err.Error())
		return
	}
	m.pub.Publish("file_private_upload", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  info.Name,
	})
	m.recordOperation(operationContext{
		Module:    "file",
		Action:    "upload_private",
		Operator:  c.GetString("username"),
		Target:    info.Name,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    buildDetail("area=private", fmt.Sprintf("size=%d", info.Size)),
	})
	row := gin.H{
		"name":      info.Name,
		"size":      info.Size,
		"updatedAt": info.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	response.OK(c, "上传成功", row)
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
		response.Fail(c, ecode.InvalidParams, "文件名不合法")
		return
	}
	reader, err := m.storage.Download(context.Background(), storage.AreaPrivate, name)
	if err != nil {
		response.Fail(c, ecode.InternalError, "文件不存在")
		return
	}
	defer reader.Close()
	m.pub.Publish("file_private_download", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  name,
	})
	m.recordOperation(operationContext{
		Module:    "file",
		Action:    "download_private",
		Operator:  c.GetString("username"),
		Target:    name,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "下载私有文件",
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
		response.Fail(c, ecode.InvalidParams, "参数不合法")
		return
	}
	name := sanitizeFileName(req.Name)
	if name == "" {
		response.Fail(c, ecode.InvalidParams, "文件名不合法")
		return
	}
	if err := m.storage.Delete(context.Background(), storage.AreaPublic, name); err != nil {
		response.Fail(c, ecode.InternalError, "删除公开文件失败")
		return
	}
	m.pub.Publish("file_public_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  name,
	})
	m.recordOperation(operationContext{
		Module:    "file",
		Action:    "delete_public",
		Operator:  c.GetString("username"),
		Target:    name,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "删除公开文件",
	})
	response.OK(c, "删除成功", nil)
}

func (m *Module) FilePrivateDelete(c *gin.Context) {
	var req fileDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数不合法")
		return
	}
	name := sanitizeFileName(req.Name)
	if name == "" {
		response.Fail(c, ecode.InvalidParams, "文件名不合法")
		return
	}
	if err := m.storage.Delete(context.Background(), storage.AreaPrivate, name); err != nil {
		response.Fail(c, ecode.InternalError, "删除私有文件失败")
		return
	}
	m.pub.Publish("file_private_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"file_name":  name,
	})
	m.recordOperation(operationContext{
		Module:    "file",
		Action:    "delete_private",
		Operator:  c.GetString("username"),
		Target:    name,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "删除私有文件",
	})
	response.OK(c, "删除成功", nil)
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
