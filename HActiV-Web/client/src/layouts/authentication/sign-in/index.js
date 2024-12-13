// 1028 - SignIn 가안 완성
// react-router-dom components
import { Link } from "react-router-dom";

// @mui material components
import Card from "@mui/material/Card";
import Checkbox from "@mui/material/Checkbox";
import Grid from "@mui/material/Grid";
import useMediaQuery from "@mui/material/useMediaQuery";

// Material Dashboard 2 React components
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDInput from "components/MDInput";
import MDButton from "components/MDButton";

// Authentication layout components
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import Footer from "examples/Footer";

function SignIn() {
  const isMobile = useMediaQuery("(max-width:600px)"); // 모바일 화면 감지 (600px 이하)

  return (
    <>
      <DashboardNavbar />
      <MDBox maxWidth="lg" mx="auto" px={3}> {/* 창 너비에 맞춰 중앙 정렬과 여백 유지 */}
        <Grid container justifyContent="center" alignItems="center" style={{ minHeight: "100vh" }}>
          {/* Left side description text - 모바일 화면에서는 숨김 */}
          {!isMobile && (
            <Grid item xs={12} md={6} lg={6} style={{ padding: "20px" }}>
              <MDBox>
                <MDTypography variant="h4" fontWeight="medium" color="textPrimary" mb={2}>
                  <span role="img" aria-label="shield">🛡️</span> Comprehensive Visibility
                </MDTypography>
                <MDTypography variant="button" color="textSecondary" mb={5}>
                  Gain complete visibility into user behavior and application performance, providing real-time insights across your cloud environment.
                </MDTypography>
                <MDTypography variant="h6" fontWeight="medium" color="textPrimary" mb={2}>
                  <span role="img" aria-label="lock">🔒</span> Enhanced Security
                </MDTypography>
                <MDTypography variant="button" color="textSecondary" mb={5}>
                  Monitor every action in real-time to detect and respond to potential threats, ensuring robust runtime security.
                </MDTypography>
                <MDTypography variant="h6" fontWeight="medium" color="textPrimary" mb={2}>
                  <span role="img" aria-label="chart">📊</span> Insightful Analytics
                </MDTypography>
                <MDTypography variant="button" color="textSecondary" mb={5}>
                  Visualize complex data with intuitive dashboards, enabling you to analyze trends, behaviors, and anomalies effectively.
                </MDTypography>
                <MDTypography variant="h6" fontWeight="medium" color="textPrimary" mb={2}>
                  <span role="img" aria-label="rocket">🚀</span> Scalable and Adaptable
                </MDTypography>
                <MDTypography variant="button" color="textSecondary">
                  Built to scale with your needs, our tool adapts effortlessly to evolving cloud infrastructure and security demands, helping you stay ahead.
                </MDTypography>
              </MDBox>
            </Grid>
          )}

          {/* Right side Sign In form */}
          <Grid item xs={12} md={6} lg={6} style={{ padding: "20px" }}>
            <Card
              style={{
                boxShadow: "0 8px 16px rgba(0, 0, 0, 0.1)",
                borderRadius: "12px",
                overflow: "hidden",
                paddingBottom: "80px",
                width: "100%",
                margin: "0 auto",
              }}
            >
              <MDBox
                style={{
                  backgroundColor: "#2C3E50",
                  padding: "30px 20px",
                  textAlign: "left",
                  margin: "0 15px",
                  borderRadius: "12px 12px 0 0",
                }}
              >
                <MDTypography variant="h4" fontWeight="medium" color="white" mb={1}>
                  Sign In
                </MDTypography>
                <MDTypography variant="body2" color="white">
                  Enter Email Address & Password
                </MDTypography>
              </MDBox>
              <MDBox pt={6} pb={6} px={3}>
                <MDBox component="form" role="form">
                  <MDBox mb={8}>
                    <MDInput type="email" label="Email Address" variant="standard" fullWidth />
                  </MDBox>
                  <MDBox mb={2}>
                    <MDInput type="password" label="Password" variant="standard" fullWidth />
                  </MDBox>

                  {/* Keep me signed in checkbox */}
                  <MDBox display="flex" alignItems="center" mb={4} ml={-1}>
                    <Checkbox />
                    <MDTypography
                      variant="button"
                      fontWeight="bold"
                      color="#2C3E50"
                      textGradient
                      sx={{ cursor: "pointer", userSelect: "none", ml: -1 }}
                    >
                      &nbsp;&nbsp;Remember me
                    </MDTypography>
                  </MDBox>

                  <MDBox mt={12} mb={4}>
                    <MDButton variant="gradient" fullWidth style={{ backgroundColor: "#2C3E50", color: "white" }}>
                      Login
                    </MDButton>
                  </MDBox>
                  <MDBox mt={2} textAlign="center">
                    <MDTypography variant="button" color="textSecondary" sx={{ fontSize: "1rem" }}>
                      Don't have an account?{" "}
                      <MDTypography
                        component={Link}
                        to="/dashboard/sign-up"
                        variant="button"
                        fontWeight="medium"
                        color="#2C3E50"
                        textGradient
                        sx={{ fontSize: "1rem" }}
                      >
                        Sign Up
                      </MDTypography>
                    </MDTypography>
                  </MDBox>
                </MDBox>
              </MDBox>
            </Card>
          </Grid>
        </Grid>
      </MDBox>
      <Footer />
    </>
  );
}

export default SignIn;
