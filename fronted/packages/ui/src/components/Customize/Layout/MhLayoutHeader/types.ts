import type React from "react";

export interface MhLayoutHeaderProps {
  /** 完全自定义左侧品牌区；传入后忽略 brandTitle / logoUrl */
  brand?: React.ReactNode;
  /** 完全自定义右侧操作区；传入后忽略 platformLabel / user* / actionsExtra */
  actions?: React.ReactNode;
  /** 品牌文字，默认 Maplehaze */
  brandTitle?: string;
  /** Logo 地址，默认 /logo.png */
  logoUrl?: string;
  /** 平台名称标签，如「DSP平台」 */
  platformLabel?: React.ReactNode;
  /** 用户问候语 */
  userGreeting?: React.ReactNode;
  /** 用户角色标签 */
  userRole?: React.ReactNode;
  /** 用户头像 */
  userAvatar?: React.ReactNode;
  /** 用户信息区额外内容（如权限引导 Tag） */
  actionsExtra?: React.ReactNode;
  className?: string;
}

/** 独立运行模式顶栏高度（px） */
export const MH_LAYOUT_HEADER_HEIGHT = 65;
