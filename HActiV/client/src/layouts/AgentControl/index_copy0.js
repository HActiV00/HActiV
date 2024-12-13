// layouts/AgentControl/index.js
import React, { useState } from "react";
import Card from "@mui/material/Card";
import Grid from "@mui/material/Grid";
import Modal from "@mui/material/Modal";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDButton from "components/MDButton";
import DashboardLayout from "examples/LayoutContainers/DashboardLayout";
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import Footer from "examples/Footer";

// Sample Data
const agentData = [
  { name: "Agent 1", status: "Active", containers: ["Container A", "Container B", "Container C"] },
  { name: "Agent 2", status: "Inactive", containers: ["Container D", "Container E"] },
  { name: "Agent 3", status: "Active", containers: ["Container F", "Container G", "Container H", "Container I"] },
];

function AgentControl() {
  const [selectedAgent, setSelectedAgent] = useState(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const totalAgents = agentData.length;
  const activeAgents = agentData.filter(agent => agent.status === "Active").length;
  const inactiveAgents = totalAgents - activeAgents;

  const openModal = (agent) => {
    setSelectedAgent(agent);
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setSelectedAgent(null);
  };

  return (
    <DashboardLayout>
      <DashboardNavbar />
      <MDBox py={3}>
        <Grid container spacing={3}>
          {/* 총 생성된 Agent, 시행 중인 Agent, 꺼진 Agent */}
          <Grid item xs={12} md={4}>
            <Card sx={{ p: 3, textAlign: "center" }}>
              <MDTypography variant="h5" fontWeight="medium" color="info">
                {totalAgents}
              </MDTypography>
              <MDTypography variant="caption" color="textSecondary">
                총 생성된 Agent
              </MDTypography>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card sx={{ p: 3, textAlign: "center" }}>
              <MDTypography variant="h5" fontWeight="medium" color="success">
                {activeAgents}
              </MDTypography>
              <MDTypography variant="caption" color="textSecondary">
                시행 중인 Agent
              </MDTypography>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card sx={{ p: 3, textAlign: "center" }}>
              <MDTypography variant="h5" fontWeight="medium" color="error">
                {inactiveAgents}
              </MDTypography>
              <MDTypography variant="caption" color="textSecondary">
                꺼진 Agent
              </MDTypography>
            </Card>
          </Grid>

          {/* Agent 목록 */}
          <Grid item xs={12}>
            <Card sx={{ p: 3 }}>
              <MDTypography variant="h6" fontWeight="medium" mb={2}>
                Agent 목록
              </MDTypography>
              <Grid container spacing={2}>
                {agentData.map((agent, index) => (
                  <Grid item xs={12} md={4} key={index}>
                    <Card sx={{ p: 2 }}>
                      <MDTypography variant="h6" fontWeight="medium">
                        {agent.name}
                      </MDTypography>
                      <MDTypography variant="body2" color="textSecondary">
                        상태: {agent.status}
                      </MDTypography>
                      <MDTypography variant="body2" color="textSecondary">
                        연결된 컨테이너: {agent.containers.length}
                      </MDTypography>
                      <MDBox display="flex" justifyContent="space-between" mt={1}>
                        <MDButton
                          variant="outlined"
                          color={agent.status === "Active" ? "error" : "success"}
                        >
                          {agent.status === "Active" ? "해지하기" : "등록하기"}
                        </MDButton>
                        <MDButton
                          variant="outlined"
                          color="info"
                          onClick={() => openModal(agent)}
                        >
                          상세보기
                        </MDButton>
                      </MDBox>
                    </Card>
                  </Grid>
                ))}
              </Grid>
            </Card>
          </Grid>

          {/* Agent 등록/해지 버튼 */}
          <Grid item xs={12}>
            <MDBox display="flex" justifyContent="space-between">
              <MDButton variant="contained" color="info">
                새로운 Agent 등록
              </MDButton>
              <MDButton variant="contained" color="error">
                모든 Agent 해지
              </MDButton>
            </MDBox>
          </Grid>
        </Grid>
      </MDBox>
      <Footer />

      {/* 모달 창 */}
      <Modal open={isModalOpen} onClose={closeModal} aria-labelledby="agent-container-modal" aria-describedby="agent-container-description">
        <MDBox display="flex" alignItems="center" justifyContent="center" height="100vh">
          <Card sx={{ p: 3, maxWidth: 500, width: "90%" }}>
            <MDTypography variant="h5" fontWeight="medium" mb={2}>
              {selectedAgent ? `${selectedAgent.name} - 연결된 컨테이너 목록` : "Loading..."}
            </MDTypography>
            {selectedAgent && (
              <MDBox>
                {selectedAgent.containers.map((container, idx) => (
                  <MDTypography key={idx} variant="body1">
                    {container}
                  </MDTypography>
                ))}
              </MDBox>
            )}
            <MDBox mt={3} display="flex" justifyContent="flex-end">
              <MDButton variant="contained" color="secondary" onClick={closeModal}>
                닫기
              </MDButton>
            </MDBox>
          </Card>
        </MDBox>
      </Modal>
    </DashboardLayout>
  );
}

export default AgentControl;
