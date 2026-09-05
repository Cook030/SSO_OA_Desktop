// Package sanitize 对入库快照中的敏感列做脱敏(整体替换)。
package sanitize

// Sanitizer 按列名黑名单打码。
type Sanitizer struct {
	fields      map[string]struct{}
	replacement string
}

// New 构造脱敏器。
func New(fields []string, replacement string) *Sanitizer {
	set := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		set[f] = struct{}{}
	}
	return &Sanitizer{fields: set, replacement: replacement}
}

// Row 深拷贝一行并把敏感列值替换为脱敏占位符。
func (s *Sanitizer) Row(row map[string]*string) map[string]*string {
	if row == nil {
		return nil
	}
	out := make(map[string]*string, len(row))
	for col, val := range row {
		if _, masked := s.fields[col]; masked {
			out[col] = stringPtr(s.replacement)
			continue
		}
		out[col] = cloneString(val)
	}
	return out
}

func stringPtr(v string) *string {
	return &v
}

func cloneString(v *string) *string {
	if v == nil {
		return nil
	}
	return stringPtr(*v)
}
