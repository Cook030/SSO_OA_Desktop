import type React from "react";
import { MhConPageHeaderHomeIcon, MhConPageHeaderStarIcon } from "./icons";
import styles from "./index.module.less";
import { MhConBreadcrumb } from "./MhConBreadcrumb";
import type { MhConBreadcrumbItem } from "./types";
import { useMhFavorites } from "./useMhFavorites";

export { MhConBreadcrumbSeparatorIcon, MhConPageHeaderHomeIcon, MhConPageHeaderStarIcon } from "./icons";
export type { MhConBreadcrumbProps } from "./MhConBreadcrumb";
export { MhConBreadcrumb } from "./MhConBreadcrumb";
export type { MhConBreadcrumbItem, MhConFavoriteItem } from "./types";
export {
  dispatchMhFavoritesChange,
  MH_FAVORITES_CHANGE_EVENT,
  MH_FAVORITES_STORAGE_KEY,
  useMhFavorites
} from "./useMhFavorites";

export interface MhConPageHeaderProps {
  /** 页面标题 */
  title: string;
  /** 面包屑项列表 */
  breadcrumbItems: MhConBreadcrumbItem[];
  /** 点击面包屑项的回调 */
  onBreadcrumbClick?: (item: MhConBreadcrumbItem, index: number) => void;
  /** 是否显示收藏按钮 */
  showFavorite?: boolean;
  /** 收藏状态（受控） */
  isFavorite?: boolean;
  /** 点击收藏的回调 */
  onFavoriteClick?: (isFavorite: boolean) => void;
  /** 当前页面路径，用于收藏匹配，默认 window.location.pathname */
  currentPath?: string;
  /** 自定义首页图标 */
  homeIcon?: React.ReactNode;
  /** 自定义类名 */
  className?: string;
}

const getDefaultCurrentPath = () => {
  if (typeof window === "undefined") {
    return "";
  }

  return window.location.pathname;
};

/**
 * 内容页头部：面包屑 + 收藏 + 标题
 */
export const MhConPageHeader: React.FC<MhConPageHeaderProps> = ({
  title,
  breadcrumbItems,
  onBreadcrumbClick,
  showFavorite = true,
  isFavorite: controlledFavorite,
  onFavoriteClick,
  currentPath = getDefaultCurrentPath(),
  homeIcon,
  className
}) => {
  const { favorites, addFavorite, removeFavorite } = useMhFavorites();
  const isFavorite =
    controlledFavorite !== undefined ? controlledFavorite : favorites.some(favorite => favorite.url === currentPath);

  const handleFavoriteClick = () => {
    const nextFavorite = !isFavorite;

    if (controlledFavorite === undefined) {
      if (nextFavorite) {
        addFavorite({
          id: currentPath,
          title,
          url: currentPath
        });
      } else {
        removeFavorite(currentPath);
      }
    }

    onFavoriteClick?.(nextFavorite);
  };

  return (
    <div className={[styles.pageHeader, className].filter(Boolean).join(" ")}>
      <div className={styles.breadcrumbBar}>
        <div className={styles.breadcrumbRow}>
          <div className={styles.breadcrumbList}>
            {homeIcon ?? <MhConPageHeaderHomeIcon className={styles.homeIcon} aria-hidden />}
            <MhConBreadcrumb items={breadcrumbItems} onItemClick={onBreadcrumbClick} />
          </div>

          {showFavorite ? (
            <div className={styles.favoriteAction} onClick={handleFavoriteClick}>
              <span className={styles.favoriteLabel}>收藏</span>
              <MhConPageHeaderStarIcon
                className={styles.favoriteIcon}
                filled={isFavorite}
                color={isFavorite ? "#FAAE13" : "rgba(0, 0, 0, 0.25)"}
              />
            </div>
          ) : null}
        </div>
      </div>

      <div className={styles.titleBar}>
        <div>{title}</div>
      </div>
    </div>
  );
};

export default MhConPageHeader;
