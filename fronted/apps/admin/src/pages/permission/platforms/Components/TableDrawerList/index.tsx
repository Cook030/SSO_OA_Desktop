import {
  MhButton,
  MhCheckbox,
  MhDrawer,
  MhFlex,
  MhForm,
  MhInput,
  MhInputNumber,
  MhRadio,
  MhSelect,
  MhTabs,
  MhTypography
} from "@mh-repo/ui";
import type React from "react";
import { useEffect, useState } from "react";
import SvgComp from "../../../../../components/SvgComp";
import styles from "./index.module.less";

export interface TableDrawerListSummaryItem {
  key?: string;
  label: React.ReactNode;
  value: React.ReactNode;
}

export interface TableDrawerListValues {
  styleId: string;
  mediaTypes: string[];
  description?: string;
  adSlotId?: string;
  weight?: number;
  targetOpenType: "public" | "partial" | "private";
}

export interface TableDrawerListProps {
  open: boolean;
  title?: React.ReactNode;
  width?: number;
  form?: ReturnType<typeof MhForm.useForm<TableDrawerListValues>>[0];
  initialValues?: Partial<TableDrawerListValues>;
  activeTabKey?: string;
  onTabChange?: (key: string) => void;
  summaryItems?: TableDrawerListSummaryItem[];
  confirmText?: React.ReactNode;
  cancelText?: React.ReactNode;
  confirmLoading?: boolean;
  onCancel: () => void;
  onSubmit?: (values: TableDrawerListValues) => void | Promise<void>;
}

const DEFAULT_VALUES: TableDrawerListValues = {
  styleId: "",
  mediaTypes: [],
  description: "",
  adSlotId: "",
  weight: 0,
  targetOpenType: "public"
};

const DEFAULT_SUMMARY_ITEMS: TableDrawerListSummaryItem[] = [
  { label: "某个分类", value: "" },
  { label: "某某某某：", value: "这是一条信息内容可是是文案..." },
  { label: "某某某某：", value: "这是一条信息内容可是是文案..." },
  { label: "某某某某：", value: "这是一条信息内容" },
  { label: "某某某某：", value: "这是一条信息内容" }
];

const STYLE_OPTIONS = [
  { label: "Please Select", value: "" },
  { label: "样式A", value: "style-a" },
  { label: "样式B", value: "style-b" }
];

const AD_SLOT_OPTIONS = [
  { label: "请选择", value: "" },
  { label: "开屏广告位", value: "开屏广告位" },
  { label: "信息流广告位", value: "信息流广告位" }
];

/** 媒体类型选项（不含"全部"） */
const MEDIA_TYPE_OPTIONS = [
  { label: "原生应用", value: "native_app" },
  { label: "原生游戏", value: "native_game" },
  { label: "快应用", value: "quick_app" },
  { label: "快游戏", value: "quick_game" },
  { label: "其他", value: "other" }
];

const ALL_VALUE = "all";

const getAllMediaValues = () => MEDIA_TYPE_OPTIONS.map(opt => opt.value);

const SummaryGrid: React.FC<{ items: TableDrawerListSummaryItem[] }> = ({ items }) => {
  const [, ...contentItems] = items;

  return (
    <div className={styles.summaryGrid}>
      {contentItems.map((item, index) => (
        <div key={item.key ?? `summary-${index}`} className={styles.summaryItem}>
          <span className={styles.summaryLabel}>{item.label}</span>
          <span className={styles.summaryValue}>{item.value}</span>
        </div>
      ))}
    </div>
  );
};

const TableDrawerList: React.FC<TableDrawerListProps> = ({
  open,
  title = "某某某新建",
  width = 736,
  form,
  initialValues,
  activeTabKey = "basic",
  onTabChange,
  summaryItems = DEFAULT_SUMMARY_ITEMS,
  confirmText = "确定",
  cancelText = "取消",
  confirmLoading = false,
  onCancel,
  onSubmit
}) => {
  const [innerForm] = MhForm.useForm<TableDrawerListValues>();
  const currentForm = form ?? innerForm;
  const [indeterminate, setIndeterminate] = useState(false);
  const [checkAll, setCheckAll] = useState(false);

  const updateCheckAllStatus = (checkedList: string[]) => {
    const allValues = getAllMediaValues();
    const checkedMediaValues = checkedList.filter(value => value !== ALL_VALUE);
    const allChecked = checkedMediaValues.length === allValues.length && allValues.length > 0;
    const someChecked = checkedMediaValues.length > 0 && checkedMediaValues.length < allValues.length;

    setCheckAll(allChecked);
    setIndeterminate(someChecked);
  };

  const handleMediaTypeChange = (checkedValues: (string | number)[]) => {
    const values = checkedValues as string[];
    updateCheckAllStatus(values);
    currentForm.setFieldsValue({ mediaTypes: values });
  };

  const onCheckAllChange = () => {
    const allValues = getAllMediaValues();
    const currentValues: string[] = currentForm.getFieldValue("mediaTypes") || [];
    const isAllChecked = currentValues.length === allValues.length && allValues.length > 0;

    if (isAllChecked) {
      setCheckAll(false);
      setIndeterminate(false);
      currentForm.setFieldsValue({ mediaTypes: [] });
      return;
    }

    setCheckAll(true);
    setIndeterminate(false);
    currentForm.setFieldsValue({ mediaTypes: allValues });
  };

  useEffect(() => {
    if (!open) {
      return;
    }

    const values = {
      ...DEFAULT_VALUES,
      ...initialValues
    };

    currentForm.setFieldsValue(values);
    updateCheckAllStatus(values.mediaTypes || []);
  }, [currentForm, initialValues, open]);

  return (
    <MhDrawer
      open={open}
      size={width}
      placement="right"
      onClose={onCancel}
      closable={false}
      maskClosable
      className={styles.drawer}
      rootClassName={styles.drawerRoot}
      footer={null}
    >
      <div className={styles.container}>
        <div className={styles.header}>
          <div className={styles.headerTitle}>{title}</div>
          <MhButton type="text" className={styles.closeButton} onClick={onCancel} aria-label="关闭">
            <SvgComp name="CloseOutlined" />
          </MhButton>
        </div>

        <div className={styles.body}>
          {summaryItems.length > 0 ? (
            <section className={styles.summarySection}>
              <div className={styles.summaryTitle}>{summaryItems[0]?.label}</div>
              <SummaryGrid items={summaryItems} />
            </section>
          ) : null}

          <MhTabs
            defaultActiveKey={activeTabKey}
            onChange={onTabChange}
            className={styles.tabs}
            items={[
              {
                key: "basic",
                label: "基础内容",
                children: (
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
                          <span>图片styleid</span>
                        </span>
                      }
                      name="styleId"
                      className={styles.formItem}
                      rules={[{ required: true, message: "请选择图片styleid" }]}
                    >
                      <MhSelect options={STYLE_OPTIONS} disabled />
                    </MhForm.Item>

                    <MhForm.Item label="媒体类型" className={styles.formItem}>
                      <div className={styles.mediaTypeGroup}>
                        <MhCheckbox indeterminate={indeterminate} onClick={onCheckAllChange} checked={checkAll}>
                          全部
                        </MhCheckbox>
                        <span className={styles.mediaTypeDivider} />
                        <MhForm.Item name="mediaTypes" noStyle>
                          <MhCheckbox.Group
                            options={MEDIA_TYPE_OPTIONS}
                            onChange={handleMediaTypeChange}
                            className={styles.mediaTypeOptions}
                          />
                        </MhForm.Item>
                      </div>
                    </MhForm.Item>

                    <MhForm.Item
                      label="描述文字"
                      name="description"
                      className={styles.formItem}
                      rules={[{ max: 200, message: "最多200个字符" }]}
                    >
                      <MhInput.TextArea
                        placeholder="请输入内容"
                        rows={3}
                        maxLength={200}
                        showCount
                        allowClear
                        className={styles.descriptionTextArea}
                      />
                    </MhForm.Item>

                    <MhForm.Item label="广告位选择" name="adSlotId" className={styles.formItem}>
                      <MhSelect showSearch options={AD_SLOT_OPTIONS} allowClear />
                    </MhForm.Item>

                    <MhForm.Item label="权重" name="weight" className={styles.formItem}>
                      <MhInputNumber min={0} max={100} placeholder="0%" className={styles.inputNumber} />
                    </MhForm.Item>

                    <MhForm.Item label="目标公开" name="targetOpenType" className={styles.formItem}>
                      <MhRadio.Group className={styles.radioGroup}>
                        <MhRadio value="public">公开</MhRadio>
                        <MhRadio value="partial">部分公开</MhRadio>
                        <MhRadio value="private">不公开</MhRadio>
                      </MhRadio.Group>
                    </MhForm.Item>
                  </MhForm>
                )
              },
              {
                key: "data",
                label: "数据内容",
                children: (
                  <div className={styles.placeholder}>
                    <MhTypography.Text className={styles.placeholderText}>数据内容待补充</MhTypography.Text>
                  </div>
                )
              }
            ]}
          />
        </div>

        <div className={styles.footer}>
          <MhFlex justify="flex-end" gap={8}>
            <MhButton onClick={onCancel}>{cancelText}</MhButton>
            <MhButton type="primary" loading={confirmLoading} onClick={() => currentForm.submit()}>
              {confirmText}
            </MhButton>
          </MhFlex>
        </div>
      </div>
    </MhDrawer>
  );
};

export default TableDrawerList;
