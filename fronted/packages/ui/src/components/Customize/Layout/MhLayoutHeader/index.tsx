import type React from "react";
import { MhAvatar } from "../../../DataDisplay/Avatar";
import { UserOutlined } from "../../../General/Icon";
import { MhLayout } from "../../../Layout/Layout";
import styles from "./index.module.less";
import type { MhLayoutHeaderProps } from "./types";

const { Header } = MhLayout;

export type { MhLayoutHeaderProps } from "./types";
export { MH_LAYOUT_HEADER_HEIGHT } from "./types";

export const MhLayoutHeader: React.FC<MhLayoutHeaderProps> = ({
  brand,
  actions,
  brandTitle = "Maplehaze",
  logoUrl = "/logo.png",
  platformLabel,
  userGreeting,
  userRole,
  userAvatar,
  actionsExtra,
  className
}) => {
  const showDefaultActions =
    actions === undefined &&
    (platformLabel !== undefined || userGreeting !== undefined || userRole !== undefined || actionsExtra !== undefined);

  return (
    <Header className={[styles["mh-layout-header"], className].filter(Boolean).join(" ")}>
      {brand ?? (
        <div className={styles["mh-layout-header-brand"]}>
          <div
            className={styles["mh-layout-header-brand-mark"]}
            style={{ backgroundImage: `url("${logoUrl}")` }}
            role="img"
            aria-label={brandTitle}
          />
          <div className={styles["mh-layout-header-brand-text"]}>{brandTitle}</div>
        </div>
      )}

      {actions ??
        (showDefaultActions ? (
          <div className={styles["mh-layout-header-actions"]}>
            {platformLabel !== undefined && platformLabel !== null && platformLabel !== "" ? (
              <div className={styles["mh-layout-header-platform-chip"]}>{platformLabel}</div>
            ) : null}
            {(userGreeting !== undefined || userRole !== undefined || actionsExtra !== undefined) && (
              <div className={styles["mh-layout-header-user-chip"]}>
                {userAvatar ?? <MhAvatar size={20} icon={<UserOutlined />} />}
                {userGreeting !== undefined && userGreeting !== null && userGreeting !== "" ? (
                  <span>{userGreeting}</span>
                ) : null}
                {userRole !== undefined && userRole !== null && userRole !== "" ? (
                  <span className={styles["mh-layout-header-role-chip"]}>{userRole}</span>
                ) : null}
                {actionsExtra}
              </div>
            )}
          </div>
        ) : null)}
    </Header>
  );
};

export default MhLayoutHeader;
