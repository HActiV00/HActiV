// profile/components/Policy.js
import React from "react";
import Card from "@mui/material/Card";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDButton from "components/MDButton";

function Policy() {
  // 상세보기 버튼 클릭 시 알림 표시
  const handleViewMoreClick = () => {
    alert("상세보기 기능 준비 중");
  };

  return (
    <Card>
      <MDBox p={2} position="relative">
        {/* 상세보기 버튼 */}
        <MDBox position="absolute" top={16} right={16}>
          <MDButton variant="text" color="dark" size="small" onClick={handleViewMoreClick}>
            상세보기
          </MDButton>
        </MDBox>

        {/* 카드 내용 */}
        <MDTypography variant="h6" fontWeight="medium">
          정책 안내
        </MDTypography>
        <MDTypography variant="body2">
          정책과 규정을 확인하세요.
        </MDTypography>
      </MDBox>
    </Card>
  );
}

export default Policy;
