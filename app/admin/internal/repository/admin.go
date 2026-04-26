package repository

import (
	"context"
	"fmt"
	v1 "nunu-layout-monorepo/app/admin/api/v1"
	"nunu-layout-monorepo/model"
	"strings"

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
	UpdateRolePermission(ctx context.Context, role string, permissions map[string]struct{}) error
	UpdateUserRoles(ctx context.Context, uid uint, roles []string) error
	DeleteUserRoles(ctx context.Context, uid uint) error

	GetMenuList(ctx context.Context) ([]model.Menu, error)
	MenuUpdate(ctx context.Context, m *model.Menu) error
	MenuCreate(ctx context.Context, m *model.Menu) error
	MenuDelete(ctx context.Context, id uint) error

	GetRoles(ctx context.Context, req *v1.GetRoleListRequest) ([]model.Role, int64, error)
	RoleUpdate(ctx context.Context, m *model.Role) error
	RoleCreate(ctx context.Context, m *model.Role) error
	RoleDelete(ctx context.Context, id uint) error
	CasbinRoleDelete(ctx context.Context, role string) error
	GetRole(ctx context.Context, id uint) (model.Role, error)
	GetRoleBySid(ctx context.Context, sid string) (model.Role, error)

	GetApis(ctx context.Context, req *v1.GetApisRequest) ([]model.Api, int64, error)
	GetApiGroups(ctx context.Context) ([]string, error)
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

func (r *adminRepository) UpdateRolePermission(ctx context.Context, role string, newPermSet map[string]struct{}) error {
	// 获取当前角色的所有权限
	oldPermissions, err := r.e.GetPermissionsForUser(role)
	if err != nil {
		return err
	}

	// 将旧权限转换为 map 方便查找
	oldPermSet := make(map[string]struct{})
	for _, perm := range oldPermissions {
		if len(perm) == 3 {
			oldPermSet[strings.Join([]string{perm[1], perm[2]}, model.PermSep)] = struct{}{}
		}
	}

	// 找出需要删除的权限
	var removePermissions [][]string
	for key, _ := range oldPermSet {
		if _, exists := newPermSet[key]; !exists {
			removePermissions = append(removePermissions, strings.Split(key, model.PermSep))
		}
	}

	// 找出需要添加的权限
	var addPermissions [][]string
	for key, _ := range newPermSet {
		if _, exists := oldPermSet[key]; !exists {
			addPermissions = append(addPermissions, strings.Split(key, model.PermSep))
		}

	}

	// 先移除多余的权限（使用 DeletePermissionForUser 逐条删除）
	for _, perm := range removePermissions {
		_, err := r.e.DeletePermissionForUser(role, perm...)
		if err != nil {
			return fmt.Errorf("移除权限失败: %v", err)
		}
	}

	// 再添加新的权限
	if len(addPermissions) > 0 {
		_, err = r.e.AddPermissionsForUser(role, addPermissions...)
		if err != nil {
			return fmt.Errorf("添加新权限失败: %v", err)
		}
	}

	return nil
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
	if err := scope.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Order(groupColumn + " ASC").Find(&list).Error; err != nil {
		return nil, total, err
	}
	return list, total, nil
}

func (r *adminRepository) ApiUpdate(ctx context.Context, m *model.Api) error {
	tx := r.DB(ctx).Model(&model.Api{}).Where("id = ?", m.ID).Select(
		"Group",
		"Name",
		"Path",
		"Method",
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
	tx := r.DB(ctx).Where("id = ?", id).Delete(&model.Api{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
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
	return r.DB(ctx).Where("id = ?", m.ID).Select(
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
	).Updates(m).Error
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
