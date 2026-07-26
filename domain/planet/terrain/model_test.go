package terrain

import "testing"

// TestDocumentTableName 锁定 planet 表前缀。
func TestDocumentTableName(t *testing.T) {
	if tableName := (Document{}).TableName(); tableName != "pla_terrain_document" {
		t.Fatalf("unexpected table name: %s", tableName)
	}
	if len(Tables()) != 1 {
		t.Fatalf("unexpected terrain table count: %d", len(Tables()))
	}
}
