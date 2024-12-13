//examples/LayoutContainers/DashboardLayout/index.js
//10-28 우측 대시보드 패딩 여백 수정
//10-28 사이드바와 대시보드 사이 간격 수정

/**
=========================================================
* Material Dashboard 2 React - v2.2.0
=========================================================

* Product Page: https://www.creative-tim.com/product/material-dashboard-react
* Copyright 2023 Creative Tim (https://www.creative-tim.com)

Coded by www.creative-tim.com

 =========================================================

* The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
*/

import { useEffect } from "react";

// react-router-dom components
import { useLocation } from "react-router-dom";

// prop-types is a library for typechecking of props.
import PropTypes from "prop-types";

// Material Dashboard 2 React components
import MDBox from "components/MDBox";

// Material Dashboard 2 React context
import { useMaterialUIController, setLayout } from "context";

function DashboardLayout({ children }) {
  const [controller, dispatch] = useMaterialUIController();
  const { miniSidenav } = controller;
  const { pathname } = useLocation();

  useEffect(() => {
    setLayout(dispatch, "dashboard");
  }, [pathname, dispatch]); // Added dispatch to dependency array to resolve ESLint warning

  return (
    <MDBox
      sx={({ breakpoints, transitions, functions: { pxToRem } }) => ({
        p: 1, // 10-28 대시보드 우측 패딩 여백 수정(변경 3 -> 1)
        position: "relative",

        [breakpoints.up("xl")]: {
          marginLeft: miniSidenav ? pxToRem(80) : pxToRem(250), // 사이드바와 대시보드 사이 간격 수정 (변경 120->80, 274->250)
          transition: transitions.create(["margin-left", "margin-right"], {
            easing: transitions.easing.easeInOut,
            duration: transitions.duration.standard,
          }),
        },
      })}
    >
      {children}
    </MDBox>
  );
}

// Typechecking props for the DashboardLayout
DashboardLayout.propTypes = {
  children: PropTypes.node.isRequired,
};

export default DashboardLayout;
