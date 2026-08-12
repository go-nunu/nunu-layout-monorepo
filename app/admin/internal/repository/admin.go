package repository

import (
	"context"
	"fmt"
	v1 "nunu-layout-monorepo/app/admin/api/v1"
	"nunu-layout-monorepo/model"

	"github.com/duke-git/lancet/v2/convertor"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminRepository interface {
	GetAdminUsers(ctx context.Context, req *v1.GetAdminUsersRequest) ([]model.AdminUser, int64, error)
	GetAdminUser(ctx context.Context, uid uint) (model.AdminUser, error)
	GetAdminUserByUsername(ctx context.Context, username string) (model.AdminUser, error)
	AdminUserUpdate(ctx context.Context, m *model.AdminUser) error
	AdminUserCreate(ctx context.Context, m *model.AdminUser) error
	AdminUserDelete(ctx context.Context, id uint) error

	GetUserPermissions(ctx context.Context, uid uint) ([][]string, error)
	GetUserRoles(ctx context.Context, uid uint) ([]string, error)
	GetRolePermissions(ctx context.Context, role string) ([][]string, error)
	UpdateRolePermission(ctx context.Context, role string, permissions []model.Permission) error
	UpdateUserRoles(ctx context.Context, uid uint, roles []string) error
	DeleteUserRoles(ctx context.Context, uid uint) error
	ReplacePermissionReferences(ctx context.Context, replacements map[model.Permission]model.Permission) error
	DeletePermissionReferences(ctx context.Context, permissions []model.Permission) error
	ReloadPolicy() error

	GetMenuList(ctx context.Context) ([]model.Menu, error)
	MenuUpdate(ctx context.Context, m *model.Menu) error
	MenuCreate(ctx context.Context, m *model.Menu) error
	MenuDelete(ctx context.Context, id uint) error
	RemoveMenuFromApis(ctx context.Context, menuID uint) error

	GetRoles(ctx context.Context, req *v1.GetRoleListRequest) ([]model.Role, int64, error)
	RoleUpdate(ctx context.Context, m *model.Role) error
	RoleCreate(ctx context.Context, m *model.Role) error
	RoleDelete(ctx context.Context, id uint) error
	CasbinRoleDelete(ctx context.Context, role string) error
	GetRole(ctx context.Context, id uint) (model.Role, error)
	GetRoleBySid(ctx context.Context, sid string) (model.Role, error)

	GetApis(ctx context.Context, req *v1.GetApisRequest) ([]model.Api, int64, error)
	GetApiGroups(ctx context.Context) ([]string, error)
	GetApi(ctx context.Context, id uint) (model.Api, error)
	GetApiList(ctx context.Context) ([]model.Api, error)
	ApiPermissionExists(ctx context.Context, path, method string, excludeID uint) (bool, error)
	ApiUpdate(ctx context.Context, m *model.Api) error
	ApiCreate(ctx context.Context, m *model.Api) error
	ApiDelete(ctx context.Context, id uint) error
}

func NewAdminRepository(
	repository *Repository,
) AdminRepository {
	return &adminRepository{
		Repository: repository,
	}
}

type adminRepository struct {
	*Repository
}

type casbinRule struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"size:100"`
	V0    string `gorm:"size:100"`
	V1    string `gorm:"size:100"`
	V2    string `gorm:"size:100"`
	V3    string `gorm:"size:100"`
	V4    string `gorm:"size:100"`
	V5    string `gorm:"size:100"`
}

func (casbinRule) TableName() string {
	return "casbin_rule"
}

func (r *adminRepository) CasbinRoleDelete(ctx context.Context, role string) error {
	_, err := r.e.DeleteRole(role)
	return err
}

func (r *adminRepository) GetRole(ctx context.Context, id uint) (model.Role, error) {
	m := model.Role{}
	return m, r.DB(ctx).Where("id = ?", id).First(&m).Error
}
func (r *adminRepository) GetRoleBySid(ctx context.Context, sid string) (model.Role, error) {
	m := model.Role{}
	return m, r.DB(ctx).Where("sid = ?", sid).First(&m).Error
}

func (r *adminRepository) DeleteUserRoles(ctx context.Context, uid uint) error {
	_, err := r.e.DeleteRolesForUser(convertor.ToString(uid))
	return err
}
func (r *adminRepository) UpdateUserRoles(ctx context.Context, uid uint, roles []string) error {
	if len(roles) == 0 {
		_, err := r.e.DeleteRolesForUser(convertor.ToString(uid))
		return err
	}
	old, err := r.e.GetRolesForUser(convertor.ToString(uid))
	if err != nil {
		return err
	}
	oldMap := make(map[string]struct{})
	newMap := make(map[string]struct{})
	for _, v := range old {
		oldMap[v] = struct{}{}
	}
	for _, v := range roles {
		newMap[v] = struct{}{}
	}
	addRoles := make([]string, 0)
	delRoles := make([]string, 0)

	for key, _ := range oldMap {
		if _, exists := newMap[key]; !exists {
			delRoles = append(delRoles, key)
		}
	}
	for key, _ := range newMap {
		if _, exists := oldMap[key]; !exists {
			addRoles = append(addRoles, key)
		}
	}
	if len(addRoles) == 0 && len(delRoles) == 0 {
		return nil
	}
	for _, role := range delRoles {
		if _, err := r.e.DeleteRoleForUser(convertor.ToString(uid), role); err != nil {
			r.logger.WithContext(ctx).Error("DeleteRoleForUser error", zap.Error(err))
			return err
		}
	}

	if len(addRoles) > 0 {
		_, err = r.e.AddRolesForUser(convertor.ToString(uid), addRoles)
		return err
	}
	return nil
}

func (r *adminRepository) GetAdminUserByUsername(ctx context.Context, username string) (model.AdminUser, error) {
	m := model.AdminUser{}
	return m, r.DB(ctx).Where("username = ?", username).First(&m).Error
}

func (r *adminRepository) GetAdminUsers(ctx context.Context, req *v1.GetAdminUsersRequest) ([]model.AdminUser, int64, error) {
	var list []model.AdminUser
	var total int64
	scope := r.DB(ctx).Model(&model.AdminUser{})
	if req.ID != 0 {
		scope = scope.Where("id = ?", req.ID)
	}
	if req.Username != "" {
		scope = scope.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Nickname != "" {
		scope = scope.Where("nickname LIKE ?", "%"+req.Nickname+"%")
	}
	if req.Email != "" {
		scope = scope.Where("email LIKE ?", "%"+req.Email+"%")
	}
	if req.Phone != "" {
		scope = scope.Where("phone LIKE ?", "%"+req.Phone+"%")
	}
	if err := scope.Count(&total).Error; err != nil {
		return nil, total, err
	}
	if err := scope.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, total, err
	}
	return list, total, nil
}

func (r *adminRepository) GetAdminUser(ctx context.Context, uid uint) (model.AdminUser, error) {
	m := model.AdminUser{}
	return m, r.DB(ctx).Where("id = ?", uid).First(&m).Error
}

func (r *adminRepository) AdminUserUpdate(ctx context.Context, m *model.AdminUser) error {
	tx := r.DB(ctx).Model(&model.AdminUser{}).Where("id = ?", m.ID).Select(
		"username",
		"nickname",
		"password",
		"email",
		"phone",
	).Updates(m)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return r.ensureRecordExists(ctx, &model.AdminUser{}, m.ID)
	}
	return nil
}

func (r *adminRepository) AdminUserCreate(ctx context.Context, m *model.AdminUser) error {
	return r.DB(ctx).Create(m).Error
}

func (r *adminRepository) AdminUserDelete(ctx context.Context, id uint) error {
	tx := r.DB(ctx).Where("id = ?", id).Delete(&model.AdminUser{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *adminRepository) UpdateRolePermission(ctx context.Context, role string, permissions []model.Permission) error {
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ptype = ? AND v0 = ?", "p", role).Delete(&casbinRule{}).Error; err != nil {
			return err
		}
		if len(permissions) == 0 {
			return nil
		}
		rules := make([]casbinRule, 0, len(permissions))
		for _, permission := range permissions {
			rules = append(rules, casbinRule{
				Ptype: "p",
				V0:    role,
				V1:    permission.Resource,
				V2:    permission.Action,
			})
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rules).Error
	})
	if err != nil {
		return err
	}
	return r.ReloadPolicy()
}

func (r *adminRepository) ReplacePermissionReferences(
	ctx context.Context,
	replacements map[model.Permission]model.Permission,
) error {
	for oldPermission, newPermission := range replacements {
		if oldPermission == newPermission {
			continue
		}
		var rules []casbinRule
		if err := r.DB(ctx).
			Where("ptype = ? AND v1 = ? AND v2 = ?", "p", oldPermission.Resource, oldPermission.Action).
			Find(&rules).Error; err != nil {
			return err
		}
		if len(rules) == 0 {
			continue
		}
		ids := make([]uint, 0, len(rules))
		for _, rule := range rules {
			ids = append(ids, rule.ID)
		}
		if err := r.DB(ctx).Where("id IN ?", ids).Delete(&casbinRule{}).Error; err != nil {
			return err
		}
		for index := range rules {
			rules[index].ID = 0
			rules[index].V1 = newPermission.Resource
			rules[index].V2 = newPermission.Action
		}
		if err := r.DB(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rules).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *adminRepository) DeletePermissionReferences(ctx context.Context, permissions []model.Permission) error {
	for _, permission := range permissions {
		if err := r.DB(ctx).
			Where("ptype = ? AND v1 = ? AND v2 = ?", "p", permission.Resource, permission.Action).
			Delete(&casbinRule{}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *adminRepository) ReloadPolicy() error {
	return r.e.LoadPolicy()
}

func (r *adminRepository) GetApiGroups(ctx context.Context) ([]string, error) {
	res := make([]string, 0)
	groupColumn := quotedColumn(r.DB(ctx), "group")
	if err := r.DB(ctx).Model(&model.Api{}).Distinct(groupColumn).Pluck(groupColumn, &res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *adminRepository) GetApis(ctx context.Context, req *v1.GetApisRequest) ([]model.Api, int64, error) {
	var list []model.Api
	var total int64
	scope := r.DB(ctx).Model(&model.Api{})
	groupColumn := quotedColumn(r.DB(ctx), "group")
	if req.Name != "" {
		scope = scope.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Group != "" {
		scope = scope.Where(fmt.Sprintf("%s LIKE ?", groupColumn), "%"+req.Group+"%")
	}
	if req.Path != "" {
		scope = scope.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.Method != "" {
		scope = scope.Where("method = ?", req.Method)
	}
	if err := scope.Count(&total).Error; err != nil {
		return nil, total, err
	}
	if err := scope.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Order(groupColumn + " ASC").Order("id ASC").Find(&list).Error; err != nil {
		return nil, total, err
	}
	return list, total, nil
}

func (r *adminRepository) GetApi(ctx context.Context, id uint) (model.Api, error) {
	api := model.Api{}
	return api, r.DB(ctx).Where("id = ?", id).First(&api).Error
}

func (r *adminRepository) GetApiList(ctx context.Context) ([]model.Api, error) {
	var list []model.Api
	return list, r.DB(ctx).Order("id ASC").Find(&list).Error
}

func (r *adminRepository) ApiPermissionExists(
	ctx context.Context,
	path string,
	method string,
	excludeID uint,
) (bool, error) {
	query := r.DB(ctx).Model(&model.Api{}).Where("path = ? AND method = ?", path, method)
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *adminRepository) ApiUpdate(ctx context.Context, m *model.Api) error {
	tx := r.DB(ctx).Model(&model.Api{}).Where("id = ?", m.ID).Select(
		"group",
		"name",
		"path",
		"method",
		"menu_ids",
	).Updates(m)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return r.ensureRecordExists(ctx, &model.Api{}, m.ID)
	}
	return nil
}

func (r *adminRepository) ApiCreate(ctx context.Context, m *model.Api) error {
	return r.DB(ctx).Create(m).Error
}

func (r *adminRepository) ApiDelete(ctx context.Context, id uint) error {
	tx := r.DB(ctx).Unscoped().Where("id = ?", id).Delete(&model.Api{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *adminRepository) RemoveMenuFromApis(ctx context.Context, menuID uint) error {
	var apis []model.Api
	if err := r.DB(ctx).Find(&apis).Error; err != nil {
		return err
	}
	for _, api := range apis {
		menuIDs := make([]uint, 0, len(api.MenuIDs))
		changed := false
		for _, id := range api.MenuIDs {
			if id == menuID {
				changed = true
				continue
			}
			menuIDs = append(menuIDs, id)
		}
		if changed {
			if err := r.DB(ctx).Model(&model.Api{}).Where("id = ?", api.ID).
				Select("menu_ids").Updates(&model.Api{MenuIDs: menuIDs}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *adminRepository) GetUserPermissions(ctx context.Context, uid uint) ([][]string, error) {
	return r.e.GetImplicitPermissionsForUser(convertor.ToString(uid))

}
func (r *adminRepository) GetRolePermissions(ctx context.Context, role string) ([][]string, error) {
	return r.e.GetPermissionsForUser(role)
}
func (r *adminRepository) GetUserRoles(ctx context.Context, uid uint) ([]string, error) {
	return r.e.GetRolesForUser(convertor.ToString(uid))
}
func (r *adminRepository) MenuUpdate(ctx context.Context, m *model.Menu) error {
	tx := r.DB(ctx).Where("id = ?", m.ID).Select(
		"parent_id",
		"path",
		"title",
		"name",
		"component",
		"locale",
		"icon",
		"redirect",
		"url",
		"link",
		"target",
		"active_path",
		"show_text_badge",
		"weight",
		"is_enable",
		"is_menu",
		"keep_alive",
		"hide_in_menu",
		"is_hide",
		"is_hide_tab",
		"is_iframe",
		"show_badge",
		"fixed_tab",
		"is_full_page",
		"roles",
		"auth_list",
	).Updates(m)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return r.ensureRecordExists(ctx, &model.Menu{}, m.ID)
	}
	return nil
}

func (r *adminRepository) MenuCreate(ctx context.Context, m *model.Menu) error {
	return r.DB(ctx).Create(m).Error
}

func (r *adminRepository) MenuDelete(ctx context.Context, id uint) error {
	tx := r.DB(ctx).Where("id = ?", id).Delete(&model.Menu{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *adminRepository) GetMenuList(ctx context.Context) ([]model.Menu, error) {
	var menuList []model.Menu
	if err := r.DB(ctx).Order("weight DESC").Find(&menuList).Error; err != nil {
		return nil, err
	}
	return menuList, nil
}

func (r *adminRepository) RoleUpdate(ctx context.Context, m *model.Role) error {
	tx := r.DB(ctx).Model(&model.Role{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"name": m.Name,
	})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return r.ensureRecordExists(ctx, &model.Role{}, m.ID)
	}
	return nil
}

func (r *adminRepository) RoleCreate(ctx context.Context, m *model.Role) error {
	return r.DB(ctx).Create(m).Error
}

func (r *adminRepository) RoleDelete(ctx context.Context, id uint) error {
	tx := r.DB(ctx).Where("id = ?", id).Delete(&model.Role{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *adminRepository) GetRoles(ctx context.Context, req *v1.GetRoleListRequest) ([]model.Role, int64, error) {
	var list []model.Role
	var total int64
	scope := r.DB(ctx).Model(&model.Role{})
	if req.Name != "" {
		scope = scope.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Sid != "" {
		scope = scope.Where("sid = ?", req.Sid)
	}
	if err := scope.Count(&total).Error; err != nil {
		return nil, total, err
	}
	if err := scope.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, total, err
	}
	return list, total, nil
}

func quotedColumn(db *gorm.DB, name string) string {
	stmt := &gorm.Statement{DB: db}
	return stmt.Quote(clause.Column{Name: name})
}

func (r *adminRepository) ensureRecordExists(ctx context.Context, modelValue interface{}, id uint) error {
	var count int64
	if err := r.DB(ctx).Model(modelValue).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
