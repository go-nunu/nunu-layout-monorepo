package service

import (
	"context"
	"errors"
	"github.com/duke-git/lancet/v2/convertor"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	v1 "nunu-layout-monorepo/app/admin/api/v1"
	"nunu-layout-monorepo/app/admin/internal/repository"
	"nunu-layout-monorepo/model"
	"strings"
	"time"
)

type AdminService interface {
	Login(ctx context.Context, req *v1.LoginRequest) (string, error)
	GetAdminUsers(ctx context.Context, req *v1.GetAdminUsersRequest) (*v1.GetAdminUsersResponseData, error)
	GetAdminUser(ctx context.Context, uid uint) (*v1.GetAdminUserResponseData, error)
	AdminUserUpdate(ctx context.Context, req *v1.AdminUserUpdateRequest) error
	AdminUserCreate(ctx context.Context, req *v1.AdminUserCreateRequest) error
	AdminUserDelete(ctx context.Context, id uint) error

	GetUserPermissions(ctx context.Context, uid uint) (*v1.GetUserPermissionsData, error)
	GetRolePermissions(ctx context.Context, role string) (*v1.GetRolePermissionsData, error)
	UpdateRolePermission(ctx context.Context, req *v1.UpdateRolePermissionRequest) error

	GetAdminMenus(ctx context.Context) (*v1.GetMenuResponseData, error)
	GetMenus(ctx context.Context, uid uint) (*v1.GetMenuResponseData, error)
	MenuUpdate(ctx context.Context, req *v1.MenuUpdateRequest) error
	MenuCreate(ctx context.Context, req *v1.MenuCreateRequest) error
	MenuDelete(ctx context.Context, id uint) error

	GetRoles(ctx context.Context, req *v1.GetRoleListRequest) (*v1.GetRolesResponseData, error)
	RoleUpdate(ctx context.Context, req *v1.RoleUpdateRequest) error
	RoleCreate(ctx context.Context, req *v1.RoleCreateRequest) error
	RoleDelete(ctx context.Context, id uint) error

	GetApis(ctx context.Context, req *v1.GetApisRequest) (*v1.GetApisResponseData, error)
	ApiUpdate(ctx context.Context, req *v1.ApiUpdateRequest) error
	ApiCreate(ctx context.Context, req *v1.ApiCreateRequest) error
	ApiDelete(ctx context.Context, id uint) error
}

func NewAdminService(
	service *Service,
	adminRepository repository.AdminRepository,
) AdminService {
	return &adminService{
		Service:         service,
		adminRepository: adminRepository,
	}
}

type adminService struct {
	*Service
	adminRepository repository.AdminRepository
}

func (s *adminService) GetAdminUser(ctx context.Context, uid uint) (*v1.GetAdminUserResponseData, error) {
	user, err := s.adminRepository.GetAdminUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	roles, _ := s.adminRepository.GetUserRoles(ctx, uid)

	return &v1.GetAdminUserResponseData{
		Email:     user.Email,
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Phone:     user.Phone,
		Roles:     roles,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *adminService) Login(ctx context.Context, req *v1.LoginRequest) (string, error) {
	user, err := s.adminRepository.GetAdminUserByUsername(ctx, req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", v1.ErrUnauthorized
		}
		return "", v1.ErrInternalServerError
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", err
	}
	token, err := s.jwt.GenToken(user.ID, time.Now().Add(time.Hour*24*90))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *adminService) GetAdminUsers(ctx context.Context, req *v1.GetAdminUsersRequest) (*v1.GetAdminUsersResponseData, error) {
	list, total, err := s.adminRepository.GetAdminUsers(ctx, req)
	if err != nil {
		return nil, err
	}
	data := &v1.GetAdminUsersResponseData{
		List:  make([]v1.AdminUserDataItem, 0),
		Total: total,
	}
	for _, user := range list {
		roles, err := s.adminRepository.GetUserRoles(ctx, user.ID)
		if err != nil {
			s.logger.Error("GetUserRoles error", zap.Error(err))
			continue
		}
		data.List = append(data.List, v1.AdminUserDataItem{
			Email:     user.Email,
			ID:        user.ID,
			Nickname:  user.Nickname,
			Username:  user.Username,
			Phone:     user.Phone,
			Roles:     roles,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return data, nil
}

func (s *adminService) AdminUserUpdate(ctx context.Context, req *v1.AdminUserUpdateRequest) error {
	old, err := s.adminRepository.GetAdminUser(ctx, req.ID)
	if err != nil {
		return err
	}
	password := old.Password
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		password = string(hash)
	}
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.AdminUserUpdate(ctx, &model.AdminUser{
			Model: gorm.Model{
				ID: req.ID,
			},
			Email:    req.Email,
			Nickname: req.Nickname,
			Password: password,
			Phone:    req.Phone,
			Username: req.Username,
		}); err != nil {
			return err
		}
		return s.adminRepository.UpdateUserRoles(ctx, req.ID, req.Roles)
	})

}

func (s *adminService) AdminUserCreate(ctx context.Context, req *v1.AdminUserCreateRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	req.Password = string(hash)
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.AdminUserCreate(ctx, &model.AdminUser{
			Email:    req.Email,
			Nickname: req.Nickname,
			Phone:    req.Phone,
			Username: req.Username,
			Password: req.Password,
		}); err != nil {
			return err
		}
		user, err := s.adminRepository.GetAdminUserByUsername(ctx, req.Username)
		if err != nil {
			return err
		}
		return s.adminRepository.UpdateUserRoles(ctx, user.ID, req.Roles)
	})

}

func (s *adminService) AdminUserDelete(ctx context.Context, id uint) error {
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.AdminUserDelete(ctx, id); err != nil {
			return err
		}
		return s.adminRepository.DeleteUserRoles(ctx, id)
	})
}

func (s *adminService) UpdateRolePermission(ctx context.Context, req *v1.UpdateRolePermissionRequest) error {
	permissions := map[string]struct{}{}
	for _, v := range req.List {
		perm := strings.SplitN(v, model.PermSep, 2)
		if len(perm) != 2 || strings.TrimSpace(perm[0]) == "" || strings.TrimSpace(perm[1]) == "" {
			return v1.ErrBadRequest
		}
		permissions[v] = struct{}{}
	}
	return s.adminRepository.UpdateRolePermission(ctx, req.Role, permissions)
}

func (s *adminService) GetApis(ctx context.Context, req *v1.GetApisRequest) (*v1.GetApisResponseData, error) {
	list, total, err := s.adminRepository.GetApis(ctx, req)
	if err != nil {
		return nil, err
	}
	groups, err := s.adminRepository.GetApiGroups(ctx)
	if err != nil {
		return nil, err
	}
	data := &v1.GetApisResponseData{
		List:   make([]v1.ApiDataItem, 0),
		Total:  total,
		Groups: groups,
	}
	for _, api := range list {
		data.List = append(data.List, v1.ApiDataItem{
			CreatedAt: api.CreatedAt.Format("2006-01-02 15:04:05"),
			Group:     api.Group,
			ID:        api.ID,
			Method:    api.Method,
			Name:      api.Name,
			Path:      api.Path,
			UpdatedAt: api.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return data, nil
}

func (s *adminService) ApiUpdate(ctx context.Context, req *v1.ApiUpdateRequest) error {
	return s.adminRepository.ApiUpdate(ctx, &model.Api{
		Group:  req.Group,
		Method: req.Method,
		Name:   req.Name,
		Path:   req.Path,
		Model: gorm.Model{
			ID: req.ID,
		},
	})
}

func (s *adminService) ApiCreate(ctx context.Context, req *v1.ApiCreateRequest) error {
	return s.adminRepository.ApiCreate(ctx, &model.Api{
		Group:  req.Group,
		Method: req.Method,
		Name:   req.Name,
		Path:   req.Path,
	})
}

func (s *adminService) ApiDelete(ctx context.Context, id uint) error {
	return s.adminRepository.ApiDelete(ctx, id)
}

func (s *adminService) GetUserPermissions(ctx context.Context, uid uint) (*v1.GetUserPermissionsData, error) {
	data := &v1.GetUserPermissionsData{
		List: []string{},
	}
	list, err := s.adminRepository.GetUserPermissions(ctx, uid)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if len(v) == 3 {
			data.List = append(data.List, strings.Join([]string{v[1], v[2]}, model.PermSep))
		}
	}
	return data, nil
}
func (s *adminService) GetRolePermissions(ctx context.Context, role string) (*v1.GetRolePermissionsData, error) {
	data := &v1.GetRolePermissionsData{
		List: []string{},
	}
	list, err := s.adminRepository.GetRolePermissions(ctx, role)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if len(v) == 3 {
			data.List = append(data.List, strings.Join([]string{v[1], v[2]}, model.PermSep))
		}
	}
	return data, nil
}

func (s *adminService) MenuUpdate(ctx context.Context, req *v1.MenuUpdateRequest) error {
	menu := menuFromPayload(menuPayload{
		ParentID:      req.ParentID,
		Weight:        req.Weight,
		Path:          req.Path,
		Title:         req.Title,
		Name:          req.Name,
		Component:     req.Component,
		Locale:        req.Locale,
		Icon:          req.Icon,
		Redirect:      req.Redirect,
		KeepAlive:     req.KeepAlive,
		HideInMenu:    req.HideInMenu,
		IsEnable:      req.IsEnable,
		IsMenu:        req.IsMenu,
		IsHide:        req.IsHide,
		IsHideTab:     req.IsHideTab,
		Link:          req.Link,
		IsIframe:      req.IsIframe,
		ShowBadge:     req.ShowBadge,
		ShowTextBadge: req.ShowTextBadge,
		FixedTab:      req.FixedTab,
		ActivePath:    req.ActivePath,
		Roles:         req.Roles,
		IsFullPage:    req.IsFullPage,
		AuthList:      req.AuthList,
		Target:        req.Target,
		URL:           req.URL,
	})
	menu.Model = gorm.Model{ID: req.ID}
	return s.adminRepository.MenuUpdate(ctx, menu)
}

func (s *adminService) MenuCreate(ctx context.Context, req *v1.MenuCreateRequest) error {
	return s.adminRepository.MenuCreate(ctx, menuFromPayload(menuPayload{
		ParentID:      req.ParentID,
		Weight:        req.Weight,
		Path:          req.Path,
		Title:         req.Title,
		Name:          req.Name,
		Component:     req.Component,
		Locale:        req.Locale,
		Icon:          req.Icon,
		Redirect:      req.Redirect,
		KeepAlive:     req.KeepAlive,
		HideInMenu:    req.HideInMenu,
		IsEnable:      req.IsEnable,
		IsMenu:        req.IsMenu,
		IsHide:        req.IsHide,
		IsHideTab:     req.IsHideTab,
		Link:          req.Link,
		IsIframe:      req.IsIframe,
		ShowBadge:     req.ShowBadge,
		ShowTextBadge: req.ShowTextBadge,
		FixedTab:      req.FixedTab,
		ActivePath:    req.ActivePath,
		Roles:         req.Roles,
		IsFullPage:    req.IsFullPage,
		AuthList:      req.AuthList,
		Target:        req.Target,
		URL:           req.URL,
	}))
}

type menuPayload struct {
	ParentID      uint
	Weight        int
	Path          string
	Title         string
	Name          string
	Component     string
	Locale        string
	Icon          string
	Redirect      string
	KeepAlive     bool
	HideInMenu    bool
	IsEnable      bool
	IsMenu        bool
	IsHide        bool
	IsHideTab     bool
	Link          string
	IsIframe      bool
	ShowBadge     bool
	ShowTextBadge string
	FixedTab      bool
	ActivePath    string
	Roles         []string
	IsFullPage    bool
	AuthList      []v1.MenuAuthDataItem
	Target        string
	URL           string
}

func menuFromPayload(payload menuPayload) *model.Menu {
	isHide := payload.IsHide || payload.HideInMenu
	return &model.Menu{
		Component:     payload.Component,
		Icon:          payload.Icon,
		KeepAlive:     payload.KeepAlive,
		HideInMenu:    isHide,
		IsHide:        isHide,
		Locale:        payload.Locale,
		Weight:        payload.Weight,
		Name:          payload.Name,
		ParentID:      payload.ParentID,
		Path:          payload.Path,
		Redirect:      payload.Redirect,
		Title:         payload.Title,
		URL:           payload.URL,
		Link:          payload.Link,
		Target:        payload.Target,
		ActivePath:    payload.ActivePath,
		ShowTextBadge: payload.ShowTextBadge,
		IsEnable:      payload.IsEnable,
		IsMenu:        payload.IsMenu,
		IsHideTab:     payload.IsHideTab,
		IsIframe:      payload.IsIframe,
		ShowBadge:     payload.ShowBadge,
		FixedTab:      payload.FixedTab,
		IsFullPage:    payload.IsFullPage,
		Roles:         payload.Roles,
		AuthList:      menuAuthListFromPayload(payload.AuthList),
	}
}

func menuAuthListFromPayload(list []v1.MenuAuthDataItem) []model.MenuAuth {
	authList := make([]model.MenuAuth, 0, len(list))
	for _, item := range list {
		if item.Title == "" && item.AuthMark == "" {
			continue
		}
		authList = append(authList, model.MenuAuth{
			Title:    item.Title,
			AuthMark: item.AuthMark,
		})
	}
	return authList
}

func menuDataItemFromModel(menu model.Menu) v1.MenuDataItem {
	isHide := menu.IsHide || menu.HideInMenu
	return v1.MenuDataItem{
		ID:            menu.ID,
		Name:          menu.Name,
		Title:         menu.Title,
		Path:          menu.Path,
		Component:     menu.Component,
		Redirect:      menu.Redirect,
		KeepAlive:     menu.KeepAlive,
		HideInMenu:    isHide,
		IsHide:        isHide,
		IsEnable:      menu.IsEnable,
		IsMenu:        menu.IsMenu,
		IsHideTab:     menu.IsHideTab,
		Link:          menu.Link,
		IsIframe:      menu.IsIframe,
		ShowBadge:     menu.ShowBadge,
		ShowTextBadge: menu.ShowTextBadge,
		FixedTab:      menu.FixedTab,
		ActivePath:    menu.ActivePath,
		Roles:         menu.Roles,
		IsFullPage:    menu.IsFullPage,
		AuthList:      menuAuthDataListFromModel(menu.AuthList),
		Target:        menu.Target,
		Locale:        menu.Locale,
		Weight:        menu.Weight,
		Icon:          menu.Icon,
		ParentID:      menu.ParentID,
		UpdatedAt:     menu.UpdatedAt.Format("2006-01-02 15:04:05"),
		URL:           menu.URL,
	}
}

func menuAuthDataListFromModel(list []model.MenuAuth) []v1.MenuAuthDataItem {
	authList := make([]v1.MenuAuthDataItem, 0, len(list))
	for _, item := range list {
		authList = append(authList, v1.MenuAuthDataItem{
			Title:    item.Title,
			AuthMark: item.AuthMark,
		})
	}
	return authList
}

func (s *adminService) MenuDelete(ctx context.Context, id uint) error {
	return s.adminRepository.MenuDelete(ctx, id)
}

func (s *adminService) GetMenus(ctx context.Context, uid uint) (*v1.GetMenuResponseData, error) {
	menuList, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Error("GetMenuList error", zap.Error(err))
		return nil, err
	}
	data := &v1.GetMenuResponseData{
		List: make([]v1.MenuDataItem, 0),
	}

	if convertor.ToString(uid) == model.AdminUserID {
		for _, menu := range menuList {
			data.List = append(data.List, menuDataItemFromModel(menu))
		}
		return data, nil
	}

	// 获取权限的菜单
	permissions, err := s.adminRepository.GetUserPermissions(ctx, uid)
	if err != nil {
		return nil, err
	}
	menuPermMap := map[string]struct{}{}
	for _, permission := range permissions {
		if len(permission) == 3 && strings.HasPrefix(permission[1], model.MenuResourcePrefix) {
			menuPermMap[strings.TrimPrefix(permission[1], model.MenuResourcePrefix)] = struct{}{}
		}
	}

	menuMap := make(map[uint]model.Menu, len(menuList))
	for _, menu := range menuList {
		menuMap[menu.ID] = menu
	}

	allowedMenuIDs := map[uint]struct{}{}
	for _, menu := range menuList {
		fullPath := fullMenuPath(menu, menuMap)
		_, hasFullPath := menuPermMap[fullPath]
		_, hasRawPath := menuPermMap[menu.Path]
		if hasFullPath || hasRawPath {
			allowMenuWithAncestors(menu, menuMap, allowedMenuIDs)
		}
	}

	for _, menu := range menuList {
		if _, ok := allowedMenuIDs[menu.ID]; ok {
			data.List = append(data.List, menuDataItemFromModel(menu))
		}
	}
	return data, nil
}

func allowMenuWithAncestors(menu model.Menu, menuMap map[uint]model.Menu, allowed map[uint]struct{}) {
	allowed[menu.ID] = struct{}{}
	if menu.ParentID == 0 {
		return
	}
	parent, ok := menuMap[menu.ParentID]
	if !ok {
		return
	}
	allowMenuWithAncestors(parent, menuMap, allowed)
}

func fullMenuPath(menu model.Menu, menuMap map[uint]model.Menu) string {
	if menu.Path == "" {
		return ""
	}
	if strings.HasPrefix(menu.Path, "http://") ||
		strings.HasPrefix(menu.Path, "https://") ||
		strings.HasPrefix(menu.Path, "/") {
		return menu.Path
	}
	if menu.ParentID == 0 {
		return "/" + strings.TrimPrefix(menu.Path, "/")
	}
	parent, ok := menuMap[menu.ParentID]
	if !ok {
		return "/" + strings.TrimPrefix(menu.Path, "/")
	}
	parentPath := strings.TrimRight(fullMenuPath(parent, menuMap), "/")
	childPath := strings.TrimLeft(menu.Path, "/")
	if parentPath == "" {
		return "/" + childPath
	}
	return parentPath + "/" + childPath
}

func (s *adminService) GetAdminMenus(ctx context.Context) (*v1.GetMenuResponseData, error) {
	menuList, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Error("GetMenuList error", zap.Error(err))
		return nil, err
	}
	data := &v1.GetMenuResponseData{
		List: make([]v1.MenuDataItem, 0),
	}
	for _, menu := range menuList {
		data.List = append(data.List, menuDataItemFromModel(menu))
	}
	return data, nil
}

func (s *adminService) RoleUpdate(ctx context.Context, req *v1.RoleUpdateRequest) error {
	return s.adminRepository.RoleUpdate(ctx, &model.Role{
		Name: req.Name,
		Sid:  req.Sid,
		Model: gorm.Model{
			ID: req.ID,
		},
	})
}

func (s *adminService) RoleCreate(ctx context.Context, req *v1.RoleCreateRequest) error {
	_, err := s.adminRepository.GetRoleBySid(ctx, req.Sid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.adminRepository.RoleCreate(ctx, &model.Role{
				Name: req.Name,
				Sid:  req.Sid,
			})
		} else {
			return err
		}
	}
	return v1.ErrRoleAlreadyUse
}

func (s *adminService) RoleDelete(ctx context.Context, id uint) error {
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		old, err := s.adminRepository.GetRole(ctx, id)
		if err != nil {
			return err
		}
		if err := s.adminRepository.RoleDelete(ctx, id); err != nil {
			return err
		}
		return s.adminRepository.CasbinRoleDelete(ctx, old.Sid)
	})
}

func (s *adminService) GetRoles(ctx context.Context, req *v1.GetRoleListRequest) (*v1.GetRolesResponseData, error) {
	list, total, err := s.adminRepository.GetRoles(ctx, req)
	if err != nil {
		return nil, err
	}
	data := &v1.GetRolesResponseData{
		List:  make([]v1.RoleDataItem, 0),
		Total: total,
	}
	for _, role := range list {
		data.List = append(data.List, v1.RoleDataItem{
			ID:        role.ID,
			Name:      role.Name,
			Sid:       role.Sid,
			UpdatedAt: role.UpdatedAt.Format("2006-01-02 15:04:05"),
			CreatedAt: role.CreatedAt.Format("2006-01-02 15:04:05"),
		})

	}
	return data, nil
}
