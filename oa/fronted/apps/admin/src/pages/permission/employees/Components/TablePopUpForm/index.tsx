import { MhButton, MhConModal, MhConModalCloseButton, MhFlex, MhForm, MhInput, MhRadio } from "@mh-repo/ui";
import type React from "react";
import { useEffect } from "react";
import styles from "./index.module.less";

export interface TablePopUpFormValues {
  mediaName: string;
  description?: string;
  os: "Android" | "iOS";
  aggregateType: "packageName" | "tagid";
}

export interface TablePopUpFormProps {
  open: boolean;
  title?: React.ReactNode;
  form?: ReturnType<typeof MhForm.useForm<TablePopUpFormValues>>[0];
  initialValues?: Partial<TablePopUpFormValues>;
  confirmText?: React.ReactNode;
  cancelText?: React.ReactNode;
  confirmLoading?: boolean;
  width?: number;
  onCancel: () => void;
  onSubmit?: (values: TablePopUpFormValues) => void | Promise<void>;
}

const DEFAULT_VALUES: TablePopUpFormValues = {
  mediaName: "",
  description: "",
  os: "Android",
  aggregateType: "packageName"
};

const TablePopUpForm: React.FC<TablePopUpFormProps> = ({
  open,
  title = "新增平台",
  form,
  initialValues,
  confirmText = "确定",
  cancelText = "取消",
  confirmLoading = false,
  width = 600,
  onCancel,
  onSubmit
}) => {
  const [innerForm] = MhForm.useForm<TablePopUpFormValues>();
  const currentForm = form ?? innerForm;

  useEffect(() => {
    if (!open) {
      return;
    }

    currentForm.setFieldsValue({
      ...DEFAULT_VALUES,
      ...initialValues
    });
  }, [currentForm, initialValues, open]);

  return (
    <MhConModal
      open={open}
      onCancel={onCancel}
      centered
      width={width}
      title={title}
      headerExtra={<MhConModalCloseButton onClick={onCancel} />}
      footer={
        <MhFlex justify="flex-end" gap={8}>
          <MhButton onClick={onCancel}>{cancelText}</MhButton>
          <MhButton type="primary" loading={confirmLoading} onClick={() => currentForm.submit()}>
            {confirmText}
          </MhButton>
        </MhFlex>
      }
    >
      <MhForm
        form={currentForm}
        layout="vertical"
        className={styles.form}
        initialValues={{ ...DEFAULT_VALUES, ...initialValues }}
        onFinish={values => onSubmit?.(values)}
        requiredMark={false}
      >
        <MhForm.Item
          label={
            <span className={styles.requiredLabel}>
              <span className={styles.requiredMark}>*</span>
              <span>媒体名称</span>
            </span>
          }
          name="mediaName"
          rules={[{ required: true, message: "请输入媒体名称" }]}
          className={styles.formItem}
        >
          <MhInput placeholder="给项目起个名字" maxLength={50} />
        </MhForm.Item>

        <MhForm.Item
          label={
            <span className={styles.requiredLabel}>
              <span className={styles.requiredMark}>*</span>
              <span>平台网址</span>
            </span>
          }
          name="description"
          className={styles.formItem}
          rules={[{ required: true, message: "请输入平台网址" }]}
        >
          <MhInput.TextArea placeholder="请输入内容" rows={3} allowClear style={{ resize: "none" }} />
        </MhForm.Item>
      </MhForm>
    </MhConModal>
  );
};

export default TablePopUpForm;
