package canal

// ---- 测试辅助: 最小 protobuf(proto3) wire 编码器, 仅用于构造测试消息 ----

func tVarint(v uint64) []byte {
	var out []byte
	for v >= 0x80 {
		out = append(out, byte(v)|0x80)
		v >>= 7
	}
	return append(out, byte(v))
}

func tTag(field, wire int) []byte { return tVarint(uint64(field)<<3 | uint64(wire)) }

func tLenField(field int, payload []byte) []byte {
	out := tTag(field, 2)
	out = append(out, tVarint(uint64(len(payload)))...)
	return append(out, payload...)
}

func tStrField(field int, s string) []byte { return tLenField(field, []byte(s)) }

func tVarintField(field int, v uint64) []byte {
	out := tTag(field, 0)
	return append(out, tVarint(v)...)
}

func tBoolField(field int, v bool) []byte {
	if v {
		return tVarintField(field, 1)
	}
	return tVarintField(field, 0)
}

// ---- message 构造 ----

// tColumn 编码 Column: name(3)/isNull(6)/value(8)。
func tColumn(name string, isNull bool, value string) []byte {
	out := tStrField(3, name)
	out = append(out, tBoolField(6, isNull)...)
	if !isNull {
		out = append(out, tStrField(8, value)...)
	}
	return out
}

// tRowData 编码 RowData: beforeColumns(1)/afterColumns(2)。
func tRowData(before, after [][]byte) []byte {
	var out []byte
	for _, b := range before {
		out = append(out, tLenField(1, b)...)
	}
	for _, a := range after {
		out = append(out, tLenField(2, a)...)
	}
	return out
}

// tRowChange 编码 RowChange: eventType(2)/isDdl(10)/rowDatas(12)。
func tRowChange(eventType int, isDdl bool, rowDatas ...[]byte) []byte {
	out := tVarintField(2, uint64(eventType))
	out = append(out, tBoolField(10, isDdl)...)
	for _, rd := range rowDatas {
		out = append(out, tLenField(12, rd)...)
	}
	return out
}

// tEntry 编码 Entry: header(1)/entryType(2)/storeValue(3)。
func tEntry(schema, table, logfile string, offset, executeTime int64, gtid string,
	entryType int, storeValue []byte) []byte {
	header := tStrField(2, logfile)
	header = append(header, tVarintField(3, uint64(offset))...)
	header = append(header, tVarintField(6, uint64(executeTime))...)
	header = append(header, tStrField(8, schema)...)
	header = append(header, tStrField(9, table)...)
	if gtid != "" {
		header = append(header, tStrField(13, gtid)...)
	}
	out := tLenField(1, header)
	out = append(out, tVarintField(2, uint64(entryType))...)
	if storeValue != nil {
		out = append(out, tLenField(3, storeValue)...)
	}
	return out
}

// tMessages 编码 Messages: batch_id(1)/messages(2, repeated)。
func tMessages(batchID int64, entries ...[]byte) []byte {
	out := tVarintField(1, uint64(batchID))
	for _, e := range entries {
		out = append(out, tLenField(2, e)...)
	}
	return out
}

// tPacket 编码 Packet: type(3)/compression(4)/body(5)。
func tPacket(msgType int, body []byte) []byte {
	out := tVarintField(3, uint64(msgType))
	out = append(out, tVarintField(4, compressionNone)...) // Compression.NONE
	out = append(out, tLenField(5, body)...)
	return out
}
