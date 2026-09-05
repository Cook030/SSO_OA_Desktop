// Package canal 解析 Canal Server(flatMessage=false) 投递到 Kafka 的 protobuf 消息。
//
// 消息封装链路(canal 1.1.7, 与 CanalMessageDeserializer 一致):
//
//	CanalPacket.Packet  (type=MESSAGES, body=CanalPacket.Messages 的序列化字节)
//	  └── CanalPacket.Messages  (batch_id + messages[]: 每个元素是 CanalEntry.Entry 的序列化字节)
//	        └── CanalEntry.Entry  (entryType=ROWDATA 时, storeValue 是 RowChange 的序列化字节)
//	              └── CanalEntry.RowChange  (eventType + rowDatas[], 每行 before/afterColumns)
//
// binlog 位点(logfileName/logfileOffset)、事件时间(executeTime)、gtid 均位于
// CanalEntry.Entry.header, 这正是 flatMessage=true 的 JSON 所不提供的字段。
package canal

import (
	"errors"
	"fmt"
)

// ---- CanalProtocol.proto / EntryProtocol.proto 枚举常量(canal-1.1.7, proto3) ----
const (
	packetTypeMessages    = 7 // CanalPacket.PacketType.MESSAGES
	compressionProto2None = 0 // Compression.COMPRESSIONCOMPATIBLEPROTO2(旧协议未压缩)
	compressionNone       = 1 // Compression.NONE
	entryTypeRowData      = 2 // CanalEntry.EntryType.ROWDATA
)

// ---- CanalEntry.EventType 值(proto3) ----
const (
	eventInsert = 1
	eventUpdate = 2
	eventDelete = 3
)

var (
	errTruncated   = errors.New("protobuf 字节流截断")
	errBadVarint   = errors.New("protobuf varint 编码非法")
	errWireType    = errors.New("不支持的 protobuf wire type")
	errCompression = errors.New("不支持的报文压缩方式")
)

// pbReader 只读 protobuf wire 流解析器, 未知字段一律跳过, 保证协议向后兼容。
type pbReader struct {
	b   []byte
	off int
}

func (r *pbReader) done() bool { return r.off >= len(r.b) }

func (r *pbReader) varint() (uint64, error) {
	var x uint64
	var s uint
	for i := 0; i < 10; i++ {
		if r.off >= len(r.b) {
			return 0, errTruncated
		}
		b := r.b[r.off]
		r.off++
		if b < 0x80 {
			if i == 9 && b > 1 {
				return 0, errBadVarint
			}
			return x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, errBadVarint
}

func (r *pbReader) bytes() ([]byte, error) {
	n, err := r.varint()
	if err != nil {
		return nil, err
	}
	if n > uint64(len(r.b)-r.off) {
		return nil, errTruncated
	}
	out := r.b[r.off : r.off+int(n)]
	r.off += int(n)
	return out, nil
}

func (r *pbReader) skip(wire int) error {
	switch wire {
	case 0: // varint
		_, err := r.varint()
		return err
	case 1: // 64-bit
		if len(r.b)-r.off < 8 {
			return errTruncated
		}
		r.off += 8
		return nil
	case 2: // length-delimited
		_, err := r.bytes()
		return err
	case 5: // 32-bit
		if len(r.b)-r.off < 4 {
			return errTruncated
		}
		r.off += 4
		return nil
	default:
		return fmt.Errorf("%w: %d", errWireType, wire)
	}
}

// nextField 读取下一个字段的 field number 与 wire type。
func (r *pbReader) nextField() (field int, wire int, err error) {
	tag, err := r.varint()
	if err != nil {
		return 0, 0, err
	}
	return int(tag >> 3), int(tag & 7), nil
}

// ---- 协议解码: 每个 message 对应一个函数, 只解需要的字段, 其余跳过 ----

// decodePacket 解析 CanalProtocol.Packet: type(3)/compression(4)/body(5)。
func decodePacket(b []byte) (msgType, compression int, body []byte, err error) {
	r := &pbReader{b: b}
	for !r.done() {
		field, wire, err := r.nextField()
		if err != nil {
			return 0, 0, nil, err
		}
		switch field {
		case 3:
			if wire != 0 {
				return 0, 0, nil, fmt.Errorf("Packet.type 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return 0, 0, nil, err
			}
			msgType = int(v)
		case 4:
			if wire != 0 {
				return 0, 0, nil, fmt.Errorf("Packet.compression 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return 0, 0, nil, err
			}
			compression = int(v)
		case 5:
			body, err = r.bytes()
			if err != nil {
				return 0, 0, nil, err
			}
		default:
			if err := r.skip(wire); err != nil {
				return 0, 0, nil, err
			}
		}
	}
	return msgType, compression, body, nil
}

// decodeMessages 解析 CanalProtocol.Messages: batch_id(1)/messages(2, repeated bytes)。
func decodeMessages(b []byte) (batchID int64, entries [][]byte, err error) {
	r := &pbReader{b: b}
	for !r.done() {
		field, wire, err := r.nextField()
		if err != nil {
			return 0, nil, err
		}
		switch field {
		case 1:
			if wire != 0 {
				return 0, nil, fmt.Errorf("Messages.batch_id 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return 0, nil, err
			}
			batchID = int64(v)
		case 2:
			if wire != 2 {
				return 0, nil, fmt.Errorf("Messages.messages 期望 bytes, 实际 wire=%d", wire)
			}
			eb, err := r.bytes()
			if err != nil {
				return 0, nil, err
			}
			entries = append(entries, eb)
		default:
			if err := r.skip(wire); err != nil {
				return 0, nil, err
			}
		}
	}
	return batchID, entries, nil
}

// decodeEntry 解析 CanalEntry.Entry 的 header 关键字段, 并取出 entryType 与 storeValue。
func decodeEntry(b []byte) (logfile string, offset, executeTime int64, schemaName, tableName, gtid string,
	entryType int, storeValue []byte, err error) {
	r := &pbReader{b: b}
	for !r.done() {
		field, wire, err := r.nextField()
		if err != nil {
			return "", 0, 0, "", "", "", 0, nil, err
		}
		switch field {
		case 1: // Header(嵌套)
			if wire != 2 {
				return "", 0, 0, "", "", "", 0, nil, fmt.Errorf("Entry.header 期望 bytes, 实际 wire=%d", wire)
			}
			hb, err := r.bytes()
			if err != nil {
				return "", 0, 0, "", "", "", 0, nil, err
			}
			logfile, offset, executeTime, schemaName, tableName, gtid, err = decodeHeader(hb)
			if err != nil {
				return "", 0, 0, "", "", "", 0, nil, err
			}
		case 2: // EntryType(oneof)
			if wire != 0 {
				return "", 0, 0, "", "", "", 0, nil, fmt.Errorf("Entry.entryType 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return "", 0, 0, "", "", "", 0, nil, err
			}
			entryType = int(v)
		case 3: // storeValue(bytes)
			storeValue, err = r.bytes()
			if err != nil {
				return "", 0, 0, "", "", "", 0, nil, err
			}
		default:
			if err := r.skip(wire); err != nil {
				return "", 0, 0, "", "", "", 0, nil, err
			}
		}
	}
	return logfile, offset, executeTime, schemaName, tableName, gtid, entryType, storeValue, nil
}

// decodeHeader 解析 CanalEntry.Header 关键字段。
func decodeHeader(b []byte) (logfile string, offset, executeTime int64, schemaName, tableName, gtid string, err error) {
	r := &pbReader{b: b}
	for !r.done() {
		field, wire, err := r.nextField()
		if err != nil {
			return "", 0, 0, "", "", "", err
		}
		switch field {
		case 2: // logfileName
			if wire != 2 {
				return "", 0, 0, "", "", "", fmt.Errorf("Header.logfileName 期望 bytes, 实际 wire=%d", wire)
			}
			v, err := r.bytes()
			if err != nil {
				return "", 0, 0, "", "", "", err
			}
			logfile = string(v)
		case 3: // logfileOffset
			if wire != 0 {
				return "", 0, 0, "", "", "", fmt.Errorf("Header.logfileOffset 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return "", 0, 0, "", "", "", err
			}
			offset = int64(v)
		case 6: // executeTime
			if wire != 0 {
				return "", 0, 0, "", "", "", fmt.Errorf("Header.executeTime 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return "", 0, 0, "", "", "", err
			}
			executeTime = int64(v)
		case 8: // schemaName
			if wire != 2 {
				return "", 0, 0, "", "", "", fmt.Errorf("Header.schemaName 期望 bytes, 实际 wire=%d", wire)
			}
			v, err := r.bytes()
			if err != nil {
				return "", 0, 0, "", "", "", err
			}
			schemaName = string(v)
		case 9: // tableName
			if wire != 2 {
				return "", 0, 0, "", "", "", fmt.Errorf("Header.tableName 期望 bytes, 实际 wire=%d", wire)
			}
			v, err := r.bytes()
			if err != nil {
				return "", 0, 0, "", "", "", err
			}
			tableName = string(v)
		case 13: // gtid
			if wire != 2 {
				return "", 0, 0, "", "", "", fmt.Errorf("Header.gtid 期望 bytes, 实际 wire=%d", wire)
			}
			v, err := r.bytes()
			if err != nil {
				return "", 0, 0, "", "", "", err
			}
			gtid = string(v)
		default:
			if err := r.skip(wire); err != nil {
				return "", 0, 0, "", "", "", err
			}
		}
	}
	return logfile, offset, executeTime, schemaName, tableName, gtid, nil
}

// column 解码后的单列信息。
type column struct {
	sqlType   int32
	name      string
	isKey     bool
	updated   bool
	isNull    bool
	value     string
	mysqlType string
}

// decodeColumn 解析 CanalEntry.Column。
func decodeColumn(b []byte) (column, error) {
	var c column
	r := &pbReader{b: b}
	for !r.done() {
		field, wire, err := r.nextField()
		if err != nil {
			return c, err
		}
		switch field {
		case 2: // sqlType
			if wire != 0 {
				return c, fmt.Errorf("Column.sqlType 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return c, err
			}
			c.sqlType = int32(v)
		case 3: // name
			if wire != 2 {
				return c, fmt.Errorf("Column.name 期望 bytes, 实际 wire=%d", wire)
			}
			v, err := r.bytes()
			if err != nil {
				return c, err
			}
			c.name = string(v)
		case 4: // isKey
			if wire != 0 {
				return c, fmt.Errorf("Column.isKey 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return c, err
			}
			c.isKey = v != 0
		case 5: // updated
			if wire != 0 {
				return c, fmt.Errorf("Column.updated 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return c, err
			}
			c.updated = v != 0
		case 6: // isNull
			if wire != 0 {
				return c, fmt.Errorf("Column.isNull 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return c, err
			}
			c.isNull = v != 0
		case 8: // value
			if wire != 2 {
				return c, fmt.Errorf("Column.value 期望 bytes, 实际 wire=%d", wire)
			}
			v, err := r.bytes()
			if err != nil {
				return c, err
			}
			c.value = string(v)
		case 10: // mysqlType
			if wire != 2 {
				return c, fmt.Errorf("Column.mysqlType 期望 bytes, 实际 wire=%d", wire)
			}
			v, err := r.bytes()
			if err != nil {
				return c, err
			}
			c.mysqlType = string(v)
		default:
			if err := r.skip(wire); err != nil {
				return c, err
			}
		}
	}
	return c, nil
}

// rowData 解码后的单行前后镜像。
type rowData struct {
	before []column
	after  []column
}

// decodeRowData 解析 CanalEntry.RowData: beforeColumns(1)/afterColumns(2)。
func decodeRowData(b []byte) (rowData, error) {
	var rd rowData
	r := &pbReader{b: b}
	for !r.done() {
		field, wire, err := r.nextField()
		if err != nil {
			return rd, err
		}
		switch field {
		case 1, 2: // beforeColumns / afterColumns
			if wire != 2 {
				return rd, fmt.Errorf("RowData.columns 期望 bytes, 实际 wire=%d", wire)
			}
			cb, err := r.bytes()
			if err != nil {
				return rd, err
			}
			c, err := decodeColumn(cb)
			if err != nil {
				return rd, err
			}
			if field == 1 {
				rd.before = append(rd.before, c)
			} else {
				rd.after = append(rd.after, c)
			}
		default:
			if err := r.skip(wire); err != nil {
				return rd, err
			}
		}
	}
	return rd, nil
}

// rowChange 解码后的一次事件变更(可能含多行)。
type rowChange struct {
	eventType int
	isDdl     bool
	sql       string
	rowDatas  []rowData
}

// decodeRowChange 解析 CanalEntry.RowChange: tableId(1)/eventType(2)/isDdl(10)/sql(11)/rowDatas(12)/ddlSchemaName(14)。
func decodeRowChange(b []byte) (rowChange, error) {
	var rc rowChange
	r := &pbReader{b: b}
	for !r.done() {
		field, wire, err := r.nextField()
		if err != nil {
			return rc, err
		}
		switch field {
		case 2: // eventType
			if wire != 0 {
				return rc, fmt.Errorf("RowChange.eventType 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return rc, err
			}
			rc.eventType = int(v)
		case 10: // isDdl
			if wire != 0 {
				return rc, fmt.Errorf("RowChange.isDdl 期望 varint, 实际 wire=%d", wire)
			}
			v, err := r.varint()
			if err != nil {
				return rc, err
			}
			rc.isDdl = v != 0
		case 11: // sql
			if wire != 2 {
				return rc, fmt.Errorf("RowChange.sql 期望 bytes, 实际 wire=%d", wire)
			}
			v, err := r.bytes()
			if err != nil {
				return rc, err
			}
			rc.sql = string(v)
		case 12: // rowDatas
			if wire != 2 {
				return rc, fmt.Errorf("RowChange.rowDatas 期望 bytes, 实际 wire=%d", wire)
			}
			rb, err := r.bytes()
			if err != nil {
				return rc, err
			}
			rd, err := decodeRowData(rb)
			if err != nil {
				return rc, err
			}
			rc.rowDatas = append(rc.rowDatas, rd)
		default:
			if err := r.skip(wire); err != nil {
				return rc, err
			}
		}
	}
	return rc, nil
}

// ---- 归一化为 FlatMessage ----

// decodeMessage 是 DecodeValue 的实现:
// Packet -> Messages -> 每条 ROWDATA 的 Entry 转成一个 FlatMessage(多行归并在 Data/Old 中)。
func decodeMessage(value []byte) ([]*FlatMessage, error) {
	if len(value) == 0 {
		return nil, nil
	}
	msgType, compression, body, err := decodePacket(value)
	if err != nil {
		return nil, fmt.Errorf("解析 Canal Packet 失败: %w", err)
	}
	if msgType != packetTypeMessages {
		// 非数据报文(心跳/握手等)不会出现在业务 topic, 直接忽略
		return nil, nil
	}
	if compression != compressionNone && compression != compressionProto2None {
		return nil, fmt.Errorf("%w: %d", errCompression, compression)
	}

	_, entries, err := decodeMessages(body)
	if err != nil {
		return nil, fmt.Errorf("解析 Canal Messages 失败: %w", err)
	}

	out := make([]*FlatMessage, 0, len(entries))
	for _, eb := range entries {
		fm, err := entryToFlat(eb)
		if err != nil {
			return nil, fmt.Errorf("解析 Canal Entry 失败: %w", err)
		}
		if fm != nil {
			out = append(out, fm)
		}
	}
	return out, nil
}

// entryToFlat 将一条 ROWDATA Entry 转成 FlatMessage; 非 ROWDATA / 非 DML 事件返回 nil。
func entryToFlat(eb []byte) (*FlatMessage, error) {
	logfile, offset, executeTime, schemaName, tableName, gtid, entryType, storeValue, err := decodeEntry(eb)
	if err != nil {
		return nil, err
	}
	if entryType != entryTypeRowData || len(storeValue) == 0 {
		return nil, nil // 事务开始/结束、心跳等
	}
	rc, err := decodeRowChange(storeValue)
	if err != nil {
		return nil, err
	}
	// 仅归一化行级 DML; DDL/QUERY/事务事件等无行数据语义, 直接忽略。
	var name string
	switch rc.eventType {
	case eventInsert:
		name = "INSERT"
	case eventUpdate:
		name = "UPDATE"
	case eventDelete:
		name = "DELETE"
	default:
		return nil, nil
	}

	fm := &FlatMessage{
		Database:        schemaName,
		Table:           tableName,
		Type:            name,
		IsDdl:           rc.isDdl,
		ES:              executeTime,
		SQL:             rc.sql,
		GTID:            gtid,
		BinlogFileName:  logfile,
		BinlogPosition:  offset,
		Data:            make([]map[string]*string, 0, len(rc.rowDatas)),
		MySQLType:       make(map[string]string),
		SQLType:         make(map[string]int32),
	}

	for _, rd := range rc.rowDatas {
		after := columnsToRow(rd.after)
		before := columnsToRow(rd.before)
		fillTypes(fm, rd)
		switch rc.eventType {
		case eventInsert:
			fm.Data = append(fm.Data, after)
		case eventUpdate:
			fm.Data = append(fm.Data, after)
			fm.Old = append(fm.Old, before)
		case eventDelete:
			// 与 flat JSON 语义保持一致: DELETE 的"当前行"放 data(beforeColumns),
			// 供 mapper 提取变更前快照与操作人。
			fm.Data = append(fm.Data, before)
		}
	}
	return fm, nil
}

// columnsToRow 将列列表转成 FlatMessage 行 map, isNull 的列映射为 nil(与 flat JSON null 一致)。
func columnsToRow(cols []column) map[string]*string {
	if len(cols) == 0 {
		return nil
	}
	m := make(map[string]*string, len(cols))
	for _, c := range cols {
		if c.isNull {
			m[c.name] = nil
			continue
		}
		v := c.value
		m[c.name] = &v
	}
	return m
}

// fillTypes 从首行 after/before 列汇总 sqlType/mysqlType(供 FlatMessage 携带类型信息)。
func fillTypes(fm *FlatMessage, rd rowData) {
	cols := rd.after
	if len(cols) == 0 {
		cols = rd.before
	}
	for _, c := range cols {
		if _, ok := fm.MySQLType[c.name]; !ok {
			fm.MySQLType[c.name] = c.mysqlType
		}
		if _, ok := fm.SQLType[c.name]; !ok {
			fm.SQLType[c.name] = c.sqlType
		}
	}
}
