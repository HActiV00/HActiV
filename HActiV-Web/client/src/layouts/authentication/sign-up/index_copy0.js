// react-router-dom components
import { Link } from "react-router-dom";

// @mui material components
import Card from "@mui/material/Card";
import Checkbox from "@mui/material/Checkbox";
import Grid from "@mui/material/Grid";

// Material Dashboard 2 React components
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDInput from "components/MDInput";
import MDButton from "components/MDButton";

// Authentication layout components
import DashboardLayout from "examples/LayoutContainers/DashboardLayout";
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import Footer from "examples/Footer";

function Cover() {
  return (
    <DashboardLayout>
      <DashboardNavbar />
      <Grid container justifyContent="center" alignItems="center" style={{ minHeight: "100vh" }}>
        
        {/* Left side Register form */}
        <Grid item xs={12} md={6} lg={4}>
          <Card
            style={{
              boxShadow: "0 8px 16px rgba(0, 0, 0, 0.1)",
              borderRadius: "12px",
              overflow: "hidden",
              paddingBottom: "60px",
              //borderTopRightRadius: "0px",  // 오른쪽 상단 곡률 제거
              //borderBottomRightRadius: "0px" // 오른쪽 하단 곡률 제거
            }}
          >
            <MDBox
              style={{
                backgroundColor: "#2C3E50",
                padding: "30px 20px",
                textAlign: "left",
                margin: "0 15px", // 양옆 여백
                borderRadius: "12px" // MDBox 곡률 유지
              }}
            >
              <MDTypography variant="h4" fontWeight="bold" color="white" mb={1}>
                Sign Up
              </MDTypography>
              <MDTypography variant="body2" color="white">
                Enter Email Address & Password
              </MDTypography>
            </MDBox>
            <MDBox pt={6} pb={5} px={3}>
              <MDBox component="form" role="form">
                <MDBox mb={3}>
                  <MDInput type="text" label="User Name" variant="standard" fullWidth />
                </MDBox>
                <MDBox mb={4}>
                  <MDInput type="text" label="Name" variant="standard" fullWidth />
                </MDBox>
                <MDBox mb={4}>
                  <MDInput type="email" label="Email Address" variant="standard" fullWidth />
                </MDBox>
                <MDBox mb={4}>
                  <MDInput type="password" label="Password" variant="standard" fullWidth />
                </MDBox>
                <MDBox display="flex" alignItems="center" mt={10} mb={2} ml={-1}>
                  <Checkbox />
                  <MDTypography
                    variant="body2"
                    fontWeight="regular"
                    color="#2C3E50"
                    sx={{ cursor: "pointer", userSelect: "none", ml: -1 }}
                  >
                    &nbsp;&nbsp;I agree to the&nbsp;
                  </MDTypography>
                  <MDTypography
                    component="a"
                    href="#"
                    variant="body2"
                    fontWeight="bold"
                    color="#2C3E50"
                    textGradient
                  >
                    Terms and Conditions
                  </MDTypography>
                </MDBox>
                <MDBox mt={5} mb={2}>
                  <MDButton variant="gradient" fullWidth style={{ backgroundColor: "#2C3E50", color: "white" }}>
                    Sign Up
                  </MDButton>
                </MDBox>
              </MDBox>
            </MDBox>
          </Card>
        </Grid>

        {/* Right side Sign In Guide Box */}
        <Grid item xs={12} md={6} lg={4}>
          <Card
            style={{
              boxShadow: "0 8px 16px rgba(0, 0, 0, 0.1)",
              borderRadius: "12px",
              overflow: "hidden",
              backgroundColor: "#2C3E50",
              color: "white",
              height: "160%",
              paddingTop: "170px",
              paddingBottom: "170px",
              borderTopLeftRadius: "0px",  // 왼쪽 상단 곡률 제거
              borderBottomLeftRadius: "0px" // 왼쪽 하단 곡률 제거
            }}
          >
            <MDBox textAlign="center" py={10} px={3}>
              <MDTypography variant="h4" fontWeight="bold" color="white" mb={1}>
                Hello, Friend
              </MDTypography>
			  <MDTypography variant="body2" color="white" mb={2}>
                Already have an account?
              </MDTypography>
              <MDTypography variant="body2" color="white" mb={3}>
                Enter your personal details and start your journey with us
              </MDTypography>
              <MDButton component={Link} to="/authentication/sign-in" variant="contained" color="inherit" style={{ backgroundColor: "white", color: "#2C3E50", fontWeight: "bold" }}>
                Sign In
              </MDButton>
            </MDBox>
          </Card>
        </Grid>
      </Grid>
      <Footer />
    </DashboardLayout>
  );
}

export default Cover;
