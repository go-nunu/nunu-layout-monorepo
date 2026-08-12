package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	v1 "nunu-layout-monorepo/app/admin/api/v1"
	"nunu-layout-monorepo/model"
	"nunu-layout-monorepo/pkg/log"
	"nunu-layout-monorepo/pkg/sid"
	"os"
	"strings"
)

type MigrateServer struct {
	db  *gorm.DB
	log *log.Logger
	sid *sid.Sid
	e   *casbin.SyncedEnforcer
}

func NewMigrateServer(
	db *gorm.DB,
	log *log.Logger,
	sid *sid.Sid,
	e *casbin.SyncedEnforcer,
) *MigrateServer {
	return &MigrateServer{
		e:   e,
		db:  db,
		log: log,
		sid: sid,
	}
}
func (m *MigrateServer) Start(ctx context.Context) error {
	m.db.Migrator().DropTable(
		&model.AdminUser{},
		&model.Menu{},
		&model.Role{},
		&model.Api{},
	)
	if err := m.db.AutoMigrate(
		&model.AdminUser{},
		&model.Menu{},
		&model.Role{},
		&model.Api{},
	); err != nil {
		m.log.Error("user migrate error", zap.Error(err))
		return err
	}
	err := m.initialAdminUser(ctx)
	if err != nil {
		m.log.Error("initialAdminUser error", zap.Error(err))
	}

	err = m.initialMenuData(ctx)
	if err != nil {
		m.log.Error("initialMenuData error", zap.Error(err))
	}

	err = m.initialApisData(ctx)
	if err != nil {
		m.log.Error("initialApisData error", zap.Error(err))
	}

	err = m.initialRBAC(ctx)
	if err != nil {
		m.log.Error("initialRBAC error", zap.Error(err))
	}

	m.log.Info("AutoMigrate success")
	os.Exit(0)
	return nil
}
func (m *MigrateServer) Stop(ctx context.Context) error {
	m.log.Info("AutoMigrate stop")
	return nil
}
func (m *MigrateServer) initialAdminUser(ctx context.Context) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = m.db.Create(&model.AdminUser{
		Model:    gorm.Model{ID: 1},
		Username: "admin",
		Password: string(hashedPassword),
		Nickname: "Admin",
	}).Error
	return m.db.Create(&model.AdminUser{
		Model:    gorm.Model{ID: 2},
		Username: "user",
		Password: string(hashedPassword),
		Nickname: "运营人员",
	}).Error

}
func (m *MigrateServer) initialRBAC(ctx context.Context) error {
	roles := []model.Role{
		{Sid: model.AdminRole, Name: "超级管理员"},
		{Sid: "1000", Name: "运营人员"},
		{Sid: "1001", Name: "访客"},
	}
	if err := m.db.Create(&roles).Error; err != nil {
		return err
	}
	m.e.ClearPolicy()
	err := m.e.SavePolicy()
	if err != nil {
		m.log.Error("m.e.SavePolicy error", zap.Error(err))
		return err
	}
	_, err = m.e.AddRoleForUser(model.AdminUserID, model.AdminRole)
	if err != nil {
		m.log.Error("m.e.AddRoleForUser error", zap.Error(err))
		return err
	}
	menuList := make([]v1.MenuDataItem, 0)
	err = json.Unmarshal([]byte(menuData), &menuList)
	if err != nil {
		m.log.Error("json.Unmarshal error", zap.Error(err))
		return err
	}
	menuList = filterSeedMenus(menuList)
	seedMenuMap := make(map[uint]v1.MenuDataItem, len(menuList))
	for _, item := range menuList {
		seedMenuMap[item.ID] = item
	}
	for _, item := range menuList {
		m.addPermissionForRole(model.AdminRole, model.MenuResourcePrefix+fullSeedMenuPath(item, seedMenuMap), "read")
	}
	apiList := make([]model.Api, 0)
	err = m.db.Find(&apiList).Error
	if err != nil {
		m.log.Error("m.db.Find(&apiList).Error error", zap.Error(err))
		return err
	}
	for _, api := range apiList {
		m.addPermissionForRole(model.AdminRole, model.ApiResourcePrefix+api.Path, api.Method)
	}

	// 添加运营人员权限
	_, err = m.e.AddRoleForUser("2", "1000")
	if err != nil {
		m.log.Error("m.e.AddRoleForUser error", zap.Error(err))
		return err
	}
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard/console", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard/analysis", "read")
	m.addPermissionForRole("1000", model.ApiResourcePrefix+"/v1/menus", http.MethodGet)
	m.addPermissionForRole("1000", model.ApiResourcePrefix+"/v1/admin/user", http.MethodGet)

	return nil
}

func fullSeedMenuPath(item v1.MenuDataItem, menuMap map[uint]v1.MenuDataItem) string {
	if item.Path == "" {
		return ""
	}
	if strings.HasPrefix(item.Path, "http://") ||
		strings.HasPrefix(item.Path, "https://") ||
		strings.HasPrefix(item.Path, "/") {
		return item.Path
	}
	if item.ParentID == 0 {
		return "/" + strings.TrimPrefix(item.Path, "/")
	}
	parent, ok := menuMap[item.ParentID]
	if !ok {
		return "/" + strings.TrimPrefix(item.Path, "/")
	}
	parentPath := strings.TrimRight(fullSeedMenuPath(parent, menuMap), "/")
	childPath := strings.TrimLeft(item.Path, "/")
	if parentPath == "" {
		return "/" + childPath
	}
	return parentPath + "/" + childPath
}

func (m *MigrateServer) addPermissionForRole(role, resource, action string) {
	_, err := m.e.AddPermissionForUser(role, resource, action)
	if err != nil {
		m.log.Sugar().Info("为角色 %s 添加权限 %s:%s 失败: %v", role, resource, action, err)
		return
	}
	fmt.Printf("为角色 %s 添加权限: %s %s\n", role, resource, action)
}
func (m *MigrateServer) initialApisData(ctx context.Context) error {
	initialApis := []model.Api{
		{Group: "基础API", Name: "获取用户菜单列表", Path: "/v1/menus", Method: http.MethodGet},
		{Group: "基础API", Name: "获取管理员信息", Path: "/v1/admin/user", Method: http.MethodGet, MenuIDs: []uint{60}},

		{Group: "权限管理/菜单", Name: "获取管理菜单", Path: "/v1/admin/menus", Method: http.MethodGet, MenuIDs: []uint{62, 63, 64}},
		{Group: "权限管理/菜单", Name: "创建菜单", Path: "/v1/admin/menu", Method: http.MethodPost, MenuIDs: []uint{63}},
		{Group: "权限管理/菜单", Name: "更新菜单", Path: "/v1/admin/menu", Method: http.MethodPut, MenuIDs: []uint{63}},
		{Group: "权限管理/菜单", Name: "删除菜单", Path: "/v1/admin/menu", Method: http.MethodDelete, MenuIDs: []uint{63}},

		{Group: "权限管理/角色", Name: "获取用户权限", Path: "/v1/admin/user/permissions", Method: http.MethodGet, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "获取角色权限", Path: "/v1/admin/role/permissions", Method: http.MethodGet, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "更新角色权限", Path: "/v1/admin/role/permissions", Method: http.MethodPut, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "获取角色列表", Path: "/v1/admin/roles", Method: http.MethodGet, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "创建角色", Path: "/v1/admin/role", Method: http.MethodPost, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "更新角色", Path: "/v1/admin/role", Method: http.MethodPut, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "删除角色", Path: "/v1/admin/role", Method: http.MethodDelete, MenuIDs: []uint{62}},

		{Group: "权限管理/用户", Name: "获取管理员列表", Path: "/v1/admin/users", Method: http.MethodGet, MenuIDs: []uint{61}},
		{Group: "权限管理/用户", Name: "更新管理员信息", Path: "/v1/admin/user", Method: http.MethodPut, MenuIDs: []uint{61}},
		{Group: "权限管理/用户", Name: "创建管理员账号", Path: "/v1/admin/user", Method: http.MethodPost, MenuIDs: []uint{61}},
		{Group: "权限管理/用户", Name: "删除管理员", Path: "/v1/admin/user", Method: http.MethodDelete, MenuIDs: []uint{61}},

		{Group: "权限管理/接口", Name: "获取API列表", Path: "/v1/admin/apis", Method: http.MethodGet, MenuIDs: []uint{62, 64}},
		{Group: "权限管理/接口", Name: "创建API", Path: "/v1/admin/api", Method: http.MethodPost, MenuIDs: []uint{64}},
		{Group: "权限管理/接口", Name: "更新API", Path: "/v1/admin/api", Method: http.MethodPut, MenuIDs: []uint{64}},
		{Group: "权限管理/接口", Name: "删除API", Path: "/v1/admin/api", Method: http.MethodDelete, MenuIDs: []uint{64}},
	}

	return m.db.Create(&initialApis).Error
}
func (m *MigrateServer) initialMenuData(ctx context.Context) error {
	menuList := make([]v1.MenuDataItem, 0)
	err := json.Unmarshal([]byte(menuData), &menuList)
	if err != nil {
		m.log.Error("json.Unmarshal error", zap.Error(err))
		return err
	}
	menuList = filterSeedMenus(menuList)
	menuListDb := make([]model.Menu, 0)
	for _, item := range menuList {
		menuListDb = append(menuListDb, model.Menu{
			Model: gorm.Model{
				ID: item.ID,
			},
			ParentID:      item.ParentID,
			Path:          item.Path,
			Title:         item.Title,
			Name:          item.Name,
			Component:     item.Component,
			Locale:        item.Locale,
			Weight:        item.Weight,
			Icon:          item.Icon,
			Redirect:      item.Redirect,
			URL:           item.URL,
			Link:          item.Link,
			Target:        item.Target,
			ActivePath:    item.ActivePath,
			ShowTextBadge: item.ShowTextBadge,
			KeepAlive:     item.KeepAlive,
			HideInMenu:    item.HideInMenu || item.IsHide,
			IsHide:        item.HideInMenu || item.IsHide,
			IsHideTab:     item.IsHideTab,
			IsIframe:      item.IsIframe,
			ShowBadge:     item.ShowBadge,
			FixedTab:      item.FixedTab,
			IsFullPage:    item.IsFullPage,
			Roles:         item.Roles,
			AuthList:      menuAuthListFromSeed(item.AuthList),
			IsEnable:      true,
			IsMenu:        true,
		})
	}
	return m.db.Create(&menuListDb).Error
}

func menuAuthListFromSeed(list []v1.MenuAuthDataItem) []model.MenuAuth {
	authList := make([]model.MenuAuth, 0, len(list))
	for _, item := range list {
		authList = append(authList, model.MenuAuth{
			Title:    item.Title,
			AuthMark: item.AuthMark,
		})
	}
	return authList
}

func filterSeedMenus(menuList []v1.MenuDataItem) []v1.MenuDataItem {
	return menuList
}

var menuData = `[
  {"id":1,"parentId":0,"path":"/dashboard","name":"Dashboard","title":"menus.dashboard.title","component":"/index/index","icon":"ri:pie-chart-line","roles":["admin","R_SUPER","R_ADMIN"],"weight":100},
  {"id":2,"parentId":1,"path":"console","name":"Console","title":"menus.dashboard.console","component":"/dashboard/console","icon":"ri:home-smile-2-line","fixedTab":true,"weight":100},
  {"id":3,"parentId":1,"path":"analysis","name":"Analysis","title":"menus.dashboard.analysis","component":"/dashboard/analysis","icon":"ri:align-item-bottom-line","weight":90},
  {"id":4,"parentId":1,"path":"ecommerce","name":"Ecommerce","title":"menus.dashboard.ecommerce","component":"/dashboard/ecommerce","icon":"ri:bar-chart-box-line","weight":80},

  {"id":10,"parentId":0,"path":"/template","name":"Template","title":"menus.template.title","component":"/index/index","icon":"ri:apps-2-line","roles":["admin"],"weight":90},
  {"id":11,"parentId":10,"path":"cards","name":"Cards","title":"menus.template.cards","component":"/template/cards","icon":"ri:wallet-line","weight":100},
  {"id":12,"parentId":10,"path":"banners","name":"Banners","title":"menus.template.banners","component":"/template/banners","icon":"ri:rectangle-line","weight":90},
  {"id":13,"parentId":10,"path":"charts","name":"Charts","title":"menus.template.charts","component":"/template/charts","icon":"ri:bar-chart-box-line","weight":80},
  {"id":14,"parentId":10,"path":"map","name":"Map","title":"menus.template.map","component":"/template/map","icon":"ri:map-pin-line","keepAlive":true,"weight":70},
  {"id":15,"parentId":10,"path":"chat","name":"Chat","title":"menus.template.chat","component":"/template/chat","icon":"ri:message-3-line","keepAlive":true,"weight":60},
  {"id":16,"parentId":10,"path":"calendar","name":"Calendar","title":"menus.template.calendar","component":"/template/calendar","icon":"ri:calendar-2-line","keepAlive":true,"weight":50},
  {"id":17,"parentId":10,"path":"pricing","name":"Pricing","title":"menus.template.pricing","component":"/template/pricing","icon":"ri:money-cny-box-line","keepAlive":true,"isFullPage":true,"weight":40},

  {"id":20,"parentId":0,"path":"/widgets","name":"Widgets","title":"menus.widgets.title","component":"/index/index","icon":"ri:apps-2-add-line","roles":["admin"],"weight":80},
  {"id":21,"parentId":20,"path":"icon","name":"Icon","title":"menus.widgets.icon","component":"/widgets/icon","icon":"ri:palette-line","keepAlive":true,"weight":130},
  {"id":22,"parentId":20,"path":"image-crop","name":"ImageCrop","title":"menus.widgets.imageCrop","component":"/widgets/image-crop","icon":"ri:screenshot-line","keepAlive":true,"weight":120},
  {"id":23,"parentId":20,"path":"excel","name":"Excel","title":"menus.widgets.excel","component":"/widgets/excel","icon":"ri:download-2-line","keepAlive":true,"weight":110},
  {"id":24,"parentId":20,"path":"video","name":"Video","title":"menus.widgets.video","component":"/widgets/video","icon":"ri:vidicon-line","keepAlive":true,"weight":100},
  {"id":25,"parentId":20,"path":"count-to","name":"CountTo","title":"menus.widgets.countTo","component":"/widgets/count-to","icon":"ri:anthropic-line","weight":90},
  {"id":26,"parentId":20,"path":"wang-editor","name":"WangEditor","title":"menus.widgets.wangEditor","component":"/widgets/wang-editor","icon":"ri:t-box-line","keepAlive":true,"weight":80},
  {"id":27,"parentId":20,"path":"watermark","name":"Watermark","title":"menus.widgets.watermark","component":"/widgets/watermark","icon":"ri:water-flash-line","keepAlive":true,"weight":70},
  {"id":28,"parentId":20,"path":"context-menu","name":"ContextMenu","title":"menus.widgets.contextMenu","component":"/widgets/context-menu","icon":"ri:menu-2-line","keepAlive":true,"weight":60},
  {"id":29,"parentId":20,"path":"qrcode","name":"Qrcode","title":"menus.widgets.qrcode","component":"/widgets/qrcode","icon":"ri:qr-code-line","keepAlive":true,"weight":50},
  {"id":30,"parentId":20,"path":"drag","name":"Drag","title":"menus.widgets.drag","component":"/widgets/drag","icon":"ri:drag-move-fill","keepAlive":true,"weight":40},
  {"id":31,"parentId":20,"path":"text-scroll","name":"TextScroll","title":"menus.widgets.textScroll","component":"/widgets/text-scroll","icon":"ri:input-method-line","keepAlive":true,"weight":30},
  {"id":32,"parentId":20,"path":"fireworks","name":"Fireworks","title":"menus.widgets.fireworks","component":"/widgets/fireworks","icon":"ri:magic-line","keepAlive":true,"showTextBadge":"Hot","weight":20},
  {"id":33,"parentId":20,"path":"/outside/iframe/elementui","name":"ElementUI","title":"menus.widgets.elementUI","component":"","icon":"ri:apps-2-line","link":"https://element-plus.org/zh-CN/component/overview.html","isIframe":true,"weight":10},

  {"id":40,"parentId":0,"path":"/examples","name":"Examples","title":"menus.examples.title","component":"/index/index","icon":"ri:sparkling-line","roles":["admin"],"weight":70},
  {"id":41,"parentId":40,"path":"permission","name":"Permission","title":"menus.examples.permission.title","component":"","icon":"ri:fingerprint-line","weight":100},
  {"id":42,"parentId":41,"path":"switch-role","name":"PermissionSwitchRole","title":"menus.examples.permission.switchRole","component":"/examples/permission/switch-role","icon":"ri:contacts-line","keepAlive":true,"weight":100},
  {"id":43,"parentId":41,"path":"button-auth","name":"PermissionButtonAuth","title":"menus.examples.permission.buttonAuth","component":"/examples/permission/button-auth","icon":"ri:mouse-line","keepAlive":true,"authList":[{"title":"新增","authMark":"add"},{"title":"编辑","authMark":"edit"},{"title":"删除","authMark":"delete"},{"title":"导出","authMark":"export"},{"title":"查看","authMark":"view"},{"title":"发布","authMark":"publish"},{"title":"配置","authMark":"config"},{"title":"管理","authMark":"manage"}],"weight":90},
  {"id":44,"parentId":41,"path":"page-visibility","name":"PermissionPageVisibility","title":"menus.examples.permission.pageVisibility","component":"/examples/permission/page-visibility","icon":"ri:user-3-line","keepAlive":true,"roles":["admin","R_SUPER"],"weight":80},
  {"id":45,"parentId":40,"path":"tabs","name":"Tabs","title":"menus.examples.tabs","component":"/examples/tabs","icon":"ri:price-tag-line","weight":90},
  {"id":46,"parentId":40,"path":"tables/basic","name":"TablesBasic","title":"menus.examples.tablesBasic","component":"/examples/tables/basic","icon":"ri:layout-grid-line","keepAlive":true,"weight":80},
  {"id":47,"parentId":40,"path":"tables","name":"Tables","title":"menus.examples.tables","component":"/examples/tables","icon":"ri:table-3","keepAlive":true,"weight":70},
  {"id":48,"parentId":40,"path":"forms","name":"Forms","title":"menus.examples.forms","component":"/examples/forms","icon":"ri:table-view","keepAlive":true,"weight":60},
  {"id":49,"parentId":40,"path":"form/search-bar","name":"SearchBar","title":"menus.examples.searchBar","component":"/examples/forms/search-bar","icon":"ri:table-line","keepAlive":true,"weight":50},
  {"id":50,"parentId":40,"path":"tables/tree","name":"TablesTree","title":"menus.examples.tablesTree","component":"/examples/tables/tree","icon":"ri:layout-2-line","keepAlive":true,"weight":40},
  {"id":51,"parentId":40,"path":"socket-chat","name":"SocketChat","title":"menus.examples.socketChat","component":"/examples/socket-chat","icon":"ri:shake-hands-line","keepAlive":true,"weight":30},

  {"id":60,"parentId":0,"path":"/admin","name":"AdminManage","title":"权限管理","component":"/index/index","icon":"ri:shield-user-line","roles":["admin"],"weight":60},
  {"id":61,"parentId":60,"path":"user","name":"AdminUser","title":"用户管理","component":"/admin/user","icon":"ri:user-line","keepAlive":true,"weight":100},
  {"id":62,"parentId":60,"path":"role","name":"AdminRole","title":"角色管理","component":"/admin/role","icon":"ri:user-settings-line","keepAlive":true,"weight":90},
  {"id":63,"parentId":60,"path":"menu","name":"AdminMenu","title":"菜单管理","component":"/admin/menu","icon":"ri:menu-line","keepAlive":true,"weight":80},
  {"id":64,"parentId":60,"path":"api","name":"AdminApi","title":"接口管理","component":"/admin/api","icon":"ri:terminal-window-line","keepAlive":true,"weight":70},

  {"id":70,"parentId":0,"path":"/system","name":"System","title":"menus.system.title","component":"/index/index","icon":"ri:user-3-line","roles":["admin","R_SUPER","R_ADMIN"],"weight":50},
  {"id":71,"parentId":70,"path":"user","name":"User","title":"menus.system.user","component":"/system/user","icon":"ri:user-line","keepAlive":true,"roles":["admin","R_SUPER","R_ADMIN"],"weight":100},
  {"id":72,"parentId":70,"path":"role","name":"Role","title":"menus.system.role","component":"/system/role","icon":"ri:user-settings-line","keepAlive":true,"roles":["admin","R_SUPER"],"weight":90},
  {"id":73,"parentId":70,"path":"user-center","name":"UserCenter","title":"menus.system.userCenter","component":"/system/user-center","icon":"ri:user-line","keepAlive":true,"isHide":true,"isHideTab":true,"weight":80},
  {"id":74,"parentId":70,"path":"menu","name":"Menus","title":"menus.system.menu","component":"/system/menu","icon":"ri:menu-line","keepAlive":true,"roles":["admin","R_SUPER"],"authList":[{"title":"新增","authMark":"add"},{"title":"编辑","authMark":"edit"},{"title":"删除","authMark":"delete"}],"weight":70},
  {"id":75,"parentId":70,"path":"nested","name":"Nested","title":"menus.system.nested","component":"","icon":"ri:menu-unfold-3-line","keepAlive":true,"weight":60},
  {"id":76,"parentId":75,"path":"menu1","name":"NestedMenu1","title":"menus.system.menu1","component":"/system/nested/menu1","icon":"ri:align-justify","keepAlive":true,"weight":100},
  {"id":77,"parentId":75,"path":"menu2","name":"NestedMenu2","title":"menus.system.menu2","component":"","icon":"ri:align-justify","keepAlive":true,"weight":90},
  {"id":78,"parentId":77,"path":"menu2-1","name":"NestedMenu2-1","title":"menus.system.menu21","component":"/system/nested/menu2","icon":"ri:align-justify","keepAlive":true,"weight":100},
  {"id":79,"parentId":75,"path":"menu3","name":"NestedMenu3","title":"menus.system.menu3","component":"","icon":"ri:align-justify","keepAlive":true,"weight":80},
  {"id":80,"parentId":79,"path":"menu3-1","name":"NestedMenu3-1","title":"menus.system.menu31","component":"/system/nested/menu3","keepAlive":true,"weight":100},
  {"id":81,"parentId":79,"path":"menu3-2","name":"NestedMenu3-2","title":"menus.system.menu32","component":"","keepAlive":true,"weight":90},
  {"id":82,"parentId":81,"path":"menu3-2-1","name":"NestedMenu3-2-1","title":"menus.system.menu321","component":"/system/nested/menu3/menu3-2","keepAlive":true,"weight":100},

  {"id":90,"parentId":0,"path":"/article","name":"Article","title":"menus.article.title","component":"/index/index","icon":"ri:book-2-line","roles":["admin","R_SUPER","R_ADMIN"],"weight":40},
  {"id":91,"parentId":90,"path":"article-list","name":"ArticleList","title":"menus.article.articleList","component":"/article/list","icon":"ri:article-line","keepAlive":true,"authList":[{"title":"新增","authMark":"add"},{"title":"编辑","authMark":"edit"}],"weight":100},
  {"id":92,"parentId":90,"path":"detail/:id","name":"ArticleDetail","title":"menus.article.articleDetail","component":"/article/detail","keepAlive":true,"isHide":true,"activePath":"/article/article-list","weight":90},
  {"id":93,"parentId":90,"path":"comment","name":"ArticleComment","title":"menus.article.comment","component":"/article/comment","icon":"ri:mail-line","keepAlive":true,"weight":80},
  {"id":94,"parentId":90,"path":"publish","name":"ArticlePublish","title":"menus.article.articlePublish","component":"/article/publish","icon":"ri:telegram-2-line","keepAlive":true,"authList":[{"title":"发布","authMark":"add"}],"weight":70},

  {"id":100,"parentId":0,"path":"/result","name":"Result","title":"menus.result.title","component":"/index/index","icon":"ri:checkbox-circle-line","roles":["admin"],"weight":30},
  {"id":101,"parentId":100,"path":"success","name":"ResultSuccess","title":"menus.result.success","component":"/result/success","icon":"ri:checkbox-circle-line","keepAlive":true,"weight":100},
  {"id":102,"parentId":100,"path":"fail","name":"ResultFail","title":"menus.result.fail","component":"/result/fail","icon":"ri:close-circle-line","keepAlive":true,"weight":90},

  {"id":110,"parentId":0,"path":"/exception","name":"Exception","title":"menus.exception.title","component":"/index/index","icon":"ri:error-warning-line","roles":["admin"],"weight":20},
  {"id":111,"parentId":110,"path":"403","name":"Exception403","title":"menus.exception.forbidden","component":"/exception/403","keepAlive":true,"isHideTab":true,"isFullPage":true,"weight":100},
  {"id":112,"parentId":110,"path":"404","name":"Exception404","title":"menus.exception.notFound","component":"/exception/404","keepAlive":true,"isHideTab":true,"isFullPage":true,"weight":90},
  {"id":113,"parentId":110,"path":"500","name":"Exception500","title":"menus.exception.serverError","component":"/exception/500","keepAlive":true,"isHideTab":true,"isFullPage":true,"weight":80},

  {"id":120,"parentId":0,"path":"/safeguard","name":"Safeguard","title":"menus.safeguard.title","component":"/index/index","icon":"ri:shield-check-line","roles":["admin"],"weight":10},
  {"id":121,"parentId":120,"path":"server","name":"SafeguardServer","title":"menus.safeguard.server","component":"/safeguard/server","icon":"ri:hard-drive-3-line","keepAlive":true,"weight":100},

  {"id":130,"parentId":0,"path":"/change/log","name":"ChangeLog","title":"menus.plan.log","component":"/change/log","icon":"ri:gamepad-line","showTextBadge":"v3.0.2","weight":1}
]`
