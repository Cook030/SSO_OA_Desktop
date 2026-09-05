package response

// EmployeeListItemDTO 员工列表项
type EmployeeListItemDTO struct {
	ID                  int64                `json:"id"`
	DisplayID           string               `json:"displayId"`
	Name                string               `json:"name"`
	Account             string               `json:"account"`
	Phone               string               `json:"phone"`
	Email               string               `json:"email"`
	Department          string               `json:"department"`
	Roles               []RoleOptionDTO      `json:"roles"`
	PlatformPermissions []PlatformPermission `json:"platformPermissions"`
}

// CreateEmployeeDTO 新增员工响应
type CreateEmployeeDTO struct {
	ID        int64  `json:"id"`
	DisplayID string `json:"displayId"`
	Name      string `json:"name"`
	Account   string `json:"account"`
	Email     string `json:"email"`
}

// UpdateEmployeeDTO 编辑员工响应
type UpdateEmployeeDTO struct {
	ID         int64  `json:"id"`
	DisplayID  string `json:"displayId"`
	Name       string `json:"name"`
	Account    string `json:"account"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Department string `json:"department"`
}

// EmployeePageResult 员工分页结果
type EmployeePageResult struct {
	Total int64                 `json:"total"`
	List  []EmployeeListItemDTO `json:"list"`
}
