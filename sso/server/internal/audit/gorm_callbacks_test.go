package audit

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestPopulateUpdateProvenance(t *testing.T) {
	ctx := WithOperatorID(WithRequestID(context.Background(), "request-1"), 42)
	values := map[string]interface{}{"name": "new name"}
	tx := &gorm.DB{Statement: &gorm.Statement{Context: ctx, Table: "sys_user", Dest: values}}

	populateUpdateProvenance(tx)

	if got := values["updated_by"]; got != uint64(42) {
		t.Fatalf("updated_by = %#v, want 42", got)
	}
	if got := values["request_id"]; got != "request-1" {
		t.Fatalf("request_id = %#v, want request-1", got)
	}
	if _, exists := values["created_by"]; exists {
		t.Fatal("update must not set created_by")
	}
}

func TestPopulateCreateProvenance(t *testing.T) {
	ctx := WithOperatorID(WithRequestID(context.Background(), "request-2"), 7)
	values := map[string]interface{}{"name": "new name"}
	tx := &gorm.DB{Statement: &gorm.Statement{Context: ctx, Table: "sys_role", Dest: values}}

	populateCreateProvenance(tx)

	if got := values["created_by"]; got != uint64(7) {
		t.Fatalf("created_by = %#v, want 7", got)
	}
	if got := values["updated_by"]; got != uint64(7) {
		t.Fatalf("updated_by = %#v, want 7", got)
	}
	if got := values["request_id"]; got != "request-2" {
		t.Fatalf("request_id = %#v, want request-2", got)
	}
}
