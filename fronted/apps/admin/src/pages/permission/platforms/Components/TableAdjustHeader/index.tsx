import {
  closestCenter,
  DndContext,
  type DragEndEvent,
  type DragOverEvent,
  DragOverlay,
  type DragStartEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors
} from "@dnd-kit/core";
import {
  defaultAnimateLayoutChanges,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { MhButton, MhCheckbox, MhFlex, MhTypography } from "@mh-repo/ui";
import React, { useCallback, useMemo } from "react";
import SvgComp from "../../../../../components/SvgComp";
import styles from "./index.module.less";

export interface TableAdjustHeaderItem {
  key: string;
  label: string;
  checked: boolean;
  disabled?: boolean;
  fixed?: "left" | "right";
}

export interface TableAdjustHeaderProps {
  items: TableAdjustHeaderItem[];
  shouldPersist?: boolean;
  onToggle: (key: string, checked: boolean) => void;
  onToggleAll: (checked: boolean) => void;
  onReset: () => void;
  onMove: (key: string, fixed?: "left" | "right") => void;
  onReorder: (sourceKey: string, targetKey: string, fixed?: "left" | "right") => void;
}

type SectionKey = "left" | "right" | "none";
type FixedType = "left" | "right" | undefined;

const SECTION_META: Array<{ key: SectionKey; title: string }> = [
  { key: "left", title: "固定在左侧" },
  { key: "right", title: "固定在右侧" },
  { key: "none", title: "不固定" }
];

/** 通用 className 拼接工具 */
const cx = (...classes: Array<string | false | null | undefined>) => classes.filter(Boolean).join(" ");

/** 阻止事件冒泡和默认行为 */
const preventEvent = (event: React.MouseEvent) => {
  event.preventDefault();
  event.stopPropagation();
};

/** 获取固定状态对应的 section key */
const getSectionKey = (fixed?: FixedType): SectionKey => fixed ?? "none";

/** 根据当前固定状态获取可用操作 */
const getMoveActions = (fixed?: FixedType): Array<{ icon: string; target: FixedType; label: string }> => {
  if (!fixed) {
    return [
      { icon: "mh-table-totop", target: "left", label: "固定在左侧" },
      { icon: "mh-table-todown", target: "right", label: "固定在右侧" }
    ];
  }
  if (fixed === "left") {
    return [
      { icon: "mh-table-topmiddle", target: undefined, label: "取消固定" },
      { icon: "mh-table-todown", target: "right", label: "固定在右侧" }
    ];
  }
  return [
    { icon: "mh-table-totop", target: "left", label: "固定在左侧" },
    { icon: "mh-table-topmiddle", target: undefined, label: "取消固定" }
  ];
};

const DragIcon: React.FC<{
  disabled?: boolean;
  attributes?: Record<string, any>;
  listeners?: Record<string, any>;
  setActivatorNodeRef?: (element: HTMLButtonElement | null) => void;
}> = ({ disabled = false, attributes, listeners, setActivatorNodeRef }) => (
  <button
    type="button"
    ref={setActivatorNodeRef}
    className={cx(styles.dragButton, disabled && styles["dragButton-disabled"])}
    aria-label="拖拽调整顺序"
    disabled={disabled}
    onClick={preventEvent}
    {...attributes}
    {...listeners}
  >
    <span className={cx(styles.dragIcon, disabled && styles["dragIcon-disabled"])} aria-hidden="true">
      {Array.from({ length: 6 }).map((_, index) => (
        <span key={index} className={styles.dragDot} />
      ))}
    </span>
  </button>
);

interface BaseRowProps {
  item: TableAdjustHeaderItem;
  className?: string;
  style?: React.CSSProperties;
  setNodeRef?: (element: HTMLElement | null) => void;
  dragIconProps?: React.ComponentProps<typeof DragIcon>;
  onToggle: (key: string, checked: boolean) => void;
  onMove: (key: string, fixed?: "left" | "right") => void;
}

const BaseRow: React.FC<BaseRowProps> = ({ item, className, style, setNodeRef, dragIconProps, onToggle, onMove }) => {
  const handleToggle = useCallback((checked: boolean) => onToggle(item.key, checked), [onToggle, item.key]);

  return (
    <label ref={setNodeRef} style={style} className={className} aria-disabled={item.disabled}>
      <div className={styles.rowMain}>
        <DragIcon disabled={item.disabled} {...dragIconProps} />
        <MhCheckbox
          checked={item.checked}
          disabled={item.disabled}
          onChange={event => handleToggle(event.target.checked)}
        >
          <span className={styles.itemLabel}>{item.label}</span>
        </MhCheckbox>
      </div>
      <span className={styles.rowAction}>
        <RowActionIcons itemKey={item.key} fixed={item.fixed} onMove={onMove} />
      </span>
    </label>
  );
};

interface SortableRowProps {
  item: TableAdjustHeaderItem;
  sectionKey: SectionKey;
  isDragOver: boolean;
  onToggle: (key: string, checked: boolean) => void;
  onMove: (key: string, fixed?: "left" | "right") => void;
}

const SortableRow: React.FC<SortableRowProps> = ({ item, sectionKey, isDragOver, onToggle, onMove }) => {
  const { attributes, listeners, setActivatorNodeRef, setNodeRef, transform, transition, isDragging } = useSortable({
    id: item.key,
    disabled: item.disabled,
    animateLayoutChanges: args => defaultAnimateLayoutChanges({ ...args, wasDragging: true }),
    data: { sectionKey }
  });

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition: transition ?? "transform 220ms cubic-bezier(0.22, 1, 0.36, 1)"
  };

  const rowClasses = cx(
    styles.row,
    item.disabled && styles["row-disabled"],
    item.fixed && styles["row-fixed"],
    isDragOver && styles["row-drag-over"],
    isDragging && styles["row-dragging"]
  );

  const dragIconProps = useMemo(
    () => ({
      disabled: item.disabled,
      attributes,
      listeners,
      setActivatorNodeRef
    }),
    [item.disabled, attributes, listeners, setActivatorNodeRef]
  );

  return (
    <BaseRow
      item={item}
      className={rowClasses}
      style={style}
      setNodeRef={setNodeRef}
      dragIconProps={dragIconProps}
      onToggle={onToggle}
      onMove={onMove}
    />
  );
};

interface RowActionIconsProps {
  itemKey: string;
  fixed?: "left" | "right";
  onMove: (key: string, fixed?: "left" | "right") => void;
}

const RowActionIcons: React.FC<RowActionIconsProps> = ({ itemKey, fixed, onMove }) => {
  const actions = useMemo(() => getMoveActions(fixed), [fixed]);

  const handleMove = useCallback(
    (target: FixedType) => (event: React.MouseEvent) => {
      preventEvent(event);
      onMove(itemKey, target);
    },
    [onMove, itemKey]
  );

  return (
    <>
      {actions.map(action => (
        <button
          key={`${itemKey}-${action.label}`}
          type="button"
          className={cx(styles.actionIcon, styles["actionIcon-active"])}
          aria-label={action.label}
          onClick={handleMove(action.target)}
        >
          <SvgComp name={action.icon} className={styles.actionIconImage} />
        </button>
      ))}
    </>
  );
};

const TableAdjustHeader: React.FC<TableAdjustHeaderProps> = ({
  items,
  shouldPersist = false,
  onToggle,
  onToggleAll,
  onReset,
  onMove,
  onReorder
}) => {
  const [draggingKey, setDraggingKey] = React.useState<string | null>(null);
  const [dragOverKey, setDragOverKey] = React.useState<string | null>(null);

  const { allChecked, indeterminate } = useMemo(() => {
    const enabled = items.filter(item => !item.disabled);
    const checkedCount = enabled.filter(item => item.checked).length;
    return {
      allChecked: enabled.length > 0 && checkedCount === enabled.length,
      indeterminate: checkedCount > 0 && checkedCount < enabled.length
    };
  }, [items]);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  );

  const sections = useMemo(() => {
    const itemsByFixed = new Map<SectionKey, TableAdjustHeaderItem[]>();

    for (const item of items) {
      const key = getSectionKey(item.fixed);
      const list = itemsByFixed.get(key) ?? [];
      list.push(item);
      itemsByFixed.set(key, list);
    }

    return SECTION_META.map(section => ({
      ...section,
      items: itemsByFixed.get(section.key) ?? []
    })).filter(section => section.items.length > 0);
  }, [items]);

  const showSectionTitle = sections.length > 1;
  const draggingItem = draggingKey ? (items.find(item => item.key === draggingKey) ?? null) : null;

  const handleDragStart = useCallback(({ active }: DragStartEvent) => {
    setDraggingKey(String(active.id));
  }, []);

  const handleDragOver = useCallback(({ active, over }: DragOverEvent) => {
    if (!over || active.id === over.id) {
      setDragOverKey(null);
      return;
    }

    const activeSection = active.data.current?.sectionKey as SectionKey | undefined;
    const overSection = over.data.current?.sectionKey as SectionKey | undefined;

    if (!activeSection || !overSection || activeSection !== overSection) {
      setDragOverKey(null);
      return;
    }

    setDragOverKey(String(over.id));
  }, []);

  const handleDragEnd = useCallback(
    ({ active, over }: DragEndEvent) => {
      setDraggingKey(null);
      setDragOverKey(null);

      if (!over || active.id === over.id) return;

      const activeKey = String(active.id);
      const targetKey = String(over.id);
      const sourceItem = items.find(item => item.key === activeKey);
      const targetItem = items.find(item => item.key === targetKey);

      if (!sourceItem || !targetItem) return;
      if (getSectionKey(sourceItem.fixed) !== getSectionKey(targetItem.fixed)) return;

      onReorder(activeKey, targetKey, targetItem.fixed);
    },
    [items, onReorder]
  );

  const handleDragCancel = useCallback(() => {
    setDraggingKey(null);
    setDragOverKey(null);
  }, []);

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragOver={handleDragOver}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <div
        className={styles.panel}
        role="dialog"
        aria-label="调整表头"
        data-persist-enabled={shouldPersist ? "true" : "false"}
      >
        <MhFlex align="center" justify="space-between" className={styles.header}>
          <MhCheckbox
            checked={allChecked}
            indeterminate={indeterminate}
            onChange={event => onToggleAll(event.target.checked)}
          >
            <span className={styles.headerLabel}>列展示</span>
          </MhCheckbox>
          <MhButton type="link" className={styles.resetButton} onClick={onReset}>
            重置
          </MhButton>
        </MhFlex>

        <div className={styles.content}>
          {sections.map(section => (
            <section key={section.key} className={styles.section}>
              {showSectionTitle ? (
                <MhTypography.Text className={styles.sectionTitle}>{section.title}</MhTypography.Text>
              ) : null}
              <SortableContext items={section.items.map(item => item.key)} strategy={verticalListSortingStrategy}>
                <div className={styles.sectionList}>
                  {section.items.map(item => (
                    <SortableRow
                      key={item.key}
                      item={item}
                      sectionKey={section.key}
                      isDragOver={draggingKey !== item.key && dragOverKey === item.key}
                      onToggle={onToggle}
                      onMove={onMove}
                    />
                  ))}
                </div>
              </SortableContext>
            </section>
          ))}
        </div>
      </div>
      <DragOverlay zIndex={1200} dropAnimation={null}>
        {draggingItem ? (
          <BaseRow
            item={draggingItem}
            className={cx(styles.row, styles["row-overlay"], draggingItem.fixed && styles["row-fixed"])}
            onToggle={() => {}}
            onMove={() => {}}
          />
        ) : null}
      </DragOverlay>
    </DndContext>
  );
};

export default TableAdjustHeader;
