package canal

import "fmt"

// decodePacket 解析 CanalProtocol.Packet: type(3)/compression(4)/body(5)。
func decodePacket(data []byte) (messageType, compression int, body []byte, err error) {
	reader := &pbReader{b: data}
	for !reader.done() {
		field, wire, err := reader.nextField()
		if err != nil {
			return 0, 0, nil, err
		}
		switch field {
		case 3:
			if wire != 0 {
				return 0, 0, nil, fmt.Errorf("Packet.type 期望 varint, 实际 wire=%d", wire)
			}
			value, err := reader.varint()
			if err != nil {
				return 0, 0, nil, err
			}
			messageType = int(value)
		case 4:
			if wire != 0 {
				return 0, 0, nil, fmt.Errorf("Packet.compression 期望 varint, 实际 wire=%d", wire)
			}
			value, err := reader.varint()
			if err != nil {
				return 0, 0, nil, err
			}
			compression = int(value)
		case 5:
			body, err = reader.bytes()
			if err != nil {
				return 0, 0, nil, err
			}
		default:
			if err := reader.skip(wire); err != nil {
				return 0, 0, nil, err
			}
		}
	}
	return messageType, compression, body, nil
}

// decodeMessages 解析 CanalProtocol.Messages 的 batch ID 与 Entry 列表。
func decodeMessages(data []byte) (batchID int64, entries [][]byte, err error) {
	reader := &pbReader{b: data}
	for !reader.done() {
		field, wire, err := reader.nextField()
		if err != nil {
			return 0, nil, err
		}
		switch field {
		case 1:
			if wire != 0 {
				return 0, nil, fmt.Errorf("Messages.batch_id 期望 varint, 实际 wire=%d", wire)
			}
			value, err := reader.varint()
			if err != nil {
				return 0, nil, err
			}
			batchID = int64(value)
		case 2:
			if wire != 2 {
				return 0, nil, fmt.Errorf("Messages.messages 期望 bytes, 实际 wire=%d", wire)
			}
			entry, err := reader.bytes()
			if err != nil {
				return 0, nil, err
			}
			entries = append(entries, entry)
		default:
			if err := reader.skip(wire); err != nil {
				return 0, nil, err
			}
		}
	}
	return batchID, entries, nil
}
