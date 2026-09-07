package mapper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"mh-audit-consumer/internal/canal"
)

// dedupKey 由数据源标识、GTID、binlog 位点和行号生成，确保同一变更重放时可幂等写入。
func dedupKey(flat *canal.FlatMessage, rowIdx int, sourceID string) string {
	position := flat.BinlogFileName
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d|%d", sourceID, flat.Database, flat.Table, flat.Type, flat.GTID, position, flat.BinlogPosition, rowIdx)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
