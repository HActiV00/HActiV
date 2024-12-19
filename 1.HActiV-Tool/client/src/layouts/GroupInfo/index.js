// src/layouts/GroupInfo/index.js

import React from "react";
import { Card, Grid, Avatar } from "@mui/material"; // Avatar 추가
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import DashboardLayout from "examples/LayoutContainers/DashboardLayout";
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import Footer from "examples/Footer";

function GroupInfo() {
  return (
    <DashboardLayout>
      <DashboardNavbar />
      <MDBox mt={4} mb={3} px={3}>
        <Grid container justifyContent="center">
          <Grid item xs={12} md={10} lg={8}>
            <Card sx={{ p: 3 }}>
              <MDTypography variant="h6" fontWeight="medium" mb={3}>
                그룹 정보
              </MDTypography>
              <MDTypography variant="body1" color="textSecondary" mb={2}>
                그룹 구성원 및 그룹 정보를 확인하고 관리할 수 있는 페이지입니다.
              </MDTypography>
              {/* 그룹 구성원 및 그룹 정보를 위한 컴포넌트 추가 */}
              <MDBox mt={3}>
                <MDTypography variant="h6" fontWeight="medium" mb={1}>
                  구성원 목록
                </MDTypography>
                {/* 구성원 리스트 예시 */}
                <MDBox>
                  {[...Array(5)].map((_, index) => (
                    <MDBox display="flex" alignItems="center" mb={1} key={index}>
                      <Avatar sx={{ width: 36, height: 36, mr: 2 }}>U</Avatar>
                      <MDTypography variant="body2">Username {index + 1}</MDTypography>
                    </MDBox>
                  ))}
                </MDBox>
              </MDBox>
            </Card>
          </Grid>
        </Grid>
      </MDBox>
      <Footer />
    </DashboardLayout>
  );
}

export default GroupInfo;
