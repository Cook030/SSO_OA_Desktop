import { ConfigProvider, type ConfigProviderProps } from "antd";
import zhCN from "antd/locale/zh_CN";
import dayjs from "dayjs";
import type React from "react";
import "dayjs/locale/zh-cn"; // 引入 dayjs 中文语言包

dayjs.locale("zh-cn"); // 全局设置 dayjs 语言

export interface MhConfigProviderProps extends ConfigProviderProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhConfigProvider: React.FC<MhConfigProviderProps> & {
  useConfig: typeof ConfigProvider.useConfig;
  config: typeof ConfigProvider.config;
} = ({ ...restProps }) => {
  return <ConfigProvider locale={zhCN} {...restProps} />;
};

MhConfigProvider.useConfig = ConfigProvider.useConfig;
MhConfigProvider.config = ConfigProvider.config;

export default MhConfigProvider;
