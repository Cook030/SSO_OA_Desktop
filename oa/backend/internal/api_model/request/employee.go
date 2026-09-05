package request

// CreateEmployeeRequest 新增员工请求
type CreateEmployeeRequest struct {
	Name        string  `json:"name"`
	Phone       string  `json:"phone"`
	EmailPrefix string  `json:"emailPrefix"`
	Account     string  `json:"account"`
	Department  string  `json:"department"`
	RoleIDs     []int64 `json:"roleIds"`
	Password    string  `json:"password"`
}

// UpdateEmployeeRequest 编辑员工请求
type UpdateEmployeeRequest struct {
	Name        string  `json:"name"`
	Phone       string  `json:"phone"`
	EmailPrefix string  `json:"emailPrefix"`
	Account     string  `json:"account"`
	Department  string  `json:"department"`
	RoleIDs     []int64 `json:"roleIds"`
}
