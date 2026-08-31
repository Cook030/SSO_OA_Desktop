import type { MhTableProps } from "@mh-repo/ui";
import {
  MhButton,
  MhConModal,
  MhConModalCloseButton,
  MhFlex,
  MhInput,
  MhPagination,
  MhSelect,
  MhSpace,
  MhTable
} from "@mh-repo/ui";
import type React from "react";
import SvgComp from "../../../../../components/SvgComp";
import styles from "./index.module.less";

export interface TablePopUpListRow {
  key: string;
  mediaName: React.ReactNode;
  mediaId?: React.ReactNode;
  description: React.ReactNode;
  aggregateType: React.ReactNode;
  aggregateValue: React.ReactNode;
  actions?: React.ReactNode;
}

export interface TablePopUpListProps {
  open: boolean;
  title?: React.ReactNode;
  searchValue?: string;
  searchPlaceholder?: string;
  selectValue?: string;
  selectOptions?: Array<{ label: React.ReactNode; value: string }>;
  moreFilterText?: React.ReactNode;
  dataSource?: TablePopUpListRow[];
  total?: number;
  current?: number;
  pageSize?: number;
  width?: number;
  confirmLoading?: boolean;
  confirmText?: React.ReactNode;
  cancelText?: React.ReactNode;
  columns?: MhTableProps<TablePopUpListRow>["columns"];
  /** 工具栏“更多筛选”按钮之后的额外内容（例如“已应用筛选条件” chips） */
  extraToolbar?: React.ReactNode;
  onSearchChange?: (value: string) => void;
  onSelectChange?: (value: string) => void;
  onMoreFilter?: () => void;
  onPageChange?: (page: number, pageSize: number) => void;
  onConfirm?: () => void;
  onCancel: () => void;
}

const DEFAULT_OPTIONS = [{ label: "某某某", value: "某某某" }];

const DEFAULT_DATA: TablePopUpListRow[] = [
  {
    key: "1",
    mediaName: "七猫APP",
    mediaId: "09454350945435",
    description: "这里可以是一段描述文字...",
    aggregateType: "包名",
    aggregateValue: "包名",
    actions: (
      <MhSpace size={16}>
        <MhButton type="link">广告位</MhButton>
        <MhButton type="link">编辑</MhButton>
        <MhButton type="link" danger className={styles.dangerAction}>
          删除
        </MhButton>
      </MhSpace>
    )
  },
  {
    key: "2",
    mediaName: "七猫APP",
    mediaId: "0945435",
    description: "这里可以是一段描述文字...",
    aggregateType: "tagid",
    aggregateValue: "tagid",
    actions: (
      <MhSpace size={16}>
        <MhButton type="link">广告位</MhButton>
        <MhButton type="link">编辑</MhButton>
        <MhButton type="link" danger className={styles.dangerAction}>
          删除
        </MhButton>
      </MhSpace>
    )
  }
];

const TablePopUpList: React.FC<TablePopUpListProps> = ({
  open,
  title = "显示信息弹窗",
  searchValue,
  searchPlaceholder = "搜索",
  selectValue = DEFAULT_OPTIONS[0].value,
  selectOptions = DEFAULT_OPTIONS,
  moreFilterText = "更多筛选",
  dataSource = DEFAULT_DATA,
  total = dataSource.length,
  current = 1,
  pageSize = 20,
  width = 912,
  confirmLoading = false,
  confirmText = "确定",
  cancelText = "取消",
  columns,
  extraToolbar,
  onSearchChange,
  onSelectChange,
  onMoreFilter,
  onPageChange,
  onConfirm,
  onCancel
}) => {
  const internalColumns: MhTableProps<TablePopUpListRow>["columns"] = columns ?? [
    {
      title: "媒体名称",
      dataIndex: "mediaName",
      key: "mediaName",
      width: 170,
      fixed: "start",
      render: (_: React.ReactNode, record: TablePopUpListRow) => (
        <div className={styles.mediaCell}>
          <div className={styles.mediaName}>{record.mediaName}</div>
          {record.mediaId ? <div className={styles.mediaId}>{record.mediaId}</div> : null}
        </div>
      )
    },
    {
      title: "描述",
      dataIndex: "description",
      key: "description",
      width: 220,
      ellipsis: true
    },
    {
      title: "聚合类型",
      dataIndex: "aggregateType",
      key: "aggregateType",
      width: 240,
      ellipsis: true
    },
    {
      title: "聚合类型",
      dataIndex: "aggregateValue",
      key: "aggregateValue",
      width: 240,
      ellipsis: true
    },
    {
      title: "操作",
      fixed: "end",
      dataIndex: "actions",
      key: "actions",
      width: 130,
      render: (value: React.ReactNode) => <div className={styles.actions}>{value}</div>
    }
  ];

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
          <MhButton type="primary" loading={confirmLoading} onClick={onConfirm ?? onCancel}>
            {confirmText}
          </MhButton>
        </MhFlex>
      }
    >
      <div className={styles.toolbar}>
        <MhInput
          value={searchValue}
          placeholder={searchPlaceholder}
          className={styles.search}
          onChange={event => onSearchChange?.(event.target.value)}
          suffix={<SvgComp name="SearchOutlined" className={styles.searchIcon} />}
        />

        <MhSelect
          value={selectValue}
          className={styles.select}
          prefix="固定筛选项："
          options={selectOptions}
          onChange={value => onSelectChange?.(String(value))}
        />

        <MhButton className={styles.filterButton} onClick={onMoreFilter}>
          <span className={styles.filterIcon}>
            <SvgComp name="mh-filter" />
          </span>
          {moreFilterText}
        </MhButton>

        {extraToolbar}
      </div>

      <div className={styles.tableWrap}>
        <MhTable<TablePopUpListRow>
          className={styles.table}
          rowKey="key"
          columns={internalColumns}
          dataSource={dataSource}
          pagination={false}
          scroll={{ x: "max-content" }}
        />
      </div>

      <div className={styles.paginationBar}>
        <div className={styles.total}>总共 {total} 条</div>
        <MhPagination
          size="small"
          current={current}
          pageSize={pageSize}
          total={total}
          showSizeChanger
          pageSizeOptions={[10, 20, 50, 100]}
          onChange={(page, size) => onPageChange?.(page, size)}
          showTotal={undefined}
        />
      </div>
    </MhConModal>
  );
};

export default TablePopUpList;
