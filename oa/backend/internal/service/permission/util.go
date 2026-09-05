package permission

// valuesOf 取 map 的 value 列表
func valuesOf(m map[int64]int64) []int64 {
	values := make([]int64, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// intersects 判断两个字符串切片是否存在交集
func intersects(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}

// dedup 去重
func dedup(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
