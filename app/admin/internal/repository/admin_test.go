package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	appmodel "nunu-layout-monorepo/model"
)

func newPermissionTestRepository(t *testing.T) *adminRepository {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		t.Fatalf("create casbin adapter: %v", err)
	}
	casbinModel, err := casbinmodel.NewModelFromString(`
[request_definition]
r = sub, obj, act
[policy_definition]
p = sub, obj, act
[role_definition]
g = _, _
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
	if err != nil {
		t.Fatalf("create casbin model: %v", err)
	}
	enforcer, err := casbin.NewSyncedEnforcer(casbinModel, adapter)
	if err != nil {
		t.Fatalf("create casbin enforcer: %v", err)
	}
	enforcer.EnableAutoSave(true)
	return &adminRepository{Repository: &Repository{db: db, e: enforcer}}
}

func TestUpdateRolePermissionReplacesOnlyTargetRole(t *testing.T) {
	repository := newPermissionTestRepository(t)
	if _, err := repository.e.AddPermissionForUser("role-a", "api:/old", "GET"); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.e.AddPermissionForUser("role-b", "api:/keep", "GET"); err != nil {
		t.Fatal(err)
	}
	want := appmodel.Permission{Resource: "api:/items,a", Action: "GET"}
	if err := repository.UpdateRolePermission(context.Background(), "role-a", []appmodel.Permission{want}); err != nil {
		t.Fatalf("UpdateRolePermission() error = %v", err)
	}
	permissions, err := repository.e.GetPermissionsForUser("role-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(permissions) != 1 || permissions[0][1] != want.Resource || permissions[0][2] != want.Action {
		t.Fatalf("role-a permissions = %#v", permissions)
	}
	kept, err := repository.e.HasPermissionForUser("role-b", "api:/keep", "GET")
	if err != nil || !kept {
		t.Fatalf("role-b permission lost, kept=%v err=%v", kept, err)
	}
}

func TestPermissionReferenceLifecycle(t *testing.T) {
	repository := newPermissionTestRepository(t)
	oldPermission := appmodel.Permission{Resource: "menu:/old", Action: "read"}
	newPermission := appmodel.Permission{Resource: "menu:/new", Action: "read"}
	if _, err := repository.e.AddPermissionForUser("role-a", oldPermission.Resource, oldPermission.Action); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := repository.ReplacePermissionReferences(ctx, map[appmodel.Permission]appmodel.Permission{
		oldPermission: newPermission,
	}); err != nil {
		t.Fatalf("ReplacePermissionReferences() error = %v", err)
	}
	if err := repository.ReloadPolicy(); err != nil {
		t.Fatal(err)
	}
	hasNew, _ := repository.e.HasPermissionForUser("role-a", newPermission.Resource, newPermission.Action)
	hasOld, _ := repository.e.HasPermissionForUser("role-a", oldPermission.Resource, oldPermission.Action)
	if !hasNew || hasOld {
		t.Fatalf("replacement result old=%v new=%v", hasOld, hasNew)
	}
	if err := repository.DeletePermissionReferences(ctx, []appmodel.Permission{newPermission}); err != nil {
		t.Fatalf("DeletePermissionReferences() error = %v", err)
	}
	if err := repository.ReloadPolicy(); err != nil {
		t.Fatal(err)
	}
	hasNew, _ = repository.e.HasPermissionForUser("role-a", newPermission.Resource, newPermission.Action)
	if hasNew {
		t.Fatal("permission was not deleted")
	}
}

func TestApiPermissionIsUniqueAndReusableAfterDelete(t *testing.T) {
	repository := newPermissionTestRepository(t)
	if err := repository.db.AutoMigrate(&appmodel.Api{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	first := &appmodel.Api{Group: "test", Name: "first", Path: "/v1/items", Method: "GET"}
	if err := repository.ApiCreate(context.Background(), first); err != nil {
		t.Fatalf("ApiCreate() error = %v", err)
	}
	duplicate := &appmodel.Api{Group: "test", Name: "duplicate", Path: "/v1/items", Method: "GET"}
	if err := repository.ApiCreate(context.Background(), duplicate); err == nil {
		t.Fatal("ApiCreate() accepted duplicate path and method")
	}
	if err := repository.ApiDelete(context.Background(), first.ID); err != nil {
		t.Fatalf("ApiDelete() error = %v", err)
	}
	replacement := &appmodel.Api{Group: "test", Name: "replacement", Path: "/v1/items", Method: "GET"}
	if err := repository.ApiCreate(context.Background(), replacement); err != nil {
		t.Fatalf("ApiCreate() after delete error = %v", err)
	}
}
