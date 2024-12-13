// src/layouts/profile/components/TeamSpaceCard.js

import React, { useState } from "react";
import { useNavigate } from "react-router-dom"; // useNavigate 추가
import Card from "@mui/material/Card";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import IconButton from "@mui/material/IconButton";
import Badge from "@mui/material/Badge";
import GroupIcon from "@mui/icons-material/Group";
import MailOutlineIcon from "@mui/icons-material/MailOutline";
import Avatar from "@mui/material/Avatar";
import megaphoneIcon from "assets/images/megaphone_icon.png";

function TeamSpaceCard() {
  const navigate = useNavigate(); // 네비게이트 훅 설정
  const [isNoticeOpen, setIsNoticeOpen] = useState(false);
  const [isUnread, setIsUnread] = useState(true);

  const handleNoticeClick = () => {
    setIsNoticeOpen(!isNoticeOpen);
    setIsUnread(false);
  };

  const handleManageClick = () => {
    navigate("/group-info"); // 그룹 정보 페이지로 이동
  };

  return (
    <Card sx={{ minHeight: "300px" }}>
      <MDBox p={2}>
        <MDBox display="flex" justifyContent="space-between" alignItems="center" mb={2}>
          <MDTypography variant="h6" fontWeight="medium">
            팀 스페이스
          </MDTypography>
          <MDBox display="flex" alignItems="center">
            <IconButton onClick={handleManageClick}>
              <GroupIcon sx={{ color: "navy" }} />
              <MDTypography variant="caption" sx={{ color: "navy", ml: 0.5 }}>
                관리
              </MDTypography>
            </IconButton>
            <IconButton>
              <MailOutlineIcon sx={{ color: "navy", ml: 1 }} />
            </IconButton>
          </MDBox>
        </MDBox>
        <MDTypography variant="body2" color="text" mt={1} mb={2}>
          구성원 목록과 그룹 정보를 확인하세요.
        </MDTypography>
        <MDBox display="flex" mt={1} mb={2} gap={1}>
          {[...Array(5)].map((_, index) => (
            <Avatar key={index} sx={{ width: 36, height: 36 }}>
              U
            </Avatar>
          ))}
        </MDBox>
        <MDBox borderTop="1px solid #ddd" pt={3} mt={2}>
          <MDTypography variant="h6" fontWeight="medium" mb={1}>
            그룹 공지
          </MDTypography>
          <MDTypography variant="body2" color="text" mb={2}>
            그룹 공지 내용을 여기에 표시합니다.
          </MDTypography>
          <MDBox
            mt={2}
            position="relative"
            display="flex"
            alignItems="center"
            onClick={handleNoticeClick}
            sx={{
              cursor: "pointer",
              width: "100%",
            }}
          >
            <img src={megaphoneIcon} alt="megaphone icon" style={{ width: 20, height: 20, marginRight: 8 }} />
            <Badge
              color="error"
              variant="dot"
              overlap="rectangular"
              anchorOrigin={{ vertical: "top", horizontal: "right" }}
              invisible={!isUnread}
            >
              <MDBox
                display="flex"
                alignItems="center"
                sx={{
                  border: "1px solid #ddd",
                  borderRadius: 1,
                  padding: "8px 72px",
                  width: "100%",
                  boxSizing: "border-box",
                }}
              >
                <Avatar sx={{ width: 24, height: 24, mr: 1 }}>U</Avatar>
                <MDTypography variant="button" fontWeight="medium">
                  Username
                </MDTypography>
              </MDBox>
            </Badge>
          </MDBox>
          {isNoticeOpen && (
            <MDBox mt={2} p={2} bgcolor="#f9f9f9" borderRadius="4px">
              <MDTypography variant="body2">
                여기에서 그룹 공지의 세부 내용을 볼 수 있습니다. 공지를 펼쳐서 더 많은 정보를 확인하세요.
              </MDTypography>
            </MDBox>
          )}
        </MDBox>
      </MDBox>
    </Card>
  );
}

export default TeamSpaceCard;
