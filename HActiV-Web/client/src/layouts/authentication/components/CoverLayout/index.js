import React from "react";
import { Link } from "react-router-dom";
import { Card, Checkbox } from "@mui/material";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDInput from "components/MDInput";
import MDButton from "components/MDButton";
import DashboardLayout from "examples/LayoutContainers/DashboardLayout"; 
import DashboardNavbar from "examples/Navbars/DashboardNavbar"; 
import Sidenav from "examples/Sidenav"; 
import Footer from "examples/Footer"; 
import bgImage from "assets/images/bg-sign-up-cover.jpeg"; // 배경 이미지 경로

const routes = [
  { type: "collapse", name: "Dashboard", icon: "dashboard", route: "/dashboard" },
  { type: "collapse", name: "Tables", icon: "table_chart", route: "/tables" },
  { type: "collapse", name: "Notifications", icon: "notifications", route: "/notifications" },
  { type: "collapse", name: "Profile", icon: "person", route: "/profile" },
  { type: "collapse", name: "Sign In", icon: "login", route: "/authentication/sign-in" },
  { type: "collapse", name: "Sign Up", icon: "assignment", route: "/authentication/sign-up" },
];

function SignUp() {
  return (
    <DashboardLayout>
      <DashboardNavbar />
      <Sidenav color="info" routes={routes} brandName="HActiV | Sign Up" showSettings={false} />
      
      <MDBox display="flex" justifyContent="center" alignItems="center" height="100vh" style={{ backgroundImage: `url(${bgImage})`, backgroundSize: "cover" }}>
        <Card>
          <MDBox variant="gradient" bgColor="info" borderRadius="lg" coloredShadow="success" mx={2} mt={-3} p={3} mb={1} textAlign="center">
            <MDTypography variant="h4" fontWeight="medium" color="white" mt={1}>
              Join us to explore new experiences
            </MDTypography>
            <MDTypography display="block" variant="button" color="white" my={1}>
              Enter your email and password to register
            </MDTypography>
          </MDBox>
          <MDBox pt={4} pb={3} px={3}>
            <MDBox component="form" role="form">
              <MDBox mb={2}>
                <MDInput type="text" label="Name" variant="standard" fullWidth />
              </MDBox>
              <MDBox mb={2}>
                <MDInput type="email" label="Email" variant="standard" fullWidth />
              </MDBox>
              <MDBox mb={2}>
                <MDInput type="password" label="Password" variant="standard" fullWidth />
              </MDBox>
              <MDBox display="flex" alignItems="center" ml={-1}>
                <Checkbox />
                <MDTypography variant="button" fontWeight="regular" color="text" sx={{ cursor: "pointer", userSelect: "none", ml: -1 }}>
                  &nbsp;&nbsp;I agree to the&nbsp;
                </MDTypography>
                <MDTypography component="a" href="#" variant="button" fontWeight="bold" color="info" textGradient>
                  Terms and Conditions
                </MDTypography>
              </MDBox>
              <MDBox mt={4} mb={1}>
                <MDButton variant="gradient" color="info" fullWidth>
                  Sign Up
                </MDButton>
              </MDBox>
              <MDBox mt={3} mb={1} textAlign="center">
                <MDTypography variant="button" color="text">
                  Already have an account?{" "}
                  <MDTypography component={Link} to="/authentication/sign-in" variant="button" color="info" fontWeight="medium" textGradient>
                    Sign In
                  </MDTypography>
                </MDTypography>
              </MDBox>
            </MDBox>
          </MDBox>
        </Card>
      </MDBox>
      <Footer />
    </DashboardLayout>
  );
}

export default SignUp;
