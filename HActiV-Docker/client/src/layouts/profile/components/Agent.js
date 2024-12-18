// layouts/profile/components/Agent.js
import React from "react";
import { useNavigate } from "react-router-dom";
import Card from "@mui/material/Card";
import Grid from "@mui/material/Grid";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDButton from "components/MDButton";

function Agent({ totalAgents = 10, activeAgents = 7, inactiveAgents = 3 }) {
  const navigate = useNavigate(); // 페이지 이동을 위한 훅

  return (
    <Card sx={{ p: 3, position: "relative" }}>
      {/* 상세보기 버튼 */}
      <MDBox position="absolute" top={16} right={16}>
        <MDButton
          variant="text"
          color="dark"
          size="small"
          onClick={() => navigate("/agent-control")} // 페이지 전환
        >
          상세보기
        </MDButton>
      </MDBox>

      {/* 카드 본문 */}
      <MDBox mb={2}>
        <MDTypography variant="h6" fontWeight="medium" color="dark">
          Agent 관리
        </MDTypography>
        <MDTypography variant="body2" color="textSecondary" sx={{ fontSize: "0.875rem" }}>
          컨테이너 형태의 에이전트 관리 및 설정
        </MDTypography>
      </MDBox>

      <Grid container spacing={2}>
        {/* 총 에이전트 개수 */}
        <Grid item xs={12} sm={4}>
          <MDBox display="flex" flexDirection="column" alignItems="center">
            <MDTypography variant="h5" color="info" fontWeight="bold">
              {totalAgents}
            </MDTypography>
            <MDTypography variant="caption" color="textSecondary">
              총 생성된 Agent
            </MDTypography>
          </MDBox>
        </Grid>
        {/* 실행 중인 에이전트 */}
        <Grid item xs={12} sm={4}>
          <MDBox display="flex" flexDirection="column" alignItems="center">
            <MDTypography variant="h5" color="success" fontWeight="bold">
              {activeAgents}
            </MDTypography>
            <MDTypography variant="caption" color="textSecondary">
              시행 중인 Agent
            </MDTypography>
          </MDBox>
        </Grid>
        {/* 꺼진 에이전트 */}
        <Grid item xs={12} sm={4}>
          <MDBox display="flex" flexDirection="column" alignItems="center">
            <MDTypography variant="h5" color="error" fontWeight="bold">
              {inactiveAgents}
            </MDTypography>
            <MDTypography variant="caption" color="textSecondary">
              꺼진 Agent
            </MDTypography>
          </MDBox>
        </Grid>
      </Grid>
    </Card>
  );
}

export default Agent;
