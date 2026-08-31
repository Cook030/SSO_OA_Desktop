import { Skeleton, type SkeletonProps } from "antd";
import type React from "react";

export interface MhSkeletonProps extends SkeletonProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhSkeleton: React.FC<MhSkeletonProps> & {
  Button: typeof Skeleton.Button;
  Avatar: typeof Skeleton.Avatar;
  Input: typeof Skeleton.Input;
  Image: typeof Skeleton.Image;
  Node: typeof Skeleton.Node;
} = ({ ...restProps }) => {
  return <Skeleton {...restProps} />;
};

MhSkeleton.Button = Skeleton.Button;
MhSkeleton.Avatar = Skeleton.Avatar;
MhSkeleton.Input = Skeleton.Input;
MhSkeleton.Image = Skeleton.Image;
MhSkeleton.Node = Skeleton.Node;

export default MhSkeleton;
