import { createSvgComp, type SvgCompProps } from "@mh-repo/ui";
import { antdIconRegistry, mhIconRegistry } from "./iconRegistry";

export type { AntdIconName, MhIconName } from "./iconRegistry";
export { antdIconRegistry, mhIconRegistry } from "./iconRegistry";
export type { SvgCompProps };

const { SvgComp, renderSvgIcon, isAntdIconName, isMhIconName } = createSvgComp({
  antdIconRegistry,
  mhIconRegistry
});

export { isAntdIconName, isMhIconName, renderSvgIcon, SvgComp };
export default SvgComp;
