package model

import (
	"strings"

	"gorm.io/gorm"
)

const (
	AdminRole          = "admin"
	AdminUserID        = "1"
	MenuResourcePrefix = "menu:"
	ApiResourcePrefix  = "api:"
	PermSep            = ","
)

type AdminUser struct {
	gorm.Model
	Username string `gorm:"type:varchar(50);not null;uniqueIndex;comment:'用户名'"`
	Nickname string `gorm:"type:varchar(50);not null;comment:'昵称'"`
	Password string `gorm:"type:varchar(255);not null;comment:'密码'"`
	Email    string `gorm:"type:varchar(100);not null;comment:'电子邮件'"`
	Phone    string `gorm:"type:varchar(20);not null;comment:'手机号'"`
}

func (m *AdminUser) TableName() string {
	return "admin_users"
}

type Role struct {
	gorm.Model
	Name string `json:"name" gorm:"column:name;type:varchar(100);uniqueIndex;comment:角色名"`
	Sid  string `json:"sid" gorm:"column:sid;type:varchar(100);uniqueIndex;comment:角色标识"`
}

func (m *Role) TableName() string {
	return "roles"
}

type Api struct {
	gorm.Model
	Group   string `gorm:"type:varchar(255);not null;comment:'API分类路径'"`
	Name    string `gorm:"type:varchar(100);not null;comment:'API名称'"`
	Path    string `gorm:"type:varchar(255);not null;uniqueIndex:idx_api_path_method;comment:'API路径'"`
	Method  string `gorm:"type:varchar(20);not null;uniqueIndex:idx_api_path_method;comment:'HTTP方法'"`
	MenuIDs []uint `gorm:"column:menu_ids;serializer:json;type:text;comment:'关联菜单ID列表'"`
}

func (m *Api) TableName() string {
	return "api"
}

type Permission struct {
	Resource string
	Action   string
}

func (p Permission) Key() string {
	return p.Resource + PermSep + p.Action
}

func ParsePermissionKey(key string) (Permission, bool) {
	separator := strings.LastIndex(key, PermSep)
	if separator <= 0 || separator == len(key)-1 {
		return Permission{}, false
	}
	permission := Permission{
		Resource: strings.TrimSpace(key[:separator]),
		Action:   strings.TrimSpace(key[separator+len(PermSep):]),
	}
	if permission.Resource == "" || permission.Action == "" {
		return Permission{}, false
	}
	return permission, true
}
