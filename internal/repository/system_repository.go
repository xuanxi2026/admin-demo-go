package repository

import (
	"admin-demo-go/internal/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SystemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

func (r *SystemRepository) ListDictItems(pageNo, pageSize int, dictType, keyword string) ([]model.DictItem, int64, error) {
	var (
		list  []model.DictItem
		total int64
	)

	query := r.db.Model(&model.DictItem{})
	if strings.TrimSpace(dictType) != "" {
		query = query.Where("dict_type = ?", strings.TrimSpace(dictType))
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		query = query.Where("dict_type LIKE ? OR label LIKE ? OR value LIKE ?", "%"+kw+"%", "%"+kw+"%", "%"+kw+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (pageNo - 1) * pageSize
	if err := query.Order("dict_type asc, sort asc, id asc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *SystemRepository) CreateDictItem(item *model.DictItem) error {
	return r.db.Create(item).Error
}

func (r *SystemRepository) UpdateDictItemByID(id uint, updates map[string]any) error {
	return r.db.Model(&model.DictItem{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SystemRepository) DeleteDictItemsByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&model.DictItem{}).Error
}

func (r *SystemRepository) ListConfigs(pageNo, pageSize int, group, keyword string) ([]model.SystemConfig, int64, error) {
	var (
		list  []model.SystemConfig
		total int64
	)

	query := r.db.Model(&model.SystemConfig{})
	query = query.Where("`group` <> ?", "storage")
	if strings.TrimSpace(group) != "" {
		query = query.Where("`group` = ?", strings.TrimSpace(group))
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		query = query.Where("config_key LIKE ? OR name LIKE ? OR `group` LIKE ?", "%"+kw+"%", "%"+kw+"%", "%"+kw+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (pageNo - 1) * pageSize
	if err := query.Order("`group` asc, id asc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *SystemRepository) CreateConfig(item *model.SystemConfig) error {
	return r.db.Create(item).Error
}

func (r *SystemRepository) FindConfigByKey(configKey string) (*model.SystemConfig, error) {
	var item model.SystemConfig
	if err := r.db.Where("config_key = ?", strings.TrimSpace(configKey)).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SystemRepository) UpdateConfigByID(id uint, updates map[string]any) error {
	return r.db.Model(&model.SystemConfig{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SystemRepository) DeleteConfigsByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&model.SystemConfig{}).Error
}

func (r *SystemRepository) ValidateConfigDeletion(ids []uint) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}
	var list []model.SystemConfig
	if err := r.db.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return false, err
	}
	for _, item := range list {
		if strings.EqualFold(item.Group, "storage") || strings.HasPrefix(strings.ToLower(strings.TrimSpace(item.ConfigKey)), "storage.") {
			return true, nil
		}
	}
	return false, nil
}

func (r *SystemRepository) CreateOperationLog(item *model.OperationLog) error {
	return r.db.Create(item).Error
}

func (r *SystemRepository) ListOperationLogs(pageNo, pageSize int, module, keyword string) ([]model.OperationLog, int64, error) {
	var (
		list  []model.OperationLog
		total int64
	)

	query := r.db.Model(&model.OperationLog{})
	if strings.TrimSpace(module) != "" {
		query = query.Where("module = ?", strings.TrimSpace(module))
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		query = query.Where("operator LIKE ? OR target LIKE ? OR detail LIKE ? OR action LIKE ?", "%"+kw+"%", "%"+kw+"%", "%"+kw+"%", "%"+kw+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (pageNo - 1) * pageSize
	if err := query.Order("created_at desc, id desc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *SystemRepository) SeedOperationLog(module, action, operator, target, requestID, ip, detail string) error {
	return r.CreateOperationLog(&model.OperationLog{
		Module:    module,
		Action:    action,
		Operator:  operator,
		Target:    target,
		RequestID: requestID,
		IP:        ip,
		Detail:    detail,
		CreatedAt: time.Now(),
	})
}

func (r *SystemRepository) ListDepartments() ([]model.Department, error) {
	var list []model.Department
	err := r.db.Order("sort asc, id asc").Find(&list).Error
	return list, err
}

func (r *SystemRepository) CreateDepartment(item *model.Department) error {
	return r.db.Create(item).Error
}

func (r *SystemRepository) UpdateDepartmentByID(id uint, updates map[string]any) error {
	return r.db.Model(&model.Department{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SystemRepository) DeleteDepartmentsByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&model.Department{}).Error
}

func (r *SystemRepository) ListNotices(pageNo, pageSize int, level, status, keyword string) ([]model.Notice, int64, error) {
	var (
		list  []model.Notice
		total int64
	)
	query := r.db.Model(&model.Notice{})
	if strings.TrimSpace(level) != "" {
		query = query.Where("level = ?", strings.TrimSpace(level))
	}
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		query = query.Where("title LIKE ? OR content LIKE ? OR publisher LIKE ?", "%"+kw+"%", "%"+kw+"%", "%"+kw+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (pageNo - 1) * pageSize
	if err := query.Order("sort asc, id desc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *SystemRepository) CreateNotice(item *model.Notice) error {
	return r.db.Create(item).Error
}

func (r *SystemRepository) UpdateNoticeByID(id uint, updates map[string]any) error {
	return r.db.Model(&model.Notice{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SystemRepository) DeleteNoticesByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&model.Notice{}).Error
}

func (r *SystemRepository) ListPublishedNoticesWithReadStatus(userID uint, pageNo, pageSize int) ([]model.Notice, map[uint]time.Time, int64, error) {
	var (
		list  []model.Notice
		total int64
	)
	query := r.db.Model(&model.Notice{}).Where("status = ?", "published")
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}
	offset := (pageNo - 1) * pageSize
	if err := query.Order("sort asc, id desc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, nil, 0, err
	}
	readMap := map[uint]time.Time{}
	if len(list) == 0 {
		return list, readMap, total, nil
	}
	noticeIDs := make([]uint, 0, len(list))
	for _, item := range list {
		noticeIDs = append(noticeIDs, item.ID)
	}
	var reads []model.NoticeRead
	if err := r.db.Where("user_id = ? AND notice_id IN ?", userID, noticeIDs).Find(&reads).Error; err != nil {
		return nil, nil, 0, err
	}
	for _, item := range reads {
		readMap[item.NoticeID] = item.ReadAt
	}
	return list, readMap, total, nil
}

func (r *SystemRepository) MarkNoticeRead(userID, noticeID uint) error {
	item := model.NoticeRead{
		UserID:   userID,
		NoticeID: noticeID,
		ReadAt:   time.Now(),
	}
	return r.db.Where("user_id = ? AND notice_id = ?", userID, noticeID).
		Assign(model.NoticeRead{ReadAt: item.ReadAt}).
		FirstOrCreate(&item).Error
}

func (r *SystemRepository) MarkNoticeUnread(userID, noticeID uint) error {
	return r.db.Where("user_id = ? AND notice_id = ?", userID, noticeID).Delete(&model.NoticeRead{}).Error
}
