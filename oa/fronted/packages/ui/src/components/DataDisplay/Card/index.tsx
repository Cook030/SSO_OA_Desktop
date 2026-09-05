import { Card, type CardProps } from "antd";
import type React from "react";

export interface MhCardProps extends CardProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhCard: React.FC<MhCardProps> & {
  Grid: typeof Card.Grid;
  Meta: typeof Card.Meta;
} = ({ ...restProps }) => {
  return <Card {...restProps} />;
};

MhCard.Grid = Card.Grid;
MhCard.Meta = Card.Meta;

export default MhCard;
