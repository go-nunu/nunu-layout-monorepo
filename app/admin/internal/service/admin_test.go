package service

import (
	"reflect"
	"testing"

	"gorm.io/gorm"
	"nunu-layout-monorepo/model"
)

func TestNormalizeApiGroup(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "single level", input: "菜单管理", want: "菜单管理"},
		{name: "nested levels", input: " 权限管理 / 角色 / 查询 ", want: "权限管理/角色/查询"},
		{name: "empty", input: "", wantErr: true},
		{name: "empty segment", input: "权限管理//角色", wantErr: true},
		{name: "trailing slash", input: "权限管理/", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeApiGroup(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeApiGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("normalizeApiGroup() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUniqueUintIDs(t *testing.T) {
	got := uniqueUintIDs([]uint{3, 0, 2, 3, 2, 5})
	want := []uint{3, 2, 5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("uniqueUintIDs() = %v, want %v", got, want)
	}
}

func TestNormalizeApiInput(t *testing.T) {
	tests := []struct {
		name    string
		group   string
		apiName string
		path    string
		method  string
		wantErr bool
	}{
		{name: "valid comma path", group: "权限 / 查询", apiName: "查询", path: "/v1/items,a", method: "get"},
		{name: "blank name", group: "权限", apiName: " ", path: "/v1/items", method: "GET", wantErr: true},
		{name: "relative path", group: "权限", apiName: "查询", path: "v1/items", method: "GET", wantErr: true},
		{name: "unknown method", group: "权限", apiName: "查询", path: "/v1/items", method: "TRACE", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api, err := normalizeApiInput(tt.group, tt.apiName, tt.path, tt.method)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeApiInput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && api.Method != "GET" {
				t.Fatalf("normalizeApiInput() method = %q", api.Method)
			}
		})
	}
}

func TestBuildMenuPathIndex(t *testing.T) {
	menus := []model.Menu{
		{Model: gorm.Model{ID: 1}, ParentID: 0, Path: "/admin"},
		{Model: gorm.Model{ID: 2}, ParentID: 1, Path: "roles"},
		{Model: gorm.Model{ID: 3}, ParentID: 2, Path: "detail"},
	}
	paths, err := buildMenuPathIndex(menus)
	if err != nil {
		t.Fatalf("buildMenuPathIndex() error = %v", err)
	}
	if paths[3] != "/admin/roles/detail" {
		t.Fatalf("buildMenuPathIndex() path = %q", paths[3])
	}
}

func TestBuildMenuPathIndexRejectsInvalidHierarchy(t *testing.T) {
	tests := []struct {
		name  string
		menus []model.Menu
	}{
		{
			name: "cycle",
			menus: []model.Menu{
				{Model: gorm.Model{ID: 1}, ParentID: 2, Path: "one"},
				{Model: gorm.Model{ID: 2}, ParentID: 1, Path: "two"},
			},
		},
		{
			name: "missing parent",
			menus: []model.Menu{
				{Model: gorm.Model{ID: 1}, ParentID: 99, Path: "one"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildMenuPathIndex(tt.menus); err == nil {
				t.Fatal("buildMenuPathIndex() expected error")
			}
		})
	}
}

func TestValidateMenuCollectionRejectsDuplicateIdentity(t *testing.T) {
	base := model.Menu{Model: gorm.Model{ID: 1}, Title: "菜单一", Name: "MenuOne", Path: "/one"}
	tests := []struct {
		name string
		menu model.Menu
	}{
		{name: "duplicate name", menu: model.Menu{Model: gorm.Model{ID: 2}, Title: "菜单二", Name: "MenuOne", Path: "/two"}},
		{name: "duplicate full path", menu: model.Menu{Model: gorm.Model{ID: 2}, Title: "菜单二", Name: "MenuTwo", Path: "/one"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := validateMenuCollection([]model.Menu{base, tt.menu}); err == nil {
				t.Fatal("validateMenuCollection() expected error")
			}
		})
	}
}

func TestIncludeMenuDescendants(t *testing.T) {
	menus := []model.Menu{
		{Model: gorm.Model{ID: 1}},
		{Model: gorm.Model{ID: 2}, ParentID: 1},
		{Model: gorm.Model{ID: 3}, ParentID: 2},
		{Model: gorm.Model{ID: 4}},
	}
	got := includeMenuDescendants(menus, map[uint]struct{}{1: {}})
	if _, exists := got[3]; !exists {
		t.Fatal("includeMenuDescendants() did not include nested child")
	}
	if _, exists := got[4]; exists {
		t.Fatal("includeMenuDescendants() included unrelated menu")
	}
}
