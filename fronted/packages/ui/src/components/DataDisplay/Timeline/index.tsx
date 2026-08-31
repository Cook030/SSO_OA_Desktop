import { Timeline, type TimelineProps } from "antd";
import type React from "react";

export interface MhTimelineProps extends TimelineProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhTimeline: React.FC<MhTimelineProps> & {
  Item: typeof Timeline.Item;
} = ({ ...restProps }) => {
  return <Timeline {...restProps} />;
};

MhTimeline.Item = Timeline.Item;

export default MhTimeline;
