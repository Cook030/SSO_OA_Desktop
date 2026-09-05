export interface MhConBreadcrumbItem {
  /** 唯一标识 */
  key: string;
  /** 显示标题 */
  title: string;
  /** 点击跳转路径，可选 */
  path?: string;
  /** 是否可点击 */
  clickable?: boolean;
}

export interface MhConFavoriteItem {
  id: string;
  title: string;
  url: string;
}
