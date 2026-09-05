// Package canal 解析 Canal Server(flatMessage=false) 经 Kafka 投递的 protobuf 消息。
//
// FlatMessage 是归一化后的"单条行级变更事件", 供 mapper 消费:
// 字段命名沿用 canal flat JSON 协议(FlatMessage.java), 便于上层语义保持一致;
// 但 binlog 位点(binlogFileName/binlogPosition)在 flat JSON 中并不存在,
// 切换 protobuf 后由 CanalEntry.Entry.header(logfileName/logfileOffset)真实填充。
package canal

// FlatMessage 表示一次 ROWDATA 事件(一条 binlog 记录, Data/Old 可含多行)。
// isNull 的列值为 nil, 非 NULL 值统一为字符串。
type FlatMessage struct {
	Database        string               `json:"database"`
	Table           string               `json:"table"`
	Type            string               `json:"type"` // INSERT / UPDATE / DELETE
	IsDdl           bool                 `json:"isDdl"`
	ES              int64                `json:"es"` // 事件时间(Unix 毫秒, 即 header.executeTime)
	SQL             string               `json:"sql"`
	SQLType         map[string]int32     `json:"sqlType"`
	MySQLType       map[string]string    `json:"mysqlType"`
	Data            []map[string]*string `json:"data"`
	Old             []map[string]*string `json:"old"`
	GTID            string               `json:"gtid"`
	BinlogFileName  string               `json:"binlogFileName"`
	BinlogPosition  int64                `json:"binlogPosition"`
}

// IsDML 是否为需要入库的数据变更事件。
func (m *FlatMessage) IsDML() bool {
	switch m.Type {
	case "INSERT", "UPDATE", "DELETE":
		return true
	}
	return false
}

// IsDDLEvent 是否为 DDL 事件(Canal 的 isDdl=true 或 type 为建表类)。
func (m *FlatMessage) IsDDLEvent() bool {
	if m.IsDdl {
		return true
	}
	switch m.Type {
	case "CREATE", "ALTER", "DROP", "TRUNCATE", "RENAME",
		"ERASE", "CINDEX", "DINDEX", "QUERY":
		return true
	}
	return false
}

// DecodeValue 解析一条 Kafka 消息体(protobuf: Packet -> Messages -> Entry[])。
// 返回值为消息内所有 ROWDATA 事件(每个 Entry 一个), 事务开始/结束、心跳、
// DDL 等非行级事件会被过滤(返回 nil 或不在结果中)。
func DecodeValue(value []byte) ([]*FlatMessage, error) {
	return decodeMessage(value)
}
