package model

import "gorm.io/gorm"

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
	Group  string `gorm:"type:varchar(100);not null;comment:'API分组'"`
	Name   string `gorm:"type:varchar(100);not null;comment:'API名称'"`
	Path   string `gorm:"type:varchar(255);not null;comment:'API路径'"`
	Method string `gorm:"type:varchar(20);not null;comment:'HTTP方法'"`
}

func (m *Api) TableName() string {
	return "api"
}
