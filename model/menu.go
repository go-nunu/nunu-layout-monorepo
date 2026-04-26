package model

import "gorm.io/gorm"

type MenuAuth struct {
	Title    string `json:"title"`
	AuthMark string `json:"authMark"`
}

type Menu struct {
	gorm.Model
	ParentID      uint       `json:"parentId,omitempty" gorm:"column:parent_id;index;comment:父级菜单的id，使用整数表示"`         // 父级菜单的id，使用整数表示
	Path          string     `json:"path" gorm:"column:path;type:varchar(255);comment:地址"`                            // 地址
	Title         string     `json:"title" gorm:"column:title;type:varchar(100);comment:标题，使用字符串表示"`                  // 标题，使用字符串表示
	Name          string     `json:"name,omitempty" gorm:"column:name;type:varchar(100);comment:同路由中的name，用于保活"`      // 同路由中的name，用于保活
	Component     string     `json:"component,omitempty" gorm:"column:component;type:varchar(255);comment:绑定的组件"`     // 绑定的组件，默认类型：Iframe、RouteView、ComponentError
	Locale        string     `json:"locale,omitempty" gorm:"column:locale;type:varchar(100);comment:本地化标识"`           // 本地化标识
	Icon          string     `json:"icon,omitempty" gorm:"column:icon;type:varchar(100);comment:图标，使用字符串表示"`          // 图标，使用字符串表示
	Redirect      string     `json:"redirect,omitempty" gorm:"column:redirect;type:varchar(255);comment:重定向地址"`       // 重定向地址
	URL           string     `json:"url,omitempty" gorm:"column:url;type:varchar(255);comment:iframe模式下的跳转url"`       // iframe模式下的跳转url，不能与path重复
	Link          string     `json:"link,omitempty" gorm:"column:link;type:varchar(255);comment:外部链接"`                // 外部链接
	Target        string     `json:"target,omitempty" gorm:"column:target;type:varchar(20);comment:全连接跳转模式"`          // 全连接跳转模式：'_blank'、'_self'、'_parent'
	ActivePath    string     `json:"activePath,omitempty" gorm:"column:active_path;type:varchar(255);comment:激活菜单路径"` // 激活菜单路径
	ShowTextBadge string     `json:"showTextBadge,omitempty" gorm:"column:show_text_badge;type:varchar(50);comment:文本徽章"`
	Weight        int        `json:"weight" gorm:"column:weight;type:int;default:0;comment:排序权重"`
	IsEnable      bool       `json:"isEnable" gorm:"column:is_enable;comment:是否启用"`
	IsMenu        bool       `json:"isMenu" gorm:"column:is_menu;comment:是否菜单"`
	KeepAlive     bool       `json:"keepAlive,omitempty" gorm:"column:keep_alive;default:false;comment:是否保活"`      // 是否保活
	HideInMenu    bool       `json:"hideInMenu,omitempty" gorm:"column:hide_in_menu;default:false;comment:是否隐藏菜单"` // 兼容旧版隐藏菜单字段
	IsHide        bool       `json:"isHide,omitempty" gorm:"column:is_hide;default:false;comment:是否隐藏菜单"`
	IsHideTab     bool       `json:"isHideTab,omitempty" gorm:"column:is_hide_tab;default:false;comment:是否隐藏标签页"`
	IsIframe      bool       `json:"isIframe,omitempty" gorm:"column:is_iframe;default:false;comment:是否内嵌"`
	ShowBadge     bool       `json:"showBadge,omitempty" gorm:"column:show_badge;default:false;comment:是否显示徽章"`
	FixedTab      bool       `json:"fixedTab,omitempty" gorm:"column:fixed_tab;default:false;comment:是否固定标签页"`
	IsFullPage    bool       `json:"isFullPage,omitempty" gorm:"column:is_full_page;default:false;comment:是否全屏页面"`
	Roles         []string   `json:"roles,omitempty" gorm:"column:roles;serializer:json;type:text;comment:角色权限"`
	AuthList      []MenuAuth `json:"authList,omitempty" gorm:"column:auth_list;serializer:json;type:text;comment:按钮权限列表"`
}

func (m *Menu) TableName() string {
	return "menu"
}
