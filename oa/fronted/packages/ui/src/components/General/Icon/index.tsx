import Icon, {
  AndroidOutlined,
  AppleOutlined,
  BarChartOutlined,
  BellOutlined,
  CheckCircleOutlined,
  CloseOutlined,
  CopyOutlined,
  createFromIconfontCN,
  DashboardOutlined,
  EditOutlined,
  ExclamationCircleFilled,
  ExportOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
  FileTextOutlined,
  FolderOpenOutlined,
  FormOutlined,
  HolderOutlined,
  LeftOutlined,
  LockOutlined,
  MailOutlined,
  MobileOutlined,
  MoreOutlined,
  PieChartOutlined,
  PlusCircleOutlined,
  PoweroffOutlined,
  ProductOutlined,
  ReloadOutlined,
  RightOutlined,
  SearchOutlined,
  SettingOutlined,
  SoundOutlined,
  UnorderedListOutlined,
  UserOutlined
} from "@ant-design/icons";
import type { IconFontProps } from "@ant-design/icons/lib/components/IconFont";
import type React from "react";

type IconProps = React.ComponentProps<typeof Icon>;

export interface MhIconFontProps extends IconFontProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const IconFont = createFromIconfontCN({
  scriptUrl: [
    "//at.alicdn.com/t/font_1788044_0dwu4guekcwr.js", // icon-javascript, icon-java, icon-shoppingcart (overridden)
    "//at.alicdn.com/t/font_1788592_a5xf2bdic3u.js" // icon-shoppingcart, icon-python
  ]
});

export const MhIconFont: React.FC<MhIconFontProps> = ({ ...restProps }) => {
  return <IconFont {...restProps} />;
};

export interface MhIconProps extends IconProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhIcon: React.FC<MhIconProps> = ({ ...restProps }) => {
  return <Icon {...restProps} />;
};

// 导出常用图标
export {
  AndroidOutlined,
  AppleOutlined,
  BarChartOutlined,
  BellOutlined,
  CheckCircleOutlined,
  CloseOutlined,
  CopyOutlined,
  DashboardOutlined,
  EditOutlined,
  ExclamationCircleFilled,
  ExportOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
  FileTextOutlined,
  FolderOpenOutlined,
  FormOutlined,
  HolderOutlined,
  LeftOutlined,
  LockOutlined,
  MailOutlined,
  MobileOutlined,
  MoreOutlined,
  PieChartOutlined,
  PlusCircleOutlined,
  PoweroffOutlined,
  ProductOutlined,
  ReloadOutlined,
  RightOutlined,
  SearchOutlined,
  SettingOutlined,
  SoundOutlined,
  UnorderedListOutlined,
  UserOutlined
};

export default MhIcon;
