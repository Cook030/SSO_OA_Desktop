package protocol

import "encoding/json"

// ServerMessage 服务端 → 客户端的消息（dev.md §5 最小协议）。
// 只保留两类：事件推送（event）与未来扩展用的业务消息。
type ServerMessage struct {
	Type      string `json:"type"`                // "event"
	EventID   string `json:"eventId"`             // 幂等键：客户端按 eventId 去重
	EventType string `json:"eventType,omitempty"` // "session_replaced" / "session_terminated"
	Reason    string `json:"reason,omitempty"`    // "replaced_by_new_login" 等
}

// ClientMessage 客户端 → 服务端的消息。目前只有事件确认（ack）。
type ClientMessage struct {
	Type    string `json:"type"`    // "ack"
	EventID string `json:"eventId"` // 被确认的事件 ID
}

// Ack 构造确认消息
func Ack(eventID string) ClientMessage {
	return ClientMessage{Type: ClientMsgTypeAck, EventID: eventID}
}

// MarshalEvent 构造事件推送负载
func MarshalEvent(eventID, eventType, reason string) ([]byte, error) {
	return json.Marshal(ServerMessage{
		Type:      ServerMsgTypeEvent,
		EventID:   eventID,
		EventType: eventType,
		Reason:    reason,
	})
}
