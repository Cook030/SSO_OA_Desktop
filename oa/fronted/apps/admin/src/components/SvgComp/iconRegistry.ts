import HomeIcon from "@assets/home.svg?react";
import Error403Icon from "@assets/images/403.svg?react";
import Error404Icon from "@assets/images/404.svg?react";
import Error500Icon from "@assets/images/500.svg?react";
import ToDownIcon from "@assets/images/table/todown.svg?react";
import ToMiddleIcon from "@assets/images/table/topmiddle.svg?react";
import ToTopIcon from "@assets/images/table/totop.svg?react";
import LogoIcon from "@assets/logo.svg?react";
import QuestionIcon from "@assets/question.svg?react";
import {
  AndroidOutlined,
  AppleOutlined,
  BarChartOutlined,
  BellOutlined,
  CloseOutlined,
  DashboardOutlined,
  EyeOutlined,
  FileTextOutlined,
  FolderOpenOutlined,
  FormOutlined,
  LockOutlined,
  MoreOutlined,
  PieChartOutlined,
  PlusCircleOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
  SoundOutlined,
  UnorderedListOutlined,
  UserOutlined
} from "@mh-repo/ui";
import type React from "react";
import { MhBreadcrumbSeparatorIcon, MhCopyIcon, MhFilterIcon, MhInfoIcon, MhStarIcon } from "./mhInlineIcons";

type AntdIconComponent = React.ComponentType<{
  className?: string;
  style?: React.CSSProperties;
  spin?: boolean;
}>;

type MhIconComponent = React.ComponentType<Record<string, unknown>>;

/** Ant Design 图标（经 @mh-repo/ui 导出） */
export const antdIconRegistry: Record<string, AntdIconComponent> = {
  AndroidOutlined,
  AppleOutlined,
  BarChartOutlined,
  BellOutlined,
  CloseOutlined,
  DashboardOutlined,
  EyeOutlined,
  FileTextOutlined,
  FolderOpenOutlined,
  FormOutlined,
  LockOutlined,
  MoreOutlined,
  PieChartOutlined,
  PlusCircleOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
  SoundOutlined,
  UnorderedListOutlined,
  UserOutlined
};

/** MapleHaze 自有图标（mh- 前缀，含 SVG 资源与内联图形） */
export const mhIconRegistry: Record<string, MhIconComponent> = {
  "mh-home": HomeIcon,
  "mh-question": QuestionIcon,
  "mh-logo": LogoIcon,
  "mh-table-totop": ToTopIcon,
  "mh-table-todown": ToDownIcon,
  "mh-table-topmiddle": ToMiddleIcon,
  "mh-403": Error403Icon,
  "mh-404": Error404Icon,
  "mh-500": Error500Icon,
  "mh-filter": MhFilterIcon,
  "mh-info": MhInfoIcon,
  "mh-star": MhStarIcon,
  "mh-copy": MhCopyIcon,
  "mh-breadcrumb-separator": MhBreadcrumbSeparatorIcon
};

export type AntdIconName = keyof typeof antdIconRegistry;
export type MhIconName = keyof typeof mhIconRegistry;
