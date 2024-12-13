// components/Sidebar.js
import React from 'react';
import Sidenav from "examples/Sidenav";
import { useMaterialUIController } from "context";
import routes from "routes";

const Sidebar = () => {
  const [controller] = useMaterialUIController();
  const { sidenavColor } = controller;

  return (
    <Sidenav
      color={sidenavColor}
      brandName="HActiV | Dashboard"
      routes={routes}
    />
  );
};

export default Sidebar;
