import { type MhConBreadcrumbItem, MhConException } from "@mh-repo/ui";
import type React from "react";
import { useNavigate } from "react-router";
import { ExceptionIllustration } from "../components/ExceptionIllustration";

const Exception403: React.FC = () => {
  const navigate = useNavigate();

  const breadcrumbItems: MhConBreadcrumbItem[] = [
    { key: "dsp", title: "DSP平台", path: "/", clickable: true },
    { key: "exception", title: "异常页", path: "/exception", clickable: true },
    { key: "403", title: "403状态" }
  ];

  return (
    <MhConException
      status="403"
      illustration={<ExceptionIllustration status="403" />}
      breadcrumbItems={breadcrumbItems}
      onBreadcrumbClick={item => item.path && navigate(item.path)}
      onBackHome={() => navigate("/")}
      cardClassName="card_minHeight"
    />
  );
};

export default Exception403;
