import type { TypographyProps } from "antd";
import { Typography } from "antd";

export interface MhTypographyProps extends TypographyProps {}

const MhTypography = Typography as React.FC<MhTypographyProps> & {
  Paragraph: typeof Typography.Paragraph;
  Text: typeof Typography.Text;
  Title: typeof Typography.Title;
  Link: typeof Typography.Link;
};

// 确保静态属性被正确赋值
MhTypography.Paragraph = Typography.Paragraph;
MhTypography.Text = Typography.Text;
MhTypography.Title = Typography.Title;
MhTypography.Link = Typography.Link;

export default MhTypography;
