import React, { useState } from 'react';
import { Avatar, Menu, MenuItem, IconButton } from '@mui/material';
import { Person, Settings, ExitToApp } from '@mui/icons-material';
import MDBox from "components/MDBox";

export default function UserProfileDropdown() {
  const [anchorEl, setAnchorEl] = useState(null);

  const handleClick = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <MDBox>
      <IconButton onClick={handleClick} size="large" edge="end" color="inherit">
        <Avatar alt="User Avatar" src="/path-to-avatar-image.jpg" />
      </IconButton>
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
      >
        <MenuItem onClick={handleClose}>
          <Person fontSize="small" style={{ marginRight: '8px' }} />
          Profile
        </MenuItem>
        <MenuItem onClick={handleClose}>
          <Settings fontSize="small" style={{ marginRight: '8px' }} />
          Settings
        </MenuItem>
        <MenuItem onClick={handleClose}>
          <ExitToApp fontSize="small" style={{ marginRight: '8px' }} />
          Logout
        </MenuItem>
      </Menu>
    </MDBox>
  );
}