import { Descriptions, type DescriptionsProps } from "antd";
import type React from "react";

export interface MhDescriptionsProps extends DescriptionsProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhDescriptions: React.FC<MhDescriptionsProps> & {
  Item: typeof Descriptions.Item;
} = ({ ...restProps }) => {
  return <Descriptions {...restProps} />;
};

MhDescriptions.Item = Descriptions.Item;

export default MhDescriptions;
