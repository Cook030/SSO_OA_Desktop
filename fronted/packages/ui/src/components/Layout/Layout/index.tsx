import { Layout, type LayoutProps } from "antd";
import type React from "react";

const { Header, Footer, Sider, Content } = Layout;

export interface MhLayoutProps extends LayoutProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhLayout: React.FC<MhLayoutProps> & {
  Header: typeof Header;
  Footer: typeof Footer;
  Sider: typeof Sider;
  Content: typeof Content;
} = ({ ...restProps }) => {
  return <Layout {...restProps} />;
};

MhLayout.Header = Header;
MhLayout.Footer = Footer;
MhLayout.Sider = Sider;
MhLayout.Content = Content;

export default MhLayout;
