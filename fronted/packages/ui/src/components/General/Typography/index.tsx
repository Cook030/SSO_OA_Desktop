import { Typography } from "antd";
import type React from "react";

const { Title, Text, Paragraph, Link } = Typography;
type AntdTypographyProps = React.ComponentProps<typeof Typography>;

export interface MhTypographyProps extends AntdTypographyProps {
  mhProps: {
    className?: string;
  };
}

export const MhTypography: React.FC<MhTypographyProps> & {
  Title: typeof Title;
  Text: typeof Text;
  Paragraph: typeof Paragraph;
  Link: typeof Link;
} = ({ ...restProps }) => {
  return <Typography {...restProps} />;
};

MhTypography.Title = Title;
MhTypography.Text = Text;
MhTypography.Paragraph = Paragraph;
MhTypography.Link = Link;

export default MhTypography;
