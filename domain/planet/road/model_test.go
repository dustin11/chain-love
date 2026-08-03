package road

import "testing"

// 验证道路领域表名保持稳定。
func TestDocumentTableName(t *testing.T) {
	if tableName := (Document{}).TableName(); tableName != "pla_road_document" {
		t.Fatalf("unexpected table name %q", tableName)
	}
}
