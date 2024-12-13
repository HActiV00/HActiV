// src/layouts/AccountSettings/index.js

import React, { useState } from "react";
import { Grid, TextField, Button, Card, Avatar } from "@mui/material";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import DashboardLayout from "examples/LayoutContainers/DashboardLayout";
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import Footer from "examples/Footer";

function AccountSettings() {
  const [username, setUsername] = useState(""); // 사용자 이름 상태
  const [password, setPassword] = useState(""); // 새 비밀번호 상태
  const [confirmPassword, setConfirmPassword] = useState(""); // 비밀번호 확인 상태
  const [profileImage, setProfileImage] = useState(null); // 프로필 이미지 상태

  // 프로필 이미지 변경 핸들러
  const handleProfileImageChange = (e) => {
    setProfileImage(URL.createObjectURL(e.target.files[0]));
  };

  // 양식 제출 핸들러 (백엔드 연동 필요)
  const handleFormSubmit = (e) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      alert("Passwords do not match.");
      return;
    }
    // 여기에 제출 시의 처리 로직 (예: API 호출)
    alert("Profile has been updated.");
  };

  return (
    <DashboardLayout>
      <DashboardNavbar />
      <MDBox mt={4} mb={3} px={3}>
        <Grid container>
          <Grid item xs={12}>
            <Card
              sx={{
                width: "100%",
                padding: 3,
              }}
            >
              <MDBox mt={2} ml={2}>
                <MDTypography variant="h6" fontWeight="medium" color="dark">
                  Account Setting
                </MDTypography>
              </MDBox>
              <MDBox p={3} component="form" onSubmit={handleFormSubmit}>
                {/* 프로필 이미지 수정 */}
                <MDBox display="flex" flexDirection="column" alignItems="center" mb={3}>
                  <Avatar
                    src={profileImage}
                    alt="Profile Image"
                    sx={{ width: 100, height: 100, mb: 2 }}
                  />
                  <Button variant="outlined" component="label" sx={{ color: "text.primary" }}>
                    Update Profile Picture
                    <input
                      type="file"
                      hidden
                      accept="image/*"
                      onChange={handleProfileImageChange}
                    />
                  </Button>
                </MDBox>

                {/* 사용자 이름 수정 */}
                <TextField
                  label="Username"
                  fullWidth
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  margin="normal"
                />

                {/* 비밀번호 변경 */}
                <TextField
                  label="New Password"
                  type="password"
                  fullWidth
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  margin="normal"
                />
                <TextField
                  label="Confirm Password"
                  type="password"
                  fullWidth
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  margin="normal"
                />

                {/* 저장 버튼 */}
                <MDBox mt={3} display="flex" justifyContent="center">
                  <Button
                    type="submit"
                    variant="contained"
                    color="primary"
                    sx={{ color: "#FFFFFF" }} // 글씨 색을 흰색으로 설정
                  >
                    Save Changes
                  </Button>
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

export default AccountSettings;
