package repository

import (
	"admin-demo-go/internal/model"
	"strings"

	"gorm.io/gorm"
)

type RBACRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) *RBACRepository {
	return &RBACRepository{db: db}
}

func (r *RBACRepository) GetPermissionsByUserID(userID uint) ([]string, error) {
	var codes []string
	err := r.db.Table("permissions p").
		Select("distinct p.code").
		Joins("join role_permissions rp on rp.permission_id = p.id").
		Joins("join user_roles ur on ur.role_id = rp.role_id").
		Where("ur.user_id = ?", userID).
		Order("p.code asc").
		Scan(&codes).Error
	return codes, err
}

func (r *RBACRepository) UserHasPermission(userID uint, code string) (bool, error) {
	var count int64
	err := r.db.Table("permissions p").
		Joins("join role_permissions rp on rp.permission_id = p.id").
		Joins("join user_roles ur on ur.role_id = rp.role_id").
		Where("ur.user_id = ? and p.code = ?", userID, code).
		Count(&count).Error
	return count > 0, err
}

func (r *RBACRepository) GetMenusByUserID(userID uint) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Table("menus m").
		Select("distinct m.*").
		Joins("join role_menus rm on rm.menu_id = m.id").
		Joins("join user_roles ur on ur.role_id = rm.role_id").
		Where("ur.user_id = ? and m.hidden = ?", userID, false).
		Order("m.parent_id asc, m.sort asc, m.id asc").
		Find(&menus).Error
	return menus, err
}

func (r *RBACRepository) GetRoleCodesByUserID(userID uint) ([]string, error) {
	var roleCodes []string
	err := r.db.Table("roles r").
		Select("r.code").
		Joins("join user_roles ur on ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Scan(&roleCodes).Error
	return roleCodes, err
}

func (r *RBACRepository) BindUserRole(userID, roleID uint) error {
	return r.db.Create(&model.UserRole{UserID: userID, RoleID: roleID}).Error
}

func (r *RBACRepository) FindRoleByCode(code string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RBACRepository) ReplaceUserRoles(userID uint, roleCodes []string) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return err
	}
	for _, code := range roleCodes {
		role, err := r.FindRoleByCode(code)
		if err != nil {
			continue
		}
		if err = r.db.Create(&model.UserRole{UserID: userID, RoleID: role.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *RBACRepository) ListRoles(pageNo, pageSize int, permission string) ([]model.Role, int64, error) {
	var (
		list  []model.Role
		total int64
	)
	query := r.db.Model(&model.Role{})
	if strings.TrimSpace(permission) != "" {
		query = query.Where("code LIKE ?", "%"+strings.TrimSpace(permission)+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (pageNo - 1) * pageSize
	if err := query.Order("id asc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *RBACRepository) CreateRole(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *RBACRepository) UpdateRoleByID(id uint, updates map[string]any) error {
	return r.db.Model(&model.Role{}).Where("id = ?", id).Updates(updates).Error
}

func (r *RBACRepository) DeleteRolesByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.Where("id IN ?", ids).Delete(&model.Role{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("role_id IN ?", ids).Delete(&model.RolePermission{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("role_id IN ?", ids).Delete(&model.RoleMenu{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("role_id IN ?", ids).Delete(&model.UserRole{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *RBACRepository) ListAllRoles() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Order("id asc").Find(&roles).Error
	return roles, err
}

func (r *RBACRepository) CreateMenu(menu *model.Menu) error {
	return r.db.Create(menu).Error
}

func (r *RBACRepository) CreateMenuMap(menu map[string]any) error {
	return r.db.Model(&model.Menu{}).Create(menu).Error
}

func (r *RBACRepository) UpdateMenuByID(id uint, updates map[string]any) error {
	return r.db.Model(&model.Menu{}).Where("id = ?", id).Updates(updates).Error
}

func (r *RBACRepository) DeleteMenusByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.Where("id IN ?", ids).Delete(&model.Menu{}).Error; err != nil {
		return err
	}
	return r.db.Where("menu_id IN ?", ids).Delete(&model.RoleMenu{}).Error
}
