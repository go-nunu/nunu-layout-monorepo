package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	v1 "nunu-layout-monorepo/app/admin/api/v1"
	"nunu-layout-monorepo/app/admin/internal/repository"
	"nunu-layout-monorepo/model"
	"sort"
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
	role := strings.TrimSpace(req.Role)
	if role == "" {
		return v1.ErrBadRequest
	}
	if _, err := s.adminRepository.GetRoleBySid(ctx, role); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.ErrBadRequest
		}
		return err
	}

	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	menuPaths, err := buildMenuPathIndex(menus)
	if err != nil {
		return err
	}
	apis, err := s.adminRepository.GetApiList(ctx)
	if err != nil {
		return err
	}

	validPermissions := make(map[string]model.Permission, len(menus)+len(apis))
	menuIDByPermission := make(map[string]uint, len(menus))
	for _, menu := range menus {
		permission := menuPermission(menuPaths[menu.ID])
		validPermissions[permission.Key()] = permission
		menuIDByPermission[permission.Key()] = menu.ID
	}
	for _, api := range apis {
		permission := apiPermission(api.Path, api.Method)
		validPermissions[permission.Key()] = permission
	}

	permissionSet := make(map[model.Permission]struct{}, len(req.List))
	selectedMenuIDs := make(map[uint]struct{})
	for _, key := range req.List {
		permission, ok := model.ParsePermissionKey(key)
		if !ok {
			return v1.ErrBadRequest
		}
		valid, exists := validPermissions[permission.Key()]
		if !exists || valid != permission {
			return v1.ErrBadRequest
		}
		permissionSet[permission] = struct{}{}
		if menuID, isMenu := menuIDByPermission[permission.Key()]; isMenu {
			selectedMenuIDs[menuID] = struct{}{}
		}
	}

	selectedMenuIDs = includeMenuDescendants(menus, selectedMenuIDs)
	for menuID := range selectedMenuIDs {
		permissionSet[menuPermission(menuPaths[menuID])] = struct{}{}
	}
	for _, api := range apis {
		if intersectsMenuIDs(api.MenuIDs, selectedMenuIDs) {
			permissionSet[apiPermission(api.Path, api.Method)] = struct{}{}
		}
	}

	permissions := make([]model.Permission, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].Key() < permissions[j].Key()
	})
	return s.adminRepository.UpdateRolePermission(ctx, role, permissions)
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
			MenuIDs:   append([]uint{}, api.MenuIDs...),
			Method:    api.Method,
			Name:      api.Name,
			Path:      api.Path,
			UpdatedAt: api.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return data, nil
}

func (s *adminService) ApiUpdate(ctx context.Context, req *v1.ApiUpdateRequest) error {
	api, err := normalizeApiInput(req.Group, req.Name, req.Path, req.Method)
	if err != nil {
		return err
	}
	api.ID = req.ID
	menuIDs := uniqueUintIDs(req.MenuIDs)
	if err := s.validateApiMenuIDs(ctx, menuIDs); err != nil {
		return err
	}
	api.MenuIDs = menuIDs
	oldApi, err := s.adminRepository.GetApi(ctx, req.ID)
	if err != nil {
		return err
	}
	exists, err := s.adminRepository.ApiPermissionExists(ctx, api.Path, api.Method, req.ID)
	if err != nil {
		return err
	}
	if exists {
		return v1.ErrBadRequest
	}

	oldPermission := apiPermission(oldApi.Path, oldApi.Method)
	newPermission := apiPermission(api.Path, api.Method)
	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.ApiUpdate(ctx, api); err != nil {
			return err
		}
		return s.adminRepository.ReplacePermissionReferences(ctx, map[model.Permission]model.Permission{
			oldPermission: newPermission,
		})
	})
	if err != nil {
		return err
	}
	if oldPermission != newPermission {
		return s.adminRepository.ReloadPolicy()
	}
	return nil
}

func (s *adminService) ApiCreate(ctx context.Context, req *v1.ApiCreateRequest) error {
	api, err := normalizeApiInput(req.Group, req.Name, req.Path, req.Method)
	if err != nil {
		return err
	}
	menuIDs := uniqueUintIDs(req.MenuIDs)
	if err := s.validateApiMenuIDs(ctx, menuIDs); err != nil {
		return err
	}
	exists, err := s.adminRepository.ApiPermissionExists(ctx, api.Path, api.Method, 0)
	if err != nil {
		return err
	}
	if exists {
		return v1.ErrBadRequest
	}
	api.MenuIDs = menuIDs
	return s.adminRepository.ApiCreate(ctx, api)
}

func (s *adminService) validateApiMenuIDs(ctx context.Context, menuIDs []uint) error {
	if len(menuIDs) == 0 {
		return nil
	}
	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	validIDs := make(map[uint]struct{}, len(menus))
	for _, menu := range menus {
		validIDs[menu.ID] = struct{}{}
	}
	for _, id := range menuIDs {
		if _, exists := validIDs[id]; !exists {
			return v1.ErrBadRequest
		}
	}
	return nil
}

func normalizeApiGroup(group string) (string, error) {
	parts := strings.Split(strings.TrimSpace(group), "/")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return "", v1.ErrBadRequest
		}
		normalized = append(normalized, part)
	}
	if len(normalized) == 0 {
		return "", v1.ErrBadRequest
	}
	result := strings.Join(normalized, "/")
	if len(result) > 255 {
		return "", v1.ErrBadRequest
	}
	return result, nil
}

func normalizeApiInput(group, name, path, method string) (*model.Api, error) {
	normalizedGroup, err := normalizeApiGroup(group)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	path = strings.TrimSpace(path)
	method = strings.ToUpper(strings.TrimSpace(method))
	if name == "" || len(name) > 100 || path == "" || len(path) > 255 || !strings.HasPrefix(path, "/") {
		return nil, v1.ErrBadRequest
	}
	allowedMethods := map[string]struct{}{
		http.MethodGet: {}, http.MethodPost: {}, http.MethodPut: {}, http.MethodPatch: {},
		http.MethodDelete: {}, http.MethodHead: {}, http.MethodOptions: {},
	}
	if _, ok := allowedMethods[method]; !ok {
		return nil, v1.ErrBadRequest
	}
	return &model.Api{Group: normalizedGroup, Name: name, Path: path, Method: method}, nil
}

func uniqueUintIDs(ids []uint) []uint {
	result := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func apiPermission(path, method string) model.Permission {
	return model.Permission{
		Resource: model.ApiResourcePrefix + strings.TrimSpace(path),
		Action:   strings.ToUpper(strings.TrimSpace(method)),
	}
}

func menuPermission(path string) model.Permission {
	return model.Permission{Resource: model.MenuResourcePrefix + path, Action: "read"}
}

func intersectsMenuIDs(menuIDs []uint, selected map[uint]struct{}) bool {
	for _, menuID := range menuIDs {
		if _, ok := selected[menuID]; ok {
			return true
		}
	}
	return false
}

func includeMenuDescendants(menus []model.Menu, selected map[uint]struct{}) map[uint]struct{} {
	result := make(map[uint]struct{}, len(selected))
	for id := range selected {
		result[id] = struct{}{}
	}
	changed := true
	for changed {
		changed = false
		for _, menu := range menus {
			if _, parentSelected := result[menu.ParentID]; !parentSelected {
				continue
			}
			if _, exists := result[menu.ID]; exists {
				continue
			}
			result[menu.ID] = struct{}{}
			changed = true
		}
	}
	return result
}

func replaceMenu(menus []model.Menu, replacement model.Menu) ([]model.Menu, bool) {
	result := append([]model.Menu(nil), menus...)
	for index := range result {
		if result[index].ID == replacement.ID {
			result[index] = replacement
			return result, true
		}
	}
	return result, false
}

func validateMenuCollection(menus []model.Menu) (map[uint]string, error) {
	names := make(map[string]uint, len(menus))
	for _, menu := range menus {
		name := strings.TrimSpace(menu.Name)
		path := strings.TrimSpace(menu.Path)
		title := strings.TrimSpace(menu.Title)
		if name == "" || len(name) > 100 || path == "" || len(path) > 255 || title == "" || len(title) > 100 {
			return nil, v1.ErrBadRequest
		}
		if existingID, exists := names[name]; exists && existingID != menu.ID {
			return nil, v1.ErrBadRequest
		}
		names[name] = menu.ID
	}
	paths, err := buildMenuPathIndex(menus)
	if err != nil {
		return nil, v1.ErrBadRequest
	}
	pathOwners := make(map[string]uint, len(paths))
	for id, path := range paths {
		if existingID, exists := pathOwners[path]; exists && existingID != id {
			return nil, v1.ErrBadRequest
		}
		pathOwners[path] = id
	}
	return paths, nil
}

func buildMenuPathIndex(menus []model.Menu) (map[uint]string, error) {
	menuMap := make(map[uint]model.Menu, len(menus))
	for _, menu := range menus {
		if _, exists := menuMap[menu.ID]; exists {
			return nil, fmt.Errorf("duplicate menu id: %d", menu.ID)
		}
		menuMap[menu.ID] = menu
	}
	paths := make(map[uint]string, len(menus))
	states := make(map[uint]uint8, len(menus))
	var resolve func(uint) (string, error)
	resolve = func(id uint) (string, error) {
		switch states[id] {
		case 1:
			return "", fmt.Errorf("menu hierarchy cycle at id %d", id)
		case 2:
			return paths[id], nil
		}
		menu, exists := menuMap[id]
		if !exists {
			return "", fmt.Errorf("menu %d not found", id)
		}
		states[id] = 1
		parentPath := ""
		if menu.ParentID != 0 {
			if _, exists := menuMap[menu.ParentID]; !exists {
				return "", fmt.Errorf("parent menu %d not found", menu.ParentID)
			}
			var err error
			parentPath, err = resolve(menu.ParentID)
			if err != nil {
				return "", err
			}
		}
		path := strings.TrimSpace(menu.Path)
		switch {
		case strings.HasPrefix(path, "http://"), strings.HasPrefix(path, "https://"), strings.HasPrefix(path, "/"):
			paths[id] = path
		case menu.ParentID == 0:
			paths[id] = "/" + strings.TrimLeft(path, "/")
		default:
			paths[id] = strings.TrimRight(parentPath, "/") + "/" + strings.TrimLeft(path, "/")
		}
		states[id] = 2
		return paths[id], nil
	}
	for id := range menuMap {
		if _, err := resolve(id); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

func (s *adminService) ApiDelete(ctx context.Context, id uint) error {
	api, err := s.adminRepository.GetApi(ctx, id)
	if err != nil {
		return err
	}
	permission := apiPermission(api.Path, api.Method)
	if err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.ApiDelete(ctx, id); err != nil {
			return err
		}
		return s.adminRepository.DeletePermissionReferences(ctx, []model.Permission{permission})
	}); err != nil {
		return err
	}
	return s.adminRepository.ReloadPolicy()
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
	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	oldPaths, err := buildMenuPathIndex(menus)
	if err != nil {
		return err
	}
	updatedMenus, found := replaceMenu(menus, *menu)
	if !found {
		return gorm.ErrRecordNotFound
	}
	newPaths, err := validateMenuCollection(updatedMenus)
	if err != nil {
		return err
	}
	replacements := make(map[model.Permission]model.Permission)
	for id, oldPath := range oldPaths {
		newPath := newPaths[id]
		if oldPath != newPath {
			replacements[menuPermission(oldPath)] = menuPermission(newPath)
		}
	}
	if err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.MenuUpdate(ctx, menu); err != nil {
			return err
		}
		return s.adminRepository.ReplacePermissionReferences(ctx, replacements)
	}); err != nil {
		return err
	}
	if len(replacements) > 0 {
		return s.adminRepository.ReloadPolicy()
	}
	return nil
}

func (s *adminService) MenuCreate(ctx context.Context, req *v1.MenuCreateRequest) error {
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
	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	if _, err := validateMenuCollection(append(menus, *menu)); err != nil {
		return err
	}
	return s.adminRepository.MenuCreate(ctx, menu)
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
		Component:     strings.TrimSpace(payload.Component),
		Icon:          strings.TrimSpace(payload.Icon),
		KeepAlive:     payload.KeepAlive,
		HideInMenu:    isHide,
		IsHide:        isHide,
		Locale:        strings.TrimSpace(payload.Locale),
		Weight:        payload.Weight,
		Name:          strings.TrimSpace(payload.Name),
		ParentID:      payload.ParentID,
		Path:          strings.TrimSpace(payload.Path),
		Redirect:      strings.TrimSpace(payload.Redirect),
		Title:         strings.TrimSpace(payload.Title),
		URL:           strings.TrimSpace(payload.URL),
		Link:          strings.TrimSpace(payload.Link),
		Target:        strings.TrimSpace(payload.Target),
		ActivePath:    strings.TrimSpace(payload.ActivePath),
		ShowTextBadge: strings.TrimSpace(payload.ShowTextBadge),
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
	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	paths, err := buildMenuPathIndex(menus)
	if err != nil {
		return err
	}
	path, exists := paths[id]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	for _, menu := range menus {
		if menu.ParentID == id {
			return v1.ErrBadRequest
		}
	}
	permission := menuPermission(path)
	if err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.RemoveMenuFromApis(ctx, id); err != nil {
			return err
		}
		if err := s.adminRepository.MenuDelete(ctx, id); err != nil {
			return err
		}
		return s.adminRepository.DeletePermissionReferences(ctx, []model.Permission{permission})
	}); err != nil {
		return err
	}
	return s.adminRepository.ReloadPolicy()
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
	menuPaths, err := buildMenuPathIndex(menuList)
	if err != nil {
		return nil, err
	}

	allowedMenuIDs := map[uint]struct{}{}
	for _, menu := range menuList {
		fullPath := menuPaths[menu.ID]
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
	seen := make(map[uint]struct{})
	current := menu
	for {
		if _, exists := seen[current.ID]; exists {
			return
		}
		seen[current.ID] = struct{}{}
		allowed[current.ID] = struct{}{}
		if current.ParentID == 0 {
			return
		}
		parent, ok := menuMap[current.ParentID]
		if !ok {
			return
		}
		current = parent
	}
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
