import type React from "react";
import { MhCard } from "../../../DataDisplay/Card";
import { MhButton } from "../../../General/Button";
import type { MhConBreadcrumbItem } from "../MhConPageHeader";
import { MhConPageHeader } from "../MhConPageHeader";
import { MhConExceptionIllustration } from "./illustrations";
import styles from "./index.module.less";
import { MH_CON_EXCEPTION_PRESETS, type MhConExceptionStatus } from "./presets";

export type { MhConExceptionIllustrationProps } from "./illustrations";
export { MhConExceptionIllustration } from "./illustrations";
export type { MhConExceptionPreset, MhConExceptionStatus } from "./presets";
export { MH_CON_EXCEPTION_PRESETS } from "./presets";

export interface MhConExceptionProps {
  /** 异常状态码 */
  status: MhConExceptionStatus;
  /** 页面标题，默认使用预设文案 */
  title?: string;
  /** 描述文案，默认使用预设文案 */
  description?: string;
  /** 面包屑项 */
  breadcrumbItems: MhConBreadcrumbItem[];
  /** 面包屑点击 */
  onBreadcrumbClick?: (item: MhConBreadcrumbItem, index: number) => void;
  /** 是否显示收藏 */
  showFavorite?: boolean;
  /** 自定义插图 */
  illustration?: React.ReactNode;
  /** 返回首页 */
  onBackHome?: () => void;
  /** 返回首页按钮文案 */
  backHomeText?: string;
  /** 是否显示返回首页按钮 */
  showBackHome?: boolean;
  /** 500 页刷新回调，status=500 时默认 window.location.reload */
  onRefresh?: () => void;
  /** 刷新按钮文案 */
  refreshText?: string;
  /** 是否显示刷新按钮（仅 500） */
  showRefresh?: boolean;
  /** 额外操作区 */
  extraActions?: React.ReactNode;
  className?: string;
  cardClassName?: string;
}

const defaultRefresh = () => {
  window.location.reload();
};

/**
 * 内容异常页：403 / 404 / 500
 */
export const MhConException: React.FC<MhConExceptionProps> = ({
  status,
  title,
  description,
  breadcrumbItems,
  onBreadcrumbClick,
  showFavorite = true,
  illustration,
  onBackHome,
  backHomeText = "返回首页",
  showBackHome = true,
  onRefresh,
  refreshText = "刷新一下",
  showRefresh,
  extraActions,
  className,
  cardClassName
}) => {
  const preset = MH_CON_EXCEPTION_PRESETS[status];
  const pageTitle = title ?? preset.title;
  const pageDescription = description ?? preset.description;
  const shouldShowRefresh = showRefresh ?? status === "500";

  return (
    <div className={[styles.page, className].filter(Boolean).join(" ")}>
      <MhConPageHeader
        title={pageTitle}
        breadcrumbItems={breadcrumbItems}
        onBreadcrumbClick={onBreadcrumbClick}
        showFavorite={showFavorite}
      />

      <MhCard className={[styles.card, cardClassName].filter(Boolean).join(" ")}>
        <div className={styles.content}>
          {illustration ?? <MhConExceptionIllustration status={status} className={styles.illustration} />}
          <div
            className={[styles.description, status === "500" ? styles.description500 : ""].filter(Boolean).join(" ")}
          >
            {pageDescription}
          </div>
          <div className={styles.actions}>
            {extraActions}
            {shouldShowRefresh ? <MhButton onClick={onRefresh ?? defaultRefresh}>{refreshText}</MhButton> : null}
            {showBackHome && onBackHome ? (
              <MhButton type="primary" onClick={onBackHome}>
                {backHomeText}
              </MhButton>
            ) : null}
          </div>
        </div>
      </MhCard>
    </div>
  );
};

export default MhConException;
