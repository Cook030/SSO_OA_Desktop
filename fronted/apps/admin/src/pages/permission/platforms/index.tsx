import { MhButton, MhCard, MhForm, MhInput, MhModal, MhPagination, MhSpace, MhTable } from "@mh-repo/ui";
// import axios from "axios";
import type React from "react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import type { BreadcrumbItem } from "../../../components/Breadcrumb";
import PageHeader from "../../../components/PageHeader";
import request from "../../../utils/request";
import type { TablePopUpFormValues } from "./Components/TablePopUpForm";
import TablePopUpForm from "./Components/TablePopUpForm";
import styles from "./index.module.less";

/** 表格数据类型 */
interface TableData {
  key: string; // 对应后端的 id
  platformName: string; // 对应后端的 name
  authorizedPerson: number; // 对应后端的 permissionCount
  platformUrl: string; // 对应后端的 link
}

/** 弹窗表单数据类型 (用于新增/编辑) */
interface PlatformFormValues {
  name: string;
  link: string;
}

const PlatformManagement: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);

  // 表格数据状态
  const [tableData, setTableData] = useState<TableData[]>([]);

  // 编辑状态
  const [editingKey, setEditingKey] = useState<string>("");
  const [editingData, setEditingData] = useState<PlatformFormValues>({
    name: "",
    link: ""
  });

  // 弹窗状态
  const [isFormOpen, setIsFormOpen] = useState<boolean>(false);
  const [form] = MhForm.useForm();

  /** 面包屑配置 */
  const breadcrumbItems: BreadcrumbItem[] = [
    { key: "permission", title: "权限管理", path: "/", clickable: true },
    { key: "/permission/platforms", title: "平台管理" }
  ];

  /** 处理面包屑点击 */
  const handleBreadcrumbClick = (item: BreadcrumbItem) => {
    if (item.path) {
      navigate(item.path);
    }
  };

  // --- 1. 获取平台列表 (GET) ---
  const fetchPlatforms = async () => {
    setLoading(true);
    try {
      // 根据文档 2.4.1 接口定义
      const response = await request.get("/platforms");
      const { list, total: totalCount } = response as any;
      // 同时解构 list 和 total，并分别赋予默认值
      // const { list = [], total: totalCount = 0 } = response?.data || {};

      // 数据转换：后端字段 -> 前端字段
      const formattedData = list.map((item: any) => ({
        key: item.id.toString(),
        platformName: item.name,
        authorizedPerson: item.permissionCount,
        platformUrl: item.link
      }));

      setTableData(formattedData);
      setTotal(totalCount);
    } catch (error) {
      console.error("获取平台列表失败", error);
    } finally {
      setLoading(false);
    }
  };

  // 页面加载时获取数据
  useEffect(() => {
    fetchPlatforms();
  }, [current, pageSize]);

  // --- 2. 新增平台 (POST) ---
  const handleAddPlatform = async (values: PlatformFormValues) => {
    try {
      // 根据文档 2.4.2 接口定义
      await request.post("/platforms", {
        name: values.name,
        link: values.link
      });
      console.log("平台创建成功");
      setIsFormOpen(false);
      form.resetFields();
      fetchPlatforms(); // 刷新列表
    } catch (error) {
      console.error("创建平台失败", error);
    }
  };

  // --- 3. 编辑平台 (PUT) ---
  const handleEditPlatform = async () => {
    try {
      // 根据文档 2.4.3 接口定义
      await request.put(`/platforms/${editingKey}`, {
        name: editingData.name,
        link: editingData.link
      });
      console.log("平台更新成功");
      setEditingKey("");
      fetchPlatforms(); // 刷新列表
    } catch (error) {
      console.error("更新平台失败", error);
    }
  };

  // --- 4. 删除平台 (DELETE) ---
  const handleDeletePlatform = (id: string) => {
    MhModal.confirm({
      title: "确认删除",
      content: "确定要删除该平台吗？删除后关联的权限记录也会被级联删除。",
      okText: "确认",
      cancelText: "取消",
      onOk: async () => {
        try {
          // 根据文档 2.4.4 接口定义
          await request.delete(`/platforms/${id}`);
          console.log("平台删除成功");
          fetchPlatforms(); // 刷新列表
        } catch (error) {
          console.log("删除平台失败", error);
        }
      }
    });
  };

  /** 处理编辑行点击 */
  const handleEditRow = (record: TableData) => {
    setEditingKey(record.key);
    setEditingData({
      name: record.platformName,
      link: record.platformUrl
    });
  };

  /** 处理取消编辑 */
  const handleCancelEdit = () => {
    setEditingKey("");
    setEditingData({ name: "", link: "" });
  };

  /** 表格列定义 */
  const columns = [
    {
      title: "平台名称",
      dataIndex: "platformName",
      key: "platformName",
      render: (text: string, record: TableData) =>
        editingKey === record.key ? (
          <MhInput
            value={editingData.name}
            onChange={e => setEditingData({ ...editingData, name: e.target.value })}
            style={{ width: 120 }}
            allowClear
          />
        ) : (
          <span>{text}</span>
        )
    },
    {
      title: "有权限(人)",
      dataIndex: "authorizedPerson",
      key: "authorizedPerson"
    },
    {
      title: "平台网址",
      dataIndex: "platformUrl",
      key: "platformUrl",
      render: (text: string, record: TableData) =>
        editingKey === record.key ? (
          <MhInput
            value={editingData.link}
            onChange={e => setEditingData({ ...editingData, link: e.target.value })}
            style={{ width: 150 }}
            allowClear
          />
        ) : (
          <span>{text}</span>
        )
    },
    {
      title: "操作",
      key: "action",
      render: (_: unknown, record: TableData) =>
        editingKey === record.key ? (
          <MhSpace>
            <a onClick={handleEditPlatform} style={{ color: "#1677ff" }}>
              保存
            </a>
            <a onClick={handleCancelEdit} style={{ color: "rgba(0, 0, 0, 0.45)" }}>
              取消
            </a>
          </MhSpace>
        ) : (
          <MhSpace>
            <a onClick={() => handleEditRow(record)} style={{ color: "#1677ff" }}>
              编辑
            </a>
            <a onClick={() => handleDeletePlatform(record.key)} style={{ color: "#1677ff" }}>
              删除
            </a>
          </MhSpace>
        )
    }
  ];

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "calc(100vh - 64px - 18px)", paddingBottom: 18 }}>
      {/* 页面头部 */}
      <PageHeader
        title="平台管理"
        breadcrumbItems={breadcrumbItems}
        onBreadcrumbClick={handleBreadcrumbClick}
        showFavorite={false}
      />

      {/* 内容区域 */}
      <MhCard style={{ flex: 1, display: "flex", flexDirection: "column" }}>
        <div style={{ marginBottom: 16 }}>
          <MhButton type="primary" onClick={() => setIsFormOpen(true)} style={{ marginRight: 8 }}>
            新增平台
          </MhButton>
          {/* <MhButton type="primary" onClick={() => navigate("/permission/platforms/AddPlatform")}>
            新增平台
          </MhButton> */}
        </div>

        {/* 表格区域 */}
        <MhTable
          columns={columns}
          dataSource={tableData}
          pagination={false}
          size="small"
          loading={loading}
          rowKey="key"
        />

        {/* 新增平台弹窗 */}
        <TablePopUpForm
          open={isFormOpen}
          onCancel={() => setIsFormOpen(false)}
          onSubmit={(values: TablePopUpFormValues) => {
            console.log("弹窗原始数据:", values);
            //  构造一个新的对象，符合 PlatformFormValues 的要求
            const formattedData: PlatformFormValues = {
              // 根据实际业务需求进行字段映射
              name: values.mediaName,
              link: values.description || ""
            };
            // 关闭弹窗
            setIsFormOpen(false);
            // 调用处理函数，传入转换后的数据
            handleAddPlatform(formattedData);
          }}
        />

        <div className={styles["table-base-footer"]}>
          <div className={styles["table-base-total"]}>总共 {tableData.length} 条</div>
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
    </div>
  );
};

export default PlatformManagement;
