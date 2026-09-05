package canal

import "testing"

// 构造一条 protobuf 编码的 Kafka 消息, 内含一个 ROWDATA UPDATE 事件。
// 位点/事件时间/gtid 位于 Entry.header(logfileName/logfileOffset/executeTime)。
func buildSampleUpdateMsg() []byte {
	beforeCols := [][]byte{
		tColumn("id", false, "1"),
		tColumn("account", false, "admin"),
		tColumn("password", false, "hash-old"),
		tColumn("name", false, "旧名"),
		tColumn("phone", true, ""),
		tColumn("created_by", false, "5"),
		tColumn("updated_by", false, "5"),
		tColumn("request_id", false, "req-000"),
		tColumn("create_time", false, "2026-09-01 10:00:00"),
		tColumn("update_time", false, "2026-09-01 10:00:00"),
	}
	afterCols := [][]byte{
		tColumn("id", false, "1"),
		tColumn("account", false, "admin"),
		tColumn("password", false, "hash-new"),
		tColumn("name", false, "管理员"),
		tColumn("phone", true, ""),
		tColumn("created_by", false, "5"),
		tColumn("updated_by", false, "9"),
		tColumn("request_id", false, "req-123"),
		tColumn("create_time", false, "2026-09-01 10:00:00"),
		tColumn("update_time", false, "2026-09-05 10:00:00"),
	}
	rowData := tRowData(beforeCols, afterCols)
	rowChange := tRowChange(eventUpdate, false, rowData)
	entry := tEntry("sso", "sys_user", "mysql-bin.000001", 500, 1757059200000,
		"abcdef:1-10", entryTypeRowData, rowChange)
	return tPacket(packetTypeMessages, tMessages(12, entry))
}

func TestDecodeValueUpdate(t *testing.T) {
	msgs, err := DecodeValue(buildSampleUpdateMsg())
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	m := msgs[0]
	if m.Database != "sso" || m.Table != "sys_user" || m.Type != "UPDATE" {
		t.Fatalf("基础字段解析错误: %+v", m)
	}
	if m.BinlogFileName != "mysql-bin.000001" || m.BinlogPosition != 500 {
		t.Fatalf("binlog 位点解析错误: %s:%d", m.BinlogFileName, m.BinlogPosition)
	}
	if m.GTID != "abcdef:1-10" {
		t.Fatalf("gtid 解析错误: %s", m.GTID)
	}
	if m.ES != 1757059200000 {
		t.Fatalf("es 解析错误: %d", m.ES)
	}
	if len(m.Data) != 1 || len(m.Old) != 1 {
		t.Fatalf("data/old 行数错误: %d/%d", len(m.Data), len(m.Old))
	}
	if !m.IsDML() || m.IsDDLEvent() {
		t.Fatalf("DML/DDL 判断错误")
	}
	if m.Data[0]["updated_by"] == nil || *m.Data[0]["updated_by"] != "9" {
		t.Fatalf("列值解析错误: %v", m.Data[0]["updated_by"])
	}
	if m.Old[0]["name"] == nil || *m.Old[0]["name"] != "旧名" {
		t.Fatalf("before 列解析错误: %v", m.Old[0]["name"])
	}
	if m.Data[0]["phone"] != nil {
		t.Fatalf("isNull 列应解析为 nil")
	}
	if m.MySQLType["id"] != "" || m.SQLType["id"] != 0 {
		t.Fatalf("类型信息未按预期填充(测试未设置 mysqlType/sqlType): %+v", m.MySQLType)
	}
}

func TestDecodeValueBatchSkipsNonRowData(t *testing.T) {
	// 一个 batch 内混入: 事务开始(entryType=1)、一条 INSERT、事务结束(entryType=3)、一条 DELETE。
	insertRC := tRowChange(eventInsert, false,
		tRowData(nil, [][]byte{tColumn("user_id", false, "3"), tColumn("role_id", false, "7")}))
	insertEntry := tEntry("sso", "sys_user_role", "mysql-bin.000002", 10, 1757059200100,
		"", entryTypeRowData, insertRC)

	deleteRC := tRowChange(eventDelete, false,
		tRowData([][]byte{tColumn("id", false, "8")}, nil))
	deleteEntry := tEntry("sso", "sys_role", "mysql-bin.000003", 20, 1757059200200,
		"", entryTypeRowData, deleteRC)

	begin := tEntry("sso", "", "", 5, 0, "", 1 /*TRANSACTIONBEGIN*/, nil)
	end := tEntry("sso", "", "", 22, 0, "", 3 /*TRANSACTIONEND*/, nil)

	body := tMessages(99, begin, insertEntry, end, deleteEntry)
	msgs, err := DecodeValue(tPacket(packetTypeMessages, body))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("len = %d, want 2(事务头尾应被过滤)", len(msgs))
	}
	if msgs[0].Type != "INSERT" || msgs[0].Table != "sys_user_role" {
		t.Fatalf("首条应为 INSERT sys_user_role: %+v", msgs[0])
	}
	if msgs[1].Type != "DELETE" || msgs[1].Table != "sys_role" {
		t.Fatalf("次条应为 DELETE sys_role: %+v", msgs[1])
	}
	// DELETE 行数据放入 data(供 mapper 提取变更前快照)
	if msgs[1].Data[0]["id"] == nil || *msgs[1].Data[0]["id"] != "8" {
		t.Fatalf("DELETE before 列应放入 data: %v", msgs[1].Data)
	}
}

func TestDecodeValueNonMessages(t *testing.T) {
	// type 非 MESSAGES 的包不应产出消息也不报错
	msgs, err := DecodeValue(tPacket(3 /*ACK*/, []byte("x")))
	if err != nil {
		t.Fatalf("不应报错: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("len = %d, want 0", len(msgs))
	}
}

func TestDecodeValueEmptyAndTruncated(t *testing.T) {
	if msgs, err := DecodeValue(nil); err != nil || len(msgs) != 0 {
		t.Fatalf("空消息应返回空: %v %v", msgs, err)
	}
	if _, err := DecodeValue([]byte{0xff, 0xff, 0xff, 0xff}); err == nil {
		t.Fatalf("截断消息应报错")
	}
	if _, err := DecodeValue(tPacket(packetTypeMessages, []byte{0x10, 0xff})); err == nil {
		t.Fatalf("body 截断应报错")
	}
}

func TestIsDDLEvent(t *testing.T) {
	for _, typ := range []string{"CREATE", "ALTER", "DROP", "TRUNCATE", "RENAME"} {
		m := &FlatMessage{Type: typ}
		if !m.IsDDLEvent() {
			t.Fatalf("%s 应判定为 DDL", typ)
		}
		if m.IsDML() {
			t.Fatalf("%s 不应判定为 DML", typ)
		}
	}
}
