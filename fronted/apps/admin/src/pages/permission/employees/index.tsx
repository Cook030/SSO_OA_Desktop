import {
  MhButton,
  MhCard,
  MhCheckbox,
  MhForm,
  MhIcon,
  MhInput,
  MhMessage,
  MhModal,
  MhPagination,
  MhSelect,
  MhSpace,
  MhTable,
  MhTag
} from "@mh-repo/ui";
import type React from "react";
import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router";
import type { BreadcrumbItem } from "../../../components/Breadcrumb";
import PageHeader from "../../../components/PageHeader";
import request from "../../../utils/request";
import TableDrawerList from "./Components/TableDrawerList";
import styles from "./index.module.less";

/** 平台权限类型（来自后端） */
interface PlatformPermission {
  id: number;
  name: string;
}

/** 员工数据类型（来自后端） */
interface Employee {
  id: number;
  displayId: string;
  name: string;
  phone: string;
  email: string;
  account: string;
  department: string;
  platformPermissions: PlatformPermission[];
}

/** 部门颜色映射 */
const departmentColorMap: Record<string, string> = {
  技术部: "geekblue",
  销售部: "lime",
  运营部: "magenta",
  人力资源部: "orange",
  行政部: "cyan",
  财务部: "gold",
  总裁办: "purple",
  研发部: "blue",
  产品部: "green",
  媒介部: "volcano",
  __default: "default"
};

/** 部门下拉选项（前端写定） */
const DEPARTMENT_OPTIONS = [
  { label: "研发部", value: "研发部" },
  { label: "产品部", value: "产品部" },
  { label: "总裁办", value: "总裁办" },
  // { label: "技术部", value: "技术部" },
  { label: "媒介部", value: "媒介部" },
  { label: "财务部", value: "财务部" },
  { label: "行政部", value: "行政部" },
  { label: "人力资源部", value: "人力资源部" },
  { label: "运营部", value: "运营部" },
  { label: "销售部", value: "销售部" }
];

const PermissionEmployees: React.FC = () => {
  const navigate = useNavigate();
  const [form] = MhForm.useForm();
  const [loading, setLoading] = useState(false);

  // ----- 列表数据 -----
  const [tableData, setTableData] = useState<Employee[]>([]);
  const [total, setTotal] = useState(0);

  // ----- 分页 -----
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  // ----- 筛选条件 -----
  const [searchText, setSearchText] = useState("");
  const [debouncedKeyword, setDebouncedKeyword] = useState("");
  const [searchDepartment, setSearchDepartment] = useState<string | undefined>(undefined);
  const [searchPlatform, setSearchPlatform] = useState<number | undefined>(undefined);
  const fetchIdRef = useRef(0);

  // ----- 下拉选项 -----
  const [platformOptions, setPlatformOptions] = useState<{ label: string; value: number }[]>([]);

  // ----- 选中行 -----
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [batchPlatforms, setBatchPlatforms] = useState<number[]>([]);

  // ----- 抽屉 -----
  // const [drawerOpen, setDrawerOpen] = useState(false);
  // const [drawerForm] = MhForm.useForm();
  const [editDrawerOpen, setEditDrawerOpen] = useState(false);
  const [editForm] = MhForm.useForm();
  const [editingRecord, setEditingRecord] = useState<Employee | null>(null);

  // ----- 面包屑 -----
  const breadcrumbItems: BreadcrumbItem[] = [
    { key: "permission", title: "权限管理", path: "/", clickable: true },
    { key: "/permission/employees", title: "员工管理" }
  ];

  const handleBreadcrumbClick = (item: BreadcrumbItem) => {
    if (item.path) navigate(item.path);
  };

  // ===== 加载员工列表 =====
  // 后端 keyword 搜索有 bug（任何非空值均返回空），改为前端过滤
  const fetchEmployees = async (page = current) => {
    const reqId = ++fetchIdRef.current;
    setLoading(true);
    try {
      const params: any = {
        page,
        pageSize
      };
      // keyword 不发送给后端，改用前端过滤
      if (searchDepartment && searchDepartment !== "__all__") params.department = searchDepartment;
      if (searchPlatform && searchPlatform > 0) params.platformId = searchPlatform;

      // 有搜索关键字时一次性拉取全部数据用于前端过滤
      if (debouncedKeyword) {
        params.page = 1;
        params.pageSize = 9999;
      }

      const response = (await request.get("/employees", { params })) as any;
      // 防止竞态：丢弃过期的响应
      if (reqId !== fetchIdRef.current) return;
      const result = response;

      if (result) {
        let list = result.list || [];
        let total = result.total || 0;

        // 前端关键字过滤（姓名/手机号/邮箱）
        if (debouncedKeyword) {
          const kw = debouncedKeyword.toLowerCase();
          list = list.filter(
            (emp: any) =>
              emp.name?.toLowerCase().includes(kw) ||
              emp.phone?.toLowerCase().includes(kw) ||
              emp.email?.toLowerCase().includes(kw)
          );
          total = list.length;
          // 前端分页
          const startIdx = (page - 1) * pageSize;
          list = list.slice(startIdx, startIdx + pageSize);
        }

        setTableData(list);
        setTotal(total);
      }
    } catch (error: any) {
      if (reqId !== fetchIdRef.current) return;
      MhMessage.error(error.message || "加载员工列表失败");
    } finally {
      if (reqId === fetchIdRef.current) setLoading(false);
    }
  };

  // ===== 加载下拉选项 =====
  const fetchOptions = async () => {
    try {
      // 获取平台列表（用于下拉和权限映射）
      const platformData = (await request.get("/platforms", { params: { page: 1, pageSize: 1000 } })) as any;
      const list = platformData.list || [];
      setPlatformOptions(list.map((p: any) => ({ label: p.name, value: p.id })));
    } catch (error: any) {
      MhMessage.warning(error.message || "加载下拉选项失败，部分筛选可能不可用");
    }
  };

  // ===== 初始加载 =====
  useEffect(() => {
    fetchOptions();
  }, []);

  // 搜索防抖：输入停止 300ms 后才触发请求（同时重置页码）
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedKeyword(searchText);
      setCurrent(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchText]);

  // 数据加载（筛选 + 分页统一在一个 effect 中，避免重复请求）
  useEffect(() => {
    fetchEmployees(current);
  }, [current, pageSize, debouncedKeyword, searchDepartment, searchPlatform]);

  // useEffect(() => {
  //   if (editingRecord) {
  //     // 将数据注入到表单中，name="ID" 会自动匹配
  //     form.setFieldsValue({
  //       ID: editingRecord.id
  //       // 如果有其他字段，也可以在这里一起赋值
  //       // name: editingRecord.name,
  //     });
  //   }
  // }, [editingRecord]);

  // ===== 行选择 =====
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => setSelectedRowKeys(keys),
    selections: [
      {
        key: "select-all",
        text: "全选所有",
        onSelect: () => {
          const allKeys = tableData.map(item => item.id);
          setSelectedRowKeys(allKeys);
        }
      },
      {
        key: "clear-all",
        text: "清空所有",
        onSelect: () => setSelectedRowKeys([])
      }
    ]
  };

  // ===== 编辑员工 =====
  const handleEditEmployee = (record: Employee) => {
    setEditingRecord(record);
    editForm.setFieldsValue({
      name: record.name,
      displayId: record.displayId,
      phone: record.phone,
      email: record.email.replace(/@maplehaze\.cn$/, ""),
      account: record.account,
      department: record.department,
      platforms: record.platformPermissions.map(p => p.id)
    });
    setEditDrawerOpen(true);
  };

  const handleEditFinish = async (values: any) => {
    if (!editingRecord) return;
    try {
      // 1. 获取表单中的值
      const rawPlatforms = values.platforms || [];
      // 2. 【关键修改】兼容处理：将选中的值转换为数字 ID
      const platformIds = rawPlatforms
        .map((val: any) => {
          // 情况 A: 如果 val 已经是数字（来自回填的 id），直接返回
          if (typeof val === "number") return val;

          // 情况 B: 如果 val 是字符串（来自用户新勾选的名称），则进行查找转换
          if (typeof val === "string") {
            const platform = platformOptions.find(p => p.label === `${val}平台`);
            return platform ? platform.value : null;
          }

          return null;
        })
        .filter((id: number | null): id is number => id !== null);
      const payload = {
        name: values.name,
        phone: values.phone,
        emailPrefix: values.email,
        account: values.account,
        department: values.department,
        platformIds: platformIds
      };

      await request.put(`/employees/${editingRecord.id}`, payload);
      MhMessage.success("员工信息已更新");
      editForm.resetFields();
      setEditDrawerOpen(false);
      setEditingRecord(null);
      fetchEmployees();
    } catch (error: any) {
      MhMessage.error(error.message || "更新员工失败");
    }
  };

  // ==== 密码重置 ====
  const handleResetPassword = () => {
    MhModal.confirm({
      title: "确定要重置用户密码？",
      content: "重置密码后密码将变更为Mhint@123",
      centered: true,
      type: "warning",
      onOk: async () => {
        if (!editingRecord?.id) {
          MhMessage.error("无法获取员工ID");
          return;
        }
        try {
          await request.put(`employees/${editingRecord.id}/reset-password`);
          MhMessage.success("密码重置成功");
          editForm.setFieldsValue({ password: "Mhint@123" });
        } catch (error: any) {
          MhMessage.error(error?.message || "重置失败，请稍后重试");
        }
      }
    });
  };

  // ===== 删除员工 =====
  const handleDeleteRow = (record: Employee) => {
    MhModal.confirm({
      title: "确定删除该用户吗？",
      content: "删除后，该用户将无法登录后台系统！",
      centered: true,
      type: "warning",
      onOk: async () => {
        try {
          await request.delete(`/employees/${record.id}`);
          MhMessage.success("删除成功");
          setSelectedRowKeys(prev => prev.filter(id => id !== record.id));
          fetchEmployees();
        } catch (error: any) {
          MhMessage.error(error.message || "删除失败");
        }
      }
    });
  };

  // ===== 批量操作 =====
  // 批量添加权限：POST /api/employees/permissions/batch
  const handleBatchApply = async () => {
    if (batchPlatforms.length === 0) {
      MhMessage.warning("请选择至少一个平台权限");
      return;
    }
    if (selectedRowKeys.length === 0) {
      MhMessage.warning("请先选择员工");
      return;
    }
    try {
      const userIds = selectedRowKeys.map(id => Number(id));
      await request.post("/employees/permissions/batch", {
        userIds,
        platformIds: batchPlatforms
      });
      MhMessage.success(`已为 ${selectedRowKeys.length} 名员工添加权限`);
      setSelectedRowKeys([]);
      setBatchPlatforms([]);
      fetchEmployees();
    } catch (error: any) {
      MhMessage.error(error.message || "批量设置权限失败");
    }
  };

  // 批量清除权限：DELETE /api/employees/permissions/batch（清除所有权限，只需 userIds）
  const handleClearPermissions = async () => {
    // 1. 校验员工选择
    if (selectedRowKeys.length === 0) {
      MhMessage.warning("请先选择员工");
      return;
    }

    // 2. 关键校验：必须选择要清除的平台（与添加逻辑保持一致）
    // 如果不传 platformIds，后端通常会理解为“清除该用户所有平台权限”
    if (batchPlatforms.length === 0) {
      MhMessage.warning("请选择要清除的平台权限");
      return;
    }

    try {
      const userIds = selectedRowKeys.map(id => Number(id));

      // 3. 构造请求体（参考 handleBatchApply 的结构）
      // 注意：DELETE 请求传 body 时，通常需要包裹在 data 属性中（取决于 request 封装）
      await request.delete("/employees/permissions/batch", {
        data: {
          userIds,
          platformIds: batchPlatforms // 传入选中的平台 ID，告诉后端只删这些
        }
      });

      MhMessage.success(`已清除 ${selectedRowKeys.length} 名员工的指定平台权限`);
      setSelectedRowKeys([]);
      setBatchPlatforms([]);
      fetchEmployees();
    } catch (error: any) {
      MhMessage.error(error.message || "批量清除权限失败");
    }
  };

  const handleCloseSelection = () => {
    setSelectedRowKeys([]);
    setBatchPlatforms([]);
  };

  // ===== 表格列 =====
  const columns = [
    { title: "员工", dataIndex: "name", key: "name" },
    { title: "手机号", dataIndex: "phone", key: "phone" },
    { title: "邮箱", dataIndex: "email", key: "email" },
    {
      title: "部门",
      dataIndex: "department",
      key: "department",
      render: (text: string) => (
        <MhTag color={departmentColorMap[text] || departmentColorMap.__default} variant="outlined">
          {text}
        </MhTag>
      )
    },
    {
      title: "平台权限",
      dataIndex: "platformPermissions",
      key: "platformPermissions",
      render: (permissions: PlatformPermission[]) => permissions.map(p => p.name).join(" ；") || "-"
    },
    {
      title: "操作",
      key: "action",
      render: (_: unknown, record: Employee) => (
        <MhSpace>
          <a onClick={() => handleEditEmployee(record)} style={{ color: "#1677ff" }}>
            编辑
          </a>
          <a onClick={() => handleDeleteRow(record)} style={{ color: "#1677ff" }}>
            删除
          </a>
        </MhSpace>
      )
    }
  ];

  // ===== 平台选项（用于批量操作下拉） =====
  const platformSelectOptions = platformOptions.map(p => ({ label: p.label, value: p.value }));
  // 抽屉中的 Checkbox 选项（值用平台名去掉"平台"后缀，以匹配原UI）
  const checkboxOptions = platformOptions.map(p => ({
    label: p.label,
    value: p.value // 这里直接用 p.value (假设 p.value 就是 ID)
  }));

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "calc(100vh - 64px - 18px)",
        paddingBottom: 18
      }}
    >
      <PageHeader
        title="员工管理"
        breadcrumbItems={breadcrumbItems}
        onBreadcrumbClick={handleBreadcrumbClick}
        showFavorite={false}
      />

      <MhForm
        form={form}
        layout="vertical"
        autoComplete="off"
        style={{ flex: 1, display: "flex", flexDirection: "column" }}
      >
        <MhCard style={{ flex: 1, display: "flex", flexDirection: "column" }}>
          {/* 工具栏 */}
          <div style={{ display: "flex", flexWrap: "wrap", gap: 10, marginBottom: 16 }}>
            <MhButton type="primary" onClick={() => navigate("/permission/employees/AddEmployee")}>
              新增员工
            </MhButton>
            <MhInput.Search
              placeholder="姓名/手机号/邮箱"
              value={searchText}
              onChange={e => setSearchText(e.target.value)}
              onSearch={value => {
                setDebouncedKeyword(value);
                setCurrent(1);
              }}
              style={{ width: 280 }}
              allowClear
            />
            <MhSelect
              placeholder="部门：全部"
              style={{ width: 250 }}
              value={searchDepartment}
              onChange={(val: string) => {
                setSearchDepartment(val);
                setCurrent(1);
              }}
              options={[{ label: "全部", value: "__all__" }, ...DEPARTMENT_OPTIONS]}
            />
            <MhSelect
              placeholder="平台：全部"
              style={{ width: 250 }}
              value={searchPlatform}
              onChange={val => {
                setSearchPlatform(val as number);
                setCurrent(1);
              }}
              options={[{ label: "全部", value: -1 }, ...platformSelectOptions]}
            />
            <MhIcon
              type="ReloadOutlined"
              onClick={() => {
                setSearchText("");
                setDebouncedKeyword("");
                setSearchDepartment(undefined);
                setSearchPlatform(undefined);
                setCurrent(1);
              }}
              style={{ cursor: "pointer", fontSize: 18, color: "#1677ff" }}
            />
          </div>

          {/* 批量操作条 */}
          <div style={{ marginTop: 16 }}>
            {selectedRowKeys.length > 0 && (
              <div style={{ display: "flex", height: 40 }}>
                <div style={{ flex: 2, backgroundColor: "#e3f1fc" }}>
                  <div style={{ color: "#1777ff", paddingTop: 8, paddingLeft: 16 }}>
                    已选择 {selectedRowKeys.length} 名员工
                  </div>
                </div>
                <div
                  style={{
                    flex: 7,
                    display: "flex",
                    justifyContent: "space-between",
                    gap: 8,
                    backgroundColor: "#e3f1fc",
                    paddingTop: 4,
                    marginRight: 8,
                    marginLeft: 8,
                    paddingLeft: 16
                  }}
                >
                  <div>
                    <MhSelect
                      mode="multiple"
                      placeholder="请选择平台权限"
                      style={{ minWidth: 300 }}
                      options={platformSelectOptions}
                      value={batchPlatforms}
                      onChange={(val: number[]) => setBatchPlatforms(val)}
                      allowClear
                    />
                  </div>
                  <div>
                    <MhButton type="primary" onClick={handleBatchApply} style={{ marginRight: 16 }}>
                      添加权限
                    </MhButton>
                    <MhButton onClick={handleClearPermissions} style={{ marginRight: 16 }} danger>
                      清除权限
                    </MhButton>
                  </div>
                </div>
                <div
                  style={{
                    textAlign: "center",
                    flex: 1,
                    backgroundColor: "#e3f1fc",
                    paddingTop: 5
                  }}
                >
                  <MhButton onClick={handleCloseSelection}>关闭</MhButton>
                </div>
              </div>
            )}
          </div>

          <MhTable
            columns={columns}
            dataSource={tableData}
            rowKey="id"
            pagination={false}
            size="small"
            rowSelection={rowSelection}
            loading={loading}
          />

          <div className={styles["table-base-footer"]}>
            <div className={styles["table-base-total"]}>总共 {total} 条</div>
            <MhPagination
              size="small"
              current={current}
              pageSize={pageSize}
              total={total}
              showSizeChanger
              pageSizeOptions={[10, 20, 50, 100]}
              onChange={(page, size) => {
                setCurrent(page);
                setPageSize(size);
              }}
            />
          </div>
        </MhCard>
      </MhForm>

      {/* 编辑员工抽屉 */}
      <TableDrawerList
        open={editDrawerOpen}
        title="编辑员工"
        width={560}
        onCancel={() => {
          editForm.resetFields();
          setEditDrawerOpen(false);
          setEditingRecord(null);
        }}
        onCustomFinish={handleEditFinish}
        confirmText="确定"
        cancelText="取消"
        form={editForm}
        initialValues={{ displayId: editingRecord?.displayId } as any}
      >
        <div style={{ marginBottom: 24 }}>
          <div
            style={{
              fontSize: 14,
              fontWeight: 500,
              marginBottom: 16,
              borderBottom: "1px solid #f0f0f0",
              paddingBottom: 8
            }}
          >
            基础信息
          </div>

          <MhForm.Item
            label={
              <span className={styles.requiredLabel}>
                <span className={styles.requiredMark}>*</span>
                <span>员工姓名</span>
              </span>
            }
            name="name"
            rules={[{ required: true, message: "请输入员工姓名" }]}
          >
            <MhInput placeholder="请输入用户名" allowClear />
          </MhForm.Item>

          <MhForm.Item
            label={
              <span className={styles.requiredLabel}>
                <span className={styles.requiredMark}>*</span>
                <span>所属部门</span>
              </span>
            }
            name="department"
            rules={[{ required: true, message: "请选择所属部门" }]}
          >
            <MhSelect placeholder="请选择" options={DEPARTMENT_OPTIONS} allowClear />
          </MhForm.Item>

          <MhForm.Item
            label={
              <span className={styles.requiredLabel}>
                <span className={styles.requiredMark}>*</span>
                <span>账号</span>
              </span>
            }
            name="account"
            required={false}
            rules={[{ required: true, message: "请输入登录账号" }]}
          >
            <MhInput placeholder="请输入登录账号" allowClear />
          </MhForm.Item>

          <MhForm.Item
            label={
              <span className={styles.requiredLabel}>
                <span className={styles.requiredMark}>*</span>
                <span>登录密码</span>
              </span>
            }
            name="password"
          >
            <div style={{ display: "flex", gap: 8 }}>
              <MhInput placeholder="*****" disabled allowClear />
              <MhButton
                variant="outlined"
                color="primary"
                onClick={() => {
                  handleResetPassword();
                }}
              >
                重置密码
              </MhButton>
            </div>
          </MhForm.Item>

          {/* <MhForm.Item
            label={
              <span className={styles.requiredLabel}>
                <span className={styles.requiredMark}>*</span>
                <span>员工ID</span>
              </span>
            }
            name="displayId"
          >
            <MhInput allowClear disabled />
          </MhForm.Item> */}

          <MhForm.Item
            label={
              <span className={styles.requiredLabel}>
                <span>手机号</span>
              </span>
            }
            name="phone"
            required={false}
            rules={[
              {
                validator: (_: unknown, value: string) => {
                  if (!value) return Promise.resolve();
                  if (!/^1\d{10}$/.test(value)) return Promise.reject(new Error("请输入有效的手机号"));
                  return Promise.resolve();
                }
              }
            ]}
          >
            <MhInput placeholder="请输入手机号" allowClear maxLength={11} />
          </MhForm.Item>

          <MhForm.Item
            label={
              <span className={styles.requiredLabel}>
                <span>员工邮箱</span>
              </span>
            }
            name="email"
            required={false}
          >
            <MhSpace.Compact style={{ width: "100%" }}>
              <MhInput
                key={editingRecord?.id ?? "empty"}
                defaultValue={editingRecord?.email?.replace("@maplehaze.cn", "")}
                allowClear
              />
              <span
                style={{
                  display: "flex",
                  alignItems: "center",
                  padding: "0 12px",
                  background: "#fafafa",
                  border: "1px solid #d9d9d9",
                  borderLeft: "none",
                  borderRadius: "0 6px 6px 0",
                  color: "rgba(0,0,0,0.45)"
                }}
              >
                @maplehaze.cn
              </span>
            </MhSpace.Compact>
          </MhForm.Item>
        </div>

        <div>
          <div
            style={{
              fontSize: 14,
              fontWeight: 500,
              marginBottom: 16,
              borderBottom: "1px solid #f0f0f0",
              paddingBottom: 8
            }}
          >
            权限管理
          </div>
          <MhForm.Item label="平台权限" name="platforms">
            <MhCheckbox.Group
              options={checkboxOptions}
              style={{
                display: "flex",
                flexDirection: "column", // 让选项竖向排列
                gap: "8px"
              }}
            />
          </MhForm.Item>
        </div>
      </TableDrawerList>
    </div>
  );
};

export default PermissionEmployees;
