import React from "react";
import PropTypes from "prop-types";
import { useNavigate } from "react-router-dom";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import typography from "assets/theme/base/typography";

function Footer({ company, links }) {
  const navigate = useNavigate();
  const { name } = company;
  const { size } = typography;

  const handleHActiVClick = () => {
    alert("HActiV 팀 응원 부탁드립니다☘️☘️");
  };

  const handleLinkClick = (path) => {
    navigate(path);
  };

  const renderLinks = () =>
    links.map((link) => {
      let path = "/";
      if (link.name === "Docs") path = "/docs";
      else if (link.name === "Blog") path = "/blog";
      else if (link.name === "About Us") path = "/about";
      else if (link.name === "License") path = "/license";

      return (
        <MDBox key={link.name} component="li" px={2} lineHeight={1}>
          <MDTypography
            variant="button"
            fontWeight="regular"
            color="text"
            onClick={() => handleLinkClick(path)}
            style={{ cursor: "pointer" }}
          >
            {link.name}
          </MDTypography>
        </MDBox>
      );
    });

  return (
    <MDBox
      width="100%"
      display="flex"
      flexDirection={{ xs: "column", lg: "row" }}
      justifyContent="space-between"
      alignItems="center"
      px={1.5}
    >
      <MDBox
        display="flex"
        justifyContent="center"
        alignItems="center"
        flexWrap="wrap"
        color="text"
        fontSize={size.sm}
        px={1.5}
      >
        &copy; {new Date().getFullYear()},
        <MDTypography
          variant="button"
          fontWeight="medium"
          onClick={handleHActiVClick}
          style={{ cursor: "pointer" }}
        >
          &nbsp;{name}&nbsp;
        </MDTypography>
        All Rights Reserved.
      </MDBox>
      <MDBox
        component="ul"
        sx={({ breakpoints }) => ({
          display: "flex",
          flexWrap: "wrap",
          alignItems: "center",
          justifyContent: "center",
          listStyle: "none",
          mt: 3,
          mb: 0,
          p: 0,

          [breakpoints.up("lg")]: {
            mt: 0,
          },
        })}
      >
        {renderLinks()}
      </MDBox>
    </MDBox>
  );
}

Footer.defaultProps = {
  company: { name: "HActiV" },
  links: [
    { name: "Docs" },
    { name: "Blog" },
    { name: "About Us" },
    { name: "License" },
  ],
};

Footer.propTypes = {
  company: PropTypes.shape({
    name: PropTypes.string,
  }),
  links: PropTypes.arrayOf(PropTypes.shape({
    name: PropTypes.string,
  })),
};

export default Footer;
