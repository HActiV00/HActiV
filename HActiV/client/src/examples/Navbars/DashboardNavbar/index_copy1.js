import React, { useState, useEffect } from "react";
import PropTypes from "prop-types";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import IconButton from "@mui/material/IconButton";
import Icon from "@mui/material/Icon";
import MDBox from "components/MDBox";
import MDInput from "components/MDInput";
import Breadcrumbs from "examples/Breadcrumbs";
import {
  navbar,
  navbarContainer,
  navbarRow,
  navbarIconButton,
  navbarMobileMenu,
} from "examples/Navbars/DashboardNavbar/styles";
import {
  useMaterialUIController,
  setTransparentNavbar,
  setMiniSidenav,
  setOpenConfigurator,
} from "context";

function DashboardNavbar({ absolute, light, isMini }) {
  const [navbarType, setNavbarType] = useState();
  const [controller, dispatch] = useMaterialUIController();
  const { miniSidenav, transparentNavbar, fixedNavbar, openConfigurator, darkMode } = controller;
  const route = window.location.pathname.split("/").slice(1);

  useEffect(() => {
    if (fixedNavbar) {
      setNavbarType("sticky");
    } else {
      setNavbarType("static");
    }

    function handleTransparentNavbar() {
      setTransparentNavbar(dispatch, (fixedNavbar && window.scrollY === 0) || !fixedNavbar);
    }

    window.addEventListener("scroll", handleTransparentNavbar);
    handleTransparentNavbar();

    return () => window.removeEventListener("scroll", handleTransparentNavbar);
  }, [dispatch, fixedNavbar]);

  const handleMiniSidenav = () => setMiniSidenav(dispatch, !miniSidenav);
  const handleConfiguratorOpen = () => setOpenConfigurator(dispatch, !openConfigurator);

  const handleClick = (event, excludeIcon) => {
    if (!excludeIcon) {
      alert("서비스 준비중입니다.");
    }
  };

  const iconsStyle = ({ palette: { dark, white, text }, functions: { rgba } }) => ({
    color: () => {
      let colorValue = light || darkMode ? white.main : dark.main;

      if (transparentNavbar && !light) {
        colorValue = darkMode ? rgba(text.main, 0.6) : text.main;
      }

      return colorValue;
    },
  });

  return (
    <AppBar
      position={absolute ? "absolute" : navbarType}
      color="inherit"
      sx={(theme) => navbar(theme, { transparentNavbar, absolute, light, darkMode })}
    >
      <Toolbar sx={(theme) => navbarContainer(theme)}>
        <MDBox color="inherit" mb={{ xs: 1, md: 0 }} sx={(theme) => navbarRow(theme, { isMini })}>
          <Breadcrumbs icon="home" title={route[route.length - 1]} route={route} light={light} />
        </MDBox>
        <MDBox sx={(theme) => navbarRow(theme, { isMini })}>
          <MDBox pr={1}>
            <MDInput 
                label="Search here" 
                />
          </MDBox>
          <MDBox color={light ? "white" : "inherit"}>
            <IconButton
              sx={navbarIconButton}
              size="small"
              disableRipple
              onClick={(event) => handleClick(event, false)}
            >
              <Icon sx={iconsStyle}>account_circle</Icon>
            </IconButton>
            <IconButton
              size="small"
              disableRipple
              color="inherit"
              sx={navbarMobileMenu}
              onClick={handleMiniSidenav}
              onClickCapture={(event) => handleClick(event, true)}
            >
              <Icon sx={iconsStyle} fontSize="medium">
                {miniSidenav ? "menu_open" : "menu"}
              </Icon>
            </IconButton>
            <IconButton
              size="small"
              disableRipple
              color="inherit"
              sx={navbarIconButton}
              onClick={handleConfiguratorOpen}
              onClickCapture={(event) => handleClick(event, true)}
            >
              <Icon sx={iconsStyle}>settings</Icon>
            </IconButton>
            <IconButton
              size="small"
              disableRipple
              color="inherit"
              sx={navbarIconButton}
              onClick={(event) => handleClick(event, false)}
            >
              <Icon sx={iconsStyle}>notifications</Icon>
            </IconButton>
          </MDBox>
        </MDBox>
      </Toolbar>
    </AppBar>
  );
}

DashboardNavbar.defaultProps = {
  absolute: false,
  light: false,
  isMini: false,
};

DashboardNavbar.propTypes = {
  absolute: PropTypes.bool,
  light: PropTypes.bool,
  isMini: PropTypes.bool,
};

export default DashboardNavbar;