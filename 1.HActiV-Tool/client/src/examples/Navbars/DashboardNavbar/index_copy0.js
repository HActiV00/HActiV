import { useState, useEffect } from "react";
import { useLocation, Link } from "react-router-dom";
import PropTypes from "prop-types";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import IconButton from "@mui/material/IconButton";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import Icon from "@mui/material/Icon";
import MDBox from "components/MDBox";
import MDInput from "components/MDInput";
import Breadcrumbs from "examples/Breadcrumbs";
import NotificationItem from "examples/Items/NotificationItem";
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
  const [openMenu, setOpenMenu] = useState(false);
  const [openUserMenu, setOpenUserMenu] = useState(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const route = useLocation().pathname.split("/").slice(1);

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
  const handleOpenMenu = (event) => setOpenMenu(event.currentTarget);
  const handleCloseMenu = () => setOpenMenu(false);
  const handleOpenUserMenu = (event) => setOpenUserMenu(event.currentTarget);
  const handleCloseUserMenu = () => setOpenUserMenu(null);

  const handleLoginLogout = () => {
    setIsLoggedIn(!isLoggedIn);
    handleCloseUserMenu();
  };

  const renderMenu = () => (
    <Menu
      anchorEl={openMenu}
      anchorReference={null}
      anchorOrigin={{
        vertical: "bottom",
        horizontal: "left",
      }}
      open={Boolean(openMenu)}
      onClose={handleCloseMenu}
      sx={{ mt: 2 }}
    >
      <NotificationItem icon={<Icon>email</Icon>} title="Check new messages" />
      <NotificationItem icon={<Icon>podcasts</Icon>} title="Manage Podcast sessions" />
      <NotificationItem icon={<Icon>shopping_cart</Icon>} title="Payment successfully completed" />
    </Menu>
  );

  const renderUserMenu = () => (
    <Menu
      anchorEl={openUserMenu}
      anchorOrigin={{
        vertical: "bottom",
        horizontal: "right",
      }}
      keepMounted
      transformOrigin={{
        vertical: "top",
        horizontal: "right",
      }}
      open={Boolean(openUserMenu)}
      onClose={handleCloseUserMenu}
    >
      {isLoggedIn ? (
        <MenuItem onClick={handleLoginLogout}>
          Sign out
        </MenuItem>
      ) : (
        <MenuItem component={Link} to="/authentication/sign-in" onClick={handleCloseUserMenu}>
          Sign in
        </MenuItem>
      )}
    </Menu>
  );

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
      sx={(theme) => ({
        ...navbar(theme, { transparentNavbar, absolute, light, darkMode }),
        width: "100%",
        left: 0,
      })}
    >
      <Toolbar sx={(theme) => navbarContainer(theme)}>
        <MDBox color="inherit" mb={{ xs: 1, md: 0 }} sx={(theme) => navbarRow(theme, { isMini })}>
          <Breadcrumbs icon="home" title={route[route.length - 1]} route={route} light={light} />
        </MDBox>
        <MDBox sx={(theme) => navbarRow(theme, { isMini })}>
          <MDBox pr={1}>
            <MDInput label="Search here" />
          </MDBox>
          <MDBox color={light ? "white" : "inherit"}>
            <IconButton
              sx={navbarIconButton}
              size="small"
              disableRipple
              onClick={handleOpenUserMenu}
              aria-controls="user-menu"
              aria-haspopup="true"
            >
              <Icon sx={iconsStyle}>account_circle</Icon>
            </IconButton>
            {renderUserMenu()}
            <IconButton
              size="small"
              disableRipple
              color="inherit"
              sx={{
                ...navbarMobileMenu,
                display: "inline-flex",  // Always show the button
              }}
              onClick={handleMiniSidenav}
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
            >
              <Icon sx={iconsStyle}>settings</Icon>
            </IconButton>
            <IconButton
              size="small"
              disableRipple
              color="inherit"
              sx={navbarIconButton}
              aria-controls="notification-menu"
              aria-haspopup="true"
              variant="contained"
              onClick={handleOpenMenu}
            >
              <Icon sx={iconsStyle}>notifications</Icon>
            </IconButton>
            {renderMenu()}
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