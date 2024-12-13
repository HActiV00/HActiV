import React from "react";
import { useNavigate } from "react-router-dom"; // useNavigate import 추가
import Card from "@mui/material/Card";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDButton from "components/MDButton";
import profileIcon from "assets/images/profile_icon.png";

function UserProfileCard({ profile }) {
  const navigate = useNavigate(); // useNavigate 훅 설정

  const handleAccountSettingsClick = () => {
    navigate("/dashboard/mypage/account-setting"); // 계정 설정 페이지로 이동
  };

  return (
    <Card sx={{ p: 3, display: "flex", flexDirection: "column", alignItems: "center", minHeight: "200px" }}>
      <MDBox mb={2}>
        <img
          src={profile.image || profileIcon}
          alt="profile"
          style={{ width: "80px", height: "80px", borderRadius: "50%" }}
        />
      </MDBox>
      <MDTypography variant="h6" fontWeight="medium">
        {profile.name || "사용자 이름"}
      </MDTypography>
      <MDTypography variant="body2" color="textSecondary">
        {profile.email || "email@example.com"}
      </MDTypography>
      <MDTypography variant="caption" color="textSecondary" mt={1}>
        마지막 로그인: {profile.lastLogin || "N/A"}
      </MDTypography>

      <MDBox mt={2} width="100%" display="flex" flexDirection="column" alignItems="center">
        <MDButton variant="outlined" color="info" fullWidth style={{ marginBottom: "8px" }} onClick={handleAccountSettingsClick}>
          계정 설정
        </MDButton>
        <MDButton variant="contained" color="error" fullWidth>
          로그아웃
        </MDButton>
      </MDBox>
    </Card>
  );
}

export default UserProfileCard;
