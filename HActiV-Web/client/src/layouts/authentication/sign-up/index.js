import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import axios from "axios"; // Axios로 API 호출

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

function SignUp() {
  const isMobile = useMediaQuery("(max-width:600px)");
  const [form, setForm] = useState({ username: "", name: "", email: "", password: "" });
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const validateForm = () => {
    const { username, name, email, password } = form;

    if (username.length === 0 || username.length > 50) {
      alert("Username must be between 1 and 50 characters.");
      return false;
    }

    if (name.length === 0 || name.length > 50) {
      alert("Name must be between 1 and 50 characters.");
      return false;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      alert("Please enter a valid email address.");
      return false;
    }

    if (password.length < 8) {
      alert("Password must be at least 8 characters long.");
      return false;
    }

    return true;
  };

  const handleSignUp = async () => {
    if (!validateForm()) return;

    try {
      const response = await axios.post(
        "/api/signup",
        form,
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
      alert("회원가입이 완료되었습니다.");
      navigate("/authentication/sign-in");
    } catch (err) {
      setError(err.response?.data?.error || "Sign up failed. Please try again.");
    }
  };

  return (
    <>
      <DashboardNavbar />
      <MDBox maxWidth="lg" mx="auto" px={3}>
        <Grid container justifyContent="center" alignItems="center" style={{ minHeight: "100vh" }}>
          {!isMobile && (
            <Grid item xs={12} md={6} lg={6} style={{ padding: "20px" }}>
              <MDBox>
                <MDTypography variant="h4" fontWeight="bold" style={{ color: "#2C3E50" }} mb={2}>
                  Hello, Friend
                </MDTypography>
                <MDTypography variant="body2" style={{ color: "#2C3E50" }} mb={3}>
                  Our product effortlessly adjusts to your needs, boosting efficiency and simplifying your tasks.
                </MDTypography>
                <MDTypography variant="h6" fontWeight="bold" style={{ color: "#2C3E50" }} mb={2}>
                  Built to last
                </MDTypography>
                <MDTypography variant="body2" style={{ color: "#2C3E50" }} mb={3}>
                  Experience unmatched durability that goes above and beyond with lasting investment.
                </MDTypography>
                <MDTypography variant="h6" fontWeight="bold" style={{ color: "#2C3E50" }} mb={2}>
                  Great user experience
                </MDTypography>
                <MDTypography variant="body2" style={{ color: "#2C3E50" }} mb={3}>
                  Integrate our product into your routine with an intuitive and easy-to-use interface.
                </MDTypography>
                <MDTypography variant="h6" fontWeight="bold" style={{ color: "#2C3E50" }} mb={2}>
                  Innovative functionality
                </MDTypography>
                <MDTypography variant="body2" style={{ color: "#2C3E50" }}>
                  Stay ahead with features that set new standards, addressing your evolving needs better than the rest.
                </MDTypography>
              </MDBox>
            </Grid>
          )}

          <Grid item xs={12} md={6} lg={6} style={{ padding: "20px" }}>
            <Card style={{ boxShadow: "0 8px 16px rgba(0, 0, 0, 0.1)", borderRadius: "12px", overflow: "hidden", paddingBottom: "60px" }}>
              <MDBox style={{ backgroundColor: "#2C3E50", padding: "30px 20px", textAlign: "left", margin: "0 15px", borderRadius: "12px" }}>
                <MDTypography variant="h4" fontWeight="bold" style={{ color: "white" }} mb={1}>Sign Up</MDTypography>
                <MDTypography variant="body2" style={{ color: "white" }}>Create an account</MDTypography>
              </MDBox>
              <MDBox pt={6} pb={5} px={3}>
                {error && (
                  <MDTypography variant="body2" color="error" mb={2}>{error}</MDTypography>
                )}
                <MDBox component="form" role="form">
                  <MDBox mb={5}>
                    <MDInput type="text" label="User Name" name="username" variant="standard" fullWidth onChange={handleChange} />
                  </MDBox>
                  <MDBox mb={5}>
                    <MDInput type="text" label="Name" name="name" variant="standard" fullWidth onChange={handleChange} />
                  </MDBox>
                  <MDBox mb={5}>
                    <MDInput type="email" label="Email Address" name="email" variant="standard" fullWidth onChange={handleChange} />
                  </MDBox>
                  <MDBox mb={3}>
                    <MDInput type="password" label="Password" name="password" variant="standard" fullWidth onChange={handleChange} />
                  </MDBox>
                  <MDBox display="flex" alignItems="center" mt={3} mb={8} ml={-0.5}>
                    <Checkbox />
                    <MDTypography
                      variant="body2"
                      fontWeight="regular"
                      style={{ color: "#2C3E50" }}
                      sx={{ cursor: "pointer", userSelect: "none", ml: -1 }}
                    >
                      &nbsp;&nbsp;I agree to the&nbsp;
                    </MDTypography>
                    <MDTypography component="a" href="#" variant="body2" fontWeight="bold" style={{ color: "#2C3E50" }} textGradient>
                      Terms and Conditions
                    </MDTypography>
                  </MDBox>
                  <MDBox mt={5} mb={2}>
                    <MDButton variant="gradient" fullWidth style={{ backgroundColor: "#2C3E50", color: "white" }} onClick={handleSignUp}>
                      Sign Up
                    </MDButton>
                  </MDBox>
                  <MDBox textAlign="center" mt={2}>
                    <MDTypography variant="body2" style={{ color: "#2C3E50" }}>
                      Already have an account?{" "}
                      <MDTypography component={Link} to="/authentication/sign-in" variant="body2" fontWeight="bold" style={{ color: "#2C3E50" }}>
                        Sign In
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

export default SignUp;
