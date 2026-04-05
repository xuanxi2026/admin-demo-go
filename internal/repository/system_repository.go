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

func (r *SystemRepository) UpdateConfigByID(id uint, updates map[string]any) error {
	return r.db.Model(&model.SystemConfig{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SystemRepository) DeleteConfigsByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&model.SystemConfig{}).Error
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
