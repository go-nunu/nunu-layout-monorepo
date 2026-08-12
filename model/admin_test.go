package model

import "testing"

func TestParsePermissionKeyUsesLastSeparator(t *testing.T) {
	permission, ok := ParsePermissionKey("api:/v1/items,a,GET")
	if !ok {
		t.Fatal("ParsePermissionKey() rejected valid key")
	}
	if permission.Resource != "api:/v1/items,a" || permission.Action != "GET" {
		t.Fatalf("ParsePermissionKey() = %#v", permission)
	}
}

func TestParsePermissionKeyRejectsMalformedValue(t *testing.T) {
	for _, key := range []string{"", "api:/v1/items", ",GET", "api:/v1/items,"} {
		if _, ok := ParsePermissionKey(key); ok {
			t.Fatalf("ParsePermissionKey(%q) expected rejection", key)
		}
	}
}
