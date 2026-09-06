package mapper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"mh-audit-consumer/internal/canal"
)

// dedupKey 由 binlog 位点和行号生成，确保同一变更重放时可幂等写入。
func dedupKey(flat *canal.FlatMessage, rowIdx int) string {
	position := flat.BinlogFileName
	if position == "" && flat.GTID != "" {
		position = "gtid:" + flat.GTID
	}
	raw := fmt.Sprintf("%s|%s|%s|%s|%d|%d", flat.Database, flat.Table, flat.Type, position, flat.BinlogPosition, rowIdx)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
