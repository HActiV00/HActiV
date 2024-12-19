// profile/components/PlatformSettings/index.js
import { useState } from "react";
import Card from "@mui/material/Card";

// Material Dashboard 2 React components
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDButton from "components/MDButton"; // 추가: 저장 버튼

function PlatformSettings() {
  const [emailNotifications, setEmailNotifications] = useState(true);
  const [alertNotifications, setAlertNotifications] = useState(false);

  return (
    <Card sx={{ boxShadow: "none" }}>
      <MDBox p={2}>
        <MDTypography variant="h6" fontWeight="medium">
          계정 설정
        </MDTypography>
      </MDBox>
      <MDBox pt={1} pb={2} px={2} lineHeight={1.25}>
        <MDTypography variant="caption" fontWeight="bold" color="text" textTransform="uppercase">
          알림 설정
        </MDTypography>
        <MDBox display="flex" alignItems="center" mb={0.5}>
          <MDBox width="80%" ml={0.5}>
            <MDTypography variant="button" fontWeight="regular" color="text">
              이메일 알림 받기
            </MDTypography>
          </MDBox>
          <MDButton
            variant="contained"
            color={emailNotifications ? "info" : "secondary"}
            onClick={() => setEmailNotifications(!emailNotifications)}
          >
            {emailNotifications ? "On" : "Off"}
          </MDButton>
        </MDBox>
        <MDBox display="flex" alignItems="center" mb={0.5}>
          <MDBox width="80%" ml={0.5}>
            <MDTypography variant="button" fontWeight="regular" color="text">
              경고 알림 받기
            </MDTypography>
          </MDBox>
          <MDButton
            variant="contained"
            color={alertNotifications ? "info" : "secondary"}
            onClick={() => setAlertNotifications(!alertNotifications)}
          >
            {alertNotifications ? "On" : "Off"}
          </MDButton>
        </MDBox>
      </MDBox>
    </Card>
  );
}

export default PlatformSettings;
