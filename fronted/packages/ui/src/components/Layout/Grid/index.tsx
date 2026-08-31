import { Col, type ColProps, Row, type RowProps } from "antd";
import type React from "react";

export interface MhGridRowProps extends RowProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export interface MhGridColProps extends ColProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhRow: React.FC<MhGridRowProps> = ({ ...restProps }) => {
  return <Row {...restProps} />;
};

export const MhCol: React.FC<MhGridColProps> = ({ ...restProps }) => {
  return <Col {...restProps} />;
};

export const MhGrid = {
  Row: MhRow,
  Col: MhCol
};

export default MhGrid;
