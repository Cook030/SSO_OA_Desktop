import { Statistic, type StatisticProps } from "antd";
import type React from "react";

export interface MhStatisticProps extends StatisticProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhStatistic: React.FC<MhStatisticProps> & {
  Countdown: typeof Statistic.Countdown;
} = ({ ...restProps }) => {
  return <Statistic {...restProps} />;
};

MhStatistic.Countdown = Statistic.Countdown;

export default MhStatistic;
