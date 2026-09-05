import type React from "react";
import { MhBreadcrumb } from "../../../Navigation/Breadcrumb";
import { MhConBreadcrumbSeparatorIcon } from "./icons";
import styles from "./index.module.less";
import type { MhConBreadcrumbItem } from "./types";

const DefaultSeparator: React.FC = () => <MhConBreadcrumbSeparatorIcon className={styles.breadcrumbSeparator} />;

export interface MhConBreadcrumbProps {
  items: MhConBreadcrumbItem[];
  separator?: React.ReactNode;
  onItemClick?: (item: MhConBreadcrumbItem, index: number) => void;
  className?: string;
}

export const MhConBreadcrumb: React.FC<MhConBreadcrumbProps> = ({
  items,
  separator = <DefaultSeparator />,
  onItemClick,
  className = ""
}) => {
  const handleItemClick = (item: MhConBreadcrumbItem, index: number) => {
    if (item.clickable !== false && item.path && onItemClick) {
      onItemClick(item, index);
    }
  };

  return (
    <MhBreadcrumb
      separator={separator}
      className={[styles.breadcrumb, className].filter(Boolean).join(" ")}
      items={items.map((item, index) => ({
        key: item.key,
        title:
          item.clickable !== false && item.path ? (
            <span className={styles.breadcrumbLink} onClick={() => handleItemClick(item, index)}>
              {item.title}
            </span>
          ) : (
            <span className={styles.breadcrumbCurrent}>{item.title}</span>
          )
      }))}
    />
  );
};

export default MhConBreadcrumb;
