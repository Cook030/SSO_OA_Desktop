package repository

import (
	"permission-system/internal/db_model"
	"permission-system/internal/db_model/query"
)

// PlatformRepository 平台数据访问层
type PlatformRepository struct {
	q *query.Query
}

// NewPlatformRepository 创建平台仓库
func NewPlatformRepository(q *query.Query) *PlatformRepository {
	return &PlatformRepository{q: q}
}

// Create 创建平台
func (r *PlatformRepository) Create(platform *db_model.SysPlatform) error {
	return r.q.SysPlatform.Create(platform)
}

// FindByID 根据ID查找平台
func (r *PlatformRepository) FindByID(id int64) (*db_model.SysPlatform, error) {
	return r.q.SysPlatform.Where(r.q.SysPlatform.ID.Eq(id)).First()
}

// Update 更新平台
func (r *PlatformRepository) Update(platform *db_model.SysPlatform) error {
	_, err := r.q.SysPlatform.Where(r.q.SysPlatform.ID.Eq(platform.ID)).Updates(platform)
	return err
}

// Delete 物理删除平台(关联权限在 service 层手动删除)
func (r *PlatformRepository) Delete(id int64) error {
	_, err := r.q.SysPlatform.Where(r.q.SysPlatform.ID.Eq(id)).Delete()
	return err
}

// List 分页查询平台列表
func (r *PlatformRepository) List(page, pageSize int) ([]*db_model.SysPlatform, int64, error) {
	offset := (page - 1) * pageSize
	return r.q.SysPlatform.Order(r.q.SysPlatform.CreateTime.Desc()).FindByPage(offset, pageSize)
}

// ExistsByName 检查平台名称是否存在
func (r *PlatformRepository) ExistsByName(name string, excludeID int64) (bool, error) {
	p := r.q.SysPlatform.Where(r.q.SysPlatform.Name.Eq(name))
	if excludeID > 0 {
		p = p.Where(r.q.SysPlatform.ID.Neq(excludeID))
	}
	count, err := p.Count()
	return count > 0, err
}

// ExistsByLink 检查平台链接是否存在
func (r *PlatformRepository) ExistsByLink(link string, excludeID int64) (bool, error) {
	p := r.q.SysPlatform.Where(r.q.SysPlatform.Link.Eq(link))
	if excludeID > 0 {
		p = p.Where(r.q.SysPlatform.ID.Neq(excludeID))
	}
	count, err := p.Count()
	return count > 0, err
}

// FindByIDs 根据ID列表批量查询平台
func (r *PlatformRepository) FindByIDs(ids []int64) ([]*db_model.SysPlatform, error) {
	if len(ids) == 0 {
		return []*db_model.SysPlatform{}, nil
	}
	return r.q.SysPlatform.Where(r.q.SysPlatform.ID.In(ids...)).Find()
}
