import {
  MhAlert,
  MhBadge,
  MhButton,
  MhCard,
  MhDescriptions,
  MhForm,
  MhIcon,
  MhInput,
  MhMessage,
  MhSelect,
  MhSpace,
  MhSteps
} from "@mh-repo/ui";
import { CheckCircleOutlined, LeftOutlined } from "@mh-repo/ui/components/General/Icon";
import { pinyin } from "pinyin-pro";
import type React from "react";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import type { BreadcrumbItem } from "../../../../components/Breadcrumb";
import PageHeader from "../../../../components/PageHeader";
import request from "../../../../utils/request";
import styles from "./index.module.less";

const { Item: FormItem } = MhForm;

/** 步骤项 */
const steps = [{ title: "基础信息" }, { title: "平台权限" }, { title: "授权结果" }];

// 密码生成函数示例
const generateRandomPassword = (length = 12) => {
  const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()";
  let password = "";
  for (let i = 0; i < length; i++) {
    const randomIndex = Math.floor(Math.random() * charset.length);
    password += charset[randomIndex];
  }
  return password;
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

/**
 * 【模板页面】分步表单01 - 横向分步表单模板
 *
 * 适用场景：分步操作流程，横向步骤条（如：填写信息 → 确认 → 完成）
 *
 * 使用说明：
 * 1. 复制此文件作为新页面的起点
 * 2. 修改组件名称 StepForm01 -> YourPageName
 * 3. 修改页面标题和步骤标题（steps 数组）
 * 4. 修改每步的内容（renderStep1/renderStep2/renderStep3 函数）
 * 5. 调整步骤切换逻辑（handleFirstStepSubmit/handleNextStep/handlePrevStep）
 * 6. 在 handleFirstStepSubmit 中处理第一步表单提交，成功后 setCurrentStep(1)
 *
 * 关键功能：
 * - 横向步骤条：使用 MhSteps 的 type="navigation" 样式
 * - 三步流程：
 *   Step 1: 表单填写（MhForm + 表单项）
 *   Step 2: 信息确认（MhDescriptions 展示已填信息）
 *   Step 3: 完成页面（自定义成功图标 + 操作按钮）
 * - 步骤切换：通过 currentStep 状态控制显示哪一步
 * - 按钮控制：上一步/下一步/完成按钮控制流程流转
 */
const AddEmployee: React.FC = () => {
  const navigate = useNavigate();
  const [form] = MhForm.useForm();
  const [loading, setLoading] = useState(false);
  const [currentStep, setCurrentStep] = useState(0);
  const [platformOptions, setPlatformOptions] = useState<{ label: string; value: number }[]>([]);
  const [authStatusList, setAuthStatusList] = useState<{ id: number; isAuthorized: boolean }[]>([]);
  const [formData, setFormData] = useState<Record<string, any>>({});
  const [randomPassword, setRandomPassword] = useState(generateRandomPassword(12));

  /** 面包屑配置 */
  const breadcrumbItems: BreadcrumbItem[] = [
    { key: "permission", title: "权限管理", path: "/permission", clickable: true },
    { key: "permission/employees", title: "员工管理", path: "/permission/employees", clickable: true },
    { key: "/permission/employees/AddEmployee", title: "新增员工" }
  ];
  /** 处理面包屑点击 */
  const handleBreadcrumbClick = (item: BreadcrumbItem) => {
    if (item.path) {
      navigate(item.path);
    }
  };

  /** 部门选项+平台列表 */
  useEffect(() => {
    const fetchPlatformOptions = async () => {
      try {
        //平台
        const platformData = (await request.get("/platforms", { params: { page: 1, pageSize: 1000 } })) as any;
        const list = platformData.list || [];
        setPlatformOptions(list.map((p: any) => ({ label: p.name, value: p.id })));
        const initialStatus = list.map((p: any) => ({
          id: p.id,
          isAuthorized: false
        }));
        setAuthStatusList(initialStatus);
      } catch (error) {
        console.error("获取平台列表失败:", error);
      }
    };

    fetchPlatformOptions();
  }, []); // 空数组 [] 表示只在组件挂载时执行一次

  /** 处理第一步表单提交 */
  const handleFirstStepSubmit = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 1000));
      console.log("第一步表单数据:", values);
      setFormData(values); // 保存到 state
      setCurrentStep(1);
    } finally {
      setLoading(false);
    }
  };

  // 点击“授权”
  const handleAuthorize = (id: number) => {
    setAuthStatusList(prev => prev.map(item => (item.id === id ? { ...item, isAuthorized: true } : item)));
    // TODO: 在这里调用后端接口保存授权
  };

  // 点击“取消”
  const handleCancelAuth = (id: number) => {
    setAuthStatusList(prev => prev.map(item => (item.id === id ? { ...item, isAuthorized: false } : item)));
    // TODO: 在这里调用后端接口取消授权
  };

  /** 处理复制信息 */
  const handleCopyInfo = async () => {
    const account = formData.email || "";
    const password = formData.password || "Mhint@123";
    // const copyText = `账号:${account}@maplehaze.cn 密码:${password}`;
    const copyText = `账号:${account} 密码:${password}`;

    try {
      await navigator.clipboard.writeText(copyText);
      MhMessage.success("复制成功！");
    } catch (err) {
      console.error("复制失败:", err);
      MhMessage.error("复制失败，请手动复制");
    }
  };

  /** 处理提交 */
  const submitEmployee = async () => {
    try {
      // const values = form.getFieldsValue();
      // 获取已授权的平台 ID 列表
      const platformIds = authStatusList.filter(item => item.isAuthorized).map(item => item.id);
      const payload = {
        name: formData.name?.trim() || "",
        phone: formData.phone || "",
        emailPrefix: formData.email || "",
        account: formData.account || "",
        department: formData.department || "",
        platformIds: platformIds,
        password: formData.password || ""
      };
      await request.post("/employees", payload);
      MhMessage.success("新增员工成功");
      return true;
    } catch (error: any) {
      MhMessage.error(error.message || "新增员工失败");
      return false;
    }
  };

  /** 处理上一步 */
  const handlePrevStep = () => {
    setCurrentStep(currentStep - 1);
  };

  const handleNextStep = async () => {
    setLoading(true);
    const success = await submitEmployee();
    setLoading(false);
    if (success) {
      setCurrentStep(2);
    }
  };

  // 重置所有平台为未授权
  const resetToFirstStep = () => {
    const newPassword = generateRandomPassword(12);
    setRandomPassword(newPassword);
    form.resetFields();
    form.setFieldsValue({
      password: newPassword
    });
    setFormData({});
    setAuthStatusList(platformOptions.map(p => ({ id: p.value, isAuthorized: false })));
    setCurrentStep(0);
    setRandomPassword(generateRandomPassword(12)); // 生成新的随机密码
  };

  /** 渲染第一步内容 */
  const renderStep1 = () => {
    // 2. 处理姓名输入，自动转换为拼音并填入登录账号
    const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      if (value) {
        // 将汉字转为拼音，不带声调，首字母大写（可选）
        const pinyinValue = pinyin(value, { toneType: "none", type: "array" }).join("");
        form.setFieldsValue({
          username: pinyinValue,
          email: pinyinValue
        });
      } else {
        form.setFieldsValue({ username: "", email: "" });
      }
    };

    return (
      <MhForm
        form={form}
        layout="vertical"
        initialValues={{
          password: randomPassword
        }}
        onFinish={handleFirstStepSubmit}
        autoComplete="off"
        style={{ maxWidth: 328, margin: "0 auto" }}
      >
        {/* 员工姓名 */}
        <MhForm.Item
          label={
            <span className={styles.requiredLabel}>
              <span className={styles.requiredMark}>*</span>
              <span>员工姓名</span>
            </span>
          }
          name="name"
          required={false}
          rules={[{ required: true, message: "请输入员工姓名" }]}
        >
          {/* 3. 添加 onChange 监听 */}
          <MhInput placeholder="请输入用户名" allowClear onChange={handleNameChange} />
        </MhForm.Item>

        {/* 媒体/所属部门 */}
        <FormItem
          label={
            <span className={styles.requiredLabel}>
              <span className={styles.requiredMark}>*</span>
              <span>所属部门</span>
            </span>
          }
          name="department"
          required={false}
          rules={[{ required: true, message: "请选择所属部门" }]}
          style={{ marginBottom: 24 }}
        >
          <MhSelect placeholder="请选择" options={DEPARTMENT_OPTIONS} allowClear />
        </FormItem>

        {/* 登录账号 */}
        <MhForm.Item
          label={
            <span className={styles.requiredLabel}>
              <span className={styles.requiredMark}>*</span>
              <span>登录账号</span>
            </span>
          }
          name="account"
          required={false}
          rules={[{ required: true, message: "请输入登录账号" }]}
        >
          <MhInput placeholder="请输入登录账号" allowClear />
        </MhForm.Item>

        {/* 登录密码 */}
        <MhForm.Item
          label={
            <span className={styles.requiredLabel}>
              <span className={styles.requiredMark}>*</span>
              <span>登录密码</span>
            </span>
          }
          name="password"
          required={false}
          rules={[{ required: true, message: "请输入登录密码" }]}
        >
          <MhInput allowClear />
        </MhForm.Item>

        {/* 手机号 */}
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
              <span>邮箱地址</span>
            </span>
          }
          name="email"
          required={false}
        >
          <MhInput
            placeholder="请输入"
            allowClear
            addonAfter={<span style={{ color: "rgba(0,0,0,0.45)" }}>@maplehaze.cn</span>}
          />
        </MhForm.Item>

        {/* 下一步按钮 */}
        <FormItem style={{ marginBottom: 0 }}>
          <MhButton type="primary" htmlType="submit" loading={loading}>
            下一步
          </MhButton>
        </FormItem>
      </MhForm>
    );
  };

  /** 渲染第二步内容 */
  const renderStep2 = () => (
    <div>
      <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
        {authStatusList.map(item => {
          // 找到对应的平台名称
          const platformInfo = platformOptions.find(p => p.value === item.id);
          const platformName = platformInfo ? platformInfo.label : "未知平台";

          return (
            <div
              key={item.id}
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                padding: "16px",
                backgroundColor: "#fafafa", // 浅灰背景
                borderRadius: "4px"
              }}
            >
              {/* --- 左侧：标题 + 状态图标 --- */}
              <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                {/* 只有已授权时显示蓝色对勾图标 */}
                {item.isAuthorized && (
                  <span style={{ color: "#1890ff", fontSize: "16px" }}>
                    <CheckCircleOutlined />
                  </span>
                )}
                <span style={{ fontWeight: 500 }}>{platformName}</span>
              </div>

              {/* --- 右侧：操作按钮 --- */}
              <div style={{ display: "flex", gap: "12px" }}>
                {/* 授权按钮 */}
                <MhButton
                  type="primary"
                  ghost={!item.isAuthorized} // 未授权时用幽灵按钮(空心)，已授权用实心
                  style={{
                    backgroundColor: item.isAuthorized ? "#e7f4ff" : "#ffffff", // 已授权时的浅蓝背景
                    borderColor: item.isAuthorized ? "#e7f4ff" : "#8fb4fd", // 边框颜色
                    color: item.isAuthorized ? "#1890ff" : "#1890ff" // 文字颜色
                    // cursor: item.isAuthorized ? "default" : "pointer"
                  }}
                  disabled={item.isAuthorized} // 已授权则禁用“授权”按钮
                  onClick={() => handleAuthorize(item.id)}
                >
                  {item.isAuthorized ? "已授权" : "授权"}
                </MhButton>

                {/* 取消按钮 */}
                <MhButton
                  ghost={!item.isAuthorized} // 未授权时用幽灵按钮(空心)，已授权用实心
                  style={{
                    backgroundColor: item.isAuthorized ? "#ffffff" : "#f0f0f0", // 已授权时的浅蓝背景
                    borderColor: item.isAuthorized ? "#91b6fc" : "#d9d9d9", // 边框颜色
                    color: item.isAuthorized ? "#5d91fd" : "#b4b4b4" // 文字颜色
                    // cursor: item.isAuthorized ? "default" : "pointer"
                  }}
                  disabled={!item.isAuthorized} // 未授权则禁用“取消”按钮
                  onClick={() => handleCancelAuth(item.id)}
                >
                  取消
                </MhButton>
              </div>
            </div>
          );
        })}
      </div>
      {/* 按钮组 */}
      <div style={{ display: "flex", gap: 12, marginTop: 26 }}>
        <MhButton onClick={handlePrevStep}>上一步</MhButton>
        <MhButton type="primary" onClick={handleNextStep} loading={loading}>
          下一步
        </MhButton>
      </div>
    </div>
  );

  /** 渲染第三步内容 */
  const renderStep3 = () => (
    <div style={{ textAlign: "center", padding: "120px 0" }}>
      {/* 完成图标 */}
      <div
        style={{
          width: 72,
          height: 72,
          margin: "0 auto 24px",
          borderRadius: "50%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center"
        }}
      >
        <svg width="72" height="72" viewBox="0 0 40 40" fill="none">
          <circle cx="20" cy="20" r="20" fill="#52C41A" />
          <path d="M12 20L17 25L28 14" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>

      {/* 标题 */}
      <div style={{ fontSize: 20, color: "rgba(0, 0, 0, 0.85)", marginBottom: 8, fontWeight: 400 }}>完成</div>

      {/* 描述文字 */}
      <div style={{ fontSize: 14, color: "rgba(0, 0, 0, 0.45)", marginBottom: 24 }}>
        {/* <span style={{ marginRight: 16 }}>账号:{formData.email || ""}@maplehaze.cn</span> */}
        <span style={{ marginRight: 16 }}>账号:{formData.email || ""}</span>
        <span>密码:{formData.password || "Mhint@123"}</span>
      </div>

      {/* 按钮组 */}
      <div style={{ display: "flex", gap: 12, justifyContent: "center" }}>
        <MhButton onClick={handleCopyInfo}>复制信息</MhButton>
        <MhButton type="primary" onClick={resetToFirstStep}>
          继续新建
        </MhButton>
      </div>
    </div>
  );

  return (
    <div>
      {/* 页面头部 */}
      <PageHeader
        title={
          (
            <>
              <LeftOutlined style={{ marginRight: 8, cursor: "pointer" }} onClick={() => navigate(-1)} />
              新增员工
            </>
          ) as any
        }
        breadcrumbItems={breadcrumbItems}
        onBreadcrumbClick={handleBreadcrumbClick}
        showFavorite={false}
      />

      {/* 表单区域 */}
      <MhCard className="card_minHeight">
        {/* 步骤条 */}
        <div style={{ marginBottom: 24 }}>
          <MhSteps current={currentStep} items={steps} type="navigation" />
        </div>

        {/* 根据当前步骤渲染不同内容 */}
        {currentStep === 0 && renderStep1()}
        {currentStep === 1 && renderStep2()}
        {currentStep === 2 && renderStep3()}
      </MhCard>
    </div>
  );
};

export default AddEmployee;
