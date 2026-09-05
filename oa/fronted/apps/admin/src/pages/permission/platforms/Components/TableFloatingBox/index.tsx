import { MhFlex, MhTypography } from "@mh-repo/ui";
import type React from "react";
import styles from "./index.module.less";

export interface TableFloatingBoxProps {
  collectorId: string;
  groupName: string;
}

const TableFloatingBoxRow: React.FC<{
  label: string;
  value: React.ReactNode;
  copy?: boolean;
}> = ({ label, value, copy }) => (
  <div className={styles.row}>
    <MhFlex align="center" justify="space-between" className={styles.label}>
      <MhTypography.Text>{label}</MhTypography.Text>
    </MhFlex>
    <MhFlex align="center" justify="space-between" className={styles.value}>
      <MhTypography.Text copyable={copy}>{value}</MhTypography.Text>
    </MhFlex>
  </div>
);

const TableFloatingBox: React.FC<TableFloatingBoxProps> = ({ collectorId, groupName }) => {
  return (
    <div className={styles.floatingBox} role="tooltip">
      <TableFloatingBoxRow
        label="采集器ID"
        value={collectorId}
        copy={true}
        // copy={<MhButton type="text" className={styles.copy} aria-label="复制采集器ID" icon={<CopyIcon />} />}
      />
      <TableFloatingBoxRow label="分组" value={groupName} />
    </div>
  );
};

export default TableFloatingBox;
