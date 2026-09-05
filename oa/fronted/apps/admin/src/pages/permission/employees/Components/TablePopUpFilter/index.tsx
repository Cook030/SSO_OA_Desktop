import {
  MhButton,
  MhConModal,
  MhConModalCloseButton,
  MhFlex,
  MhInput,
  MhRadio,
  MhSelect,
  MhTypography
} from "@mh-repo/ui";
import type React from "react";
import { useEffect, useRef, useState } from "react";
import styles from "./index.module.less";

export interface TablePopUpFilterValues {
  mediaPackage?: "valid" | "all";
  os?: "all" | "iOS" | "Android";
  requestStatus?: "all" | "requesting" | "notRequested";
  requestCountOperator?: "gte" | "lte" | "eq";
  requestCountValue: string;
}

export interface TablePopUpFilterProps {
  open: boolean;
  title?: React.ReactNode;
  width?: number;
  initialValues?: Partial<TablePopUpFilterValues>;
  confirmText?: React.ReactNode;
  cancelText?: React.ReactNode;
  confirmLoading?: boolean;
  onCancel: () => void;
  onConfirm?: (values: TablePopUpFilterValues) => void;
}

type FilterFieldKey = "mediaPackage" | "os" | "requestStatus" | "requestCount";
type FilterSectionKey = "data" | "data1";

const DEFAULT_VALUES: TablePopUpFilterValues = {
  mediaPackage: undefined,
  os: undefined,
  requestStatus: undefined,
  requestCountOperator: undefined,
  requestCountValue: ""
};

const REQUEST_COUNT_OPTIONS = [
  { label: "大于等于", value: "gte" },
  { label: "小于等于", value: "lte" },
  { label: "等于", value: "eq" }
] as const;

const FILTER_FIELDS: Array<{ key: FilterFieldKey; label: string }> = [
  { key: "mediaPackage", label: "媒体包名" },
  { key: "os", label: "OS系统" },
  { key: "requestStatus", label: "请求状态" },
  { key: "requestCount", label: "1/1000请求数" }
];
const _FILTER_FIELDS1: Array<{ key: FilterFieldKey; label: string }> = [
  { key: "mediaPackage", label: "媒体包名" },
  { key: "os", label: "OS系统" },
  { key: "requestStatus", label: "请求状态" },
  { key: "requestCount", label: "1/1000请求数" }
];

const FILTER_SECTIONS: Array<{
  key: FilterSectionKey;
  label: string;
  fields: Array<{ key: FilterFieldKey; label: string }>;
}> = [
  {
    key: "data",
    label: "数据",
    fields: FILTER_FIELDS
  }
  // {
  //   key: "data1",
  //   label: "数据1",
  //   fields: FILTER_FIELDS1
  // }
];

const TablePopUpFilter: React.FC<TablePopUpFilterProps> = ({
  open,
  title = "更多筛选",
  width = 800,
  initialValues,
  confirmText = "确定",
  cancelText = "取消",
  confirmLoading = false,
  onCancel,
  onConfirm
}) => {
  const [activeAnchor, setActiveAnchor] = useState("");
  const [values, setValues] = useState<TablePopUpFilterValues>(DEFAULT_VALUES);
  const anchorRefs = useRef<Record<string, HTMLDivElement | null>>({});

  useEffect(() => {
    if (!open) {
      return;
    }

    setActiveAnchor("");
    setValues({
      ...DEFAULT_VALUES,
      ...initialValues
    });
  }, [initialValues, open]);

  const updateValues = (patch: Partial<TablePopUpFilterValues>) => {
    setValues(currentValues => ({
      ...currentValues,
      ...patch
    }));
  };

  const scrollToAnchor = (anchorKey: string) => {
    setActiveAnchor(anchorKey);
    anchorRefs.current[anchorKey]?.scrollIntoView({
      behavior: "smooth",
      block: "start"
    });
  };

  const renderFieldContent = (fieldKey: FilterFieldKey) => {
    if (fieldKey === "mediaPackage") {
      return (
        <MhRadio.Group
          className={styles.radioGroup}
          value={values.mediaPackage}
          onChange={event => updateValues({ mediaPackage: event.target.value })}
        >
          <MhRadio value="valid">有效</MhRadio>
          <MhRadio value="all">全部</MhRadio>
        </MhRadio.Group>
      );
    }

    if (fieldKey === "os") {
      return (
        <MhRadio.Group
          className={styles.radioGroup}
          value={values.os}
          onChange={event => updateValues({ os: event.target.value })}
        >
          <MhRadio value="all">全部</MhRadio>
          <MhRadio value="iOS">IOS</MhRadio>
          <MhRadio value="Android">Android</MhRadio>
        </MhRadio.Group>
      );
    }

    if (fieldKey === "requestStatus") {
      return (
        <MhRadio.Group
          className={styles.radioGroup}
          value={values.requestStatus}
          onChange={event => updateValues({ requestStatus: event.target.value })}
        >
          <MhRadio value="all">全部</MhRadio>
          <MhRadio value="requesting">正在请求</MhRadio>
          <MhRadio value="notRequested">未请求</MhRadio>
        </MhRadio.Group>
      );
    }

    return (
      <MhFlex gap={8} className={styles.rangeRow}>
        <MhSelect
          value={values.requestCountOperator}
          options={REQUEST_COUNT_OPTIONS.map(option => ({ ...option }))}
          className={styles.operatorSelect}
          onChange={value =>
            updateValues({
              requestCountOperator: value as TablePopUpFilterValues["requestCountOperator"]
            })
          }
        />
        <MhInput
          value={values.requestCountValue}
          placeholder="请输入"
          className={styles.valueInput}
          onChange={event => updateValues({ requestCountValue: event.target.value })}
        />
      </MhFlex>
    );
  };

  return (
    <MhConModal
      open={open}
      onCancel={onCancel}
      centered
      width={width}
      maskClosable
      title={title}
      headerExtra={<MhConModalCloseButton onClick={onCancel} />}
      footer={
        <MhFlex justify="flex-end" gap={8}>
          <MhButton onClick={onCancel}>{cancelText}</MhButton>
          <MhButton type="primary" loading={confirmLoading} onClick={() => onConfirm?.(values)}>
            {confirmText}
          </MhButton>
        </MhFlex>
      }
    >
      <div className={styles.contentCard}>
        <aside className={styles.sidebar}>
          {FILTER_SECTIONS.map(section => (
            <div key={section.key} className={styles.sidebarGroup}>
              <button
                type="button"
                className={`${styles.sidebarTitle}${activeAnchor === `section-${section.key}` ? ` ${styles.sidebarTitleActive}` : ""}`}
                onClick={() => scrollToAnchor(`section-${section.key}`)}
              >
                {section.label}
              </button>

              <div className={styles.fieldList}>
                {section.fields.map(field => {
                  const anchorKey = `field-${section.key}-${field.key}`;

                  return (
                    <button
                      key={anchorKey}
                      type="button"
                      className={`${styles.fieldButton}${activeAnchor === anchorKey ? ` ${styles.fieldButtonActive}` : ""}`}
                      onClick={() => scrollToAnchor(anchorKey)}
                    >
                      {field.label}
                    </button>
                  );
                })}
              </div>
            </div>
          ))}
        </aside>

        <section className={styles.detail}>
          {FILTER_SECTIONS.map(section => (
            <div
              key={section.key}
              ref={element => {
                anchorRefs.current[`section-${section.key}`] = element;
              }}
              className={styles.detailSection}
            >
              <MhTypography.Text className={styles.sectionTitle}>{section.label}</MhTypography.Text>

              <div className={styles.sectionBlocks}>
                {section.fields.map(field => {
                  const anchorKey = `field-${section.key}-${field.key}`;

                  return (
                    <div
                      key={anchorKey}
                      ref={element => {
                        anchorRefs.current[anchorKey] = element;
                      }}
                      className={styles.formBlock}
                    >
                      <div className={styles.blockTitle}>{field.label}</div>
                      {renderFieldContent(field.key)}
                    </div>
                  );
                })}
              </div>
            </div>
          ))}
        </section>
      </div>
    </MhConModal>
  );
};

export default TablePopUpFilter;
