import React from 'react';
import { AppBar, Toolbar, Typography, Button, Box, IconButton } from '@mui/material';
import { Link } from 'react-router-dom';
import GitHubIcon from '@mui/icons-material/GitHub'; // GitHub 아이콘 import

const Header = () => {
  return (
    <AppBar position="sticky"> {/* sticky로 설정하여 스크롤 시에도 고정 */}
      <Toolbar>
        <Typography variant="h6">HACtiV</Typography>
        
        {/* 왼쪽 버튼들 */}
        <Button color="inherit" component={Link} to="/" sx={{ textTransform: 'none' }}>Home</Button>
        <Button color="inherit" component={Link} to="/features" sx={{ textTransform: 'none' }}>Docs</Button>
        <Button color="inherit" component={Link} to="/documentation" sx={{ textTransform: 'none' }}>Blog</Button>
        <Button color="inherit" component={Link} to="/pricing" sx={{ textTransform: 'none' }}>About</Button>
        <Button color="inherit" component={Link} to="/contact" sx={{ textTransform: 'none' }}>Contact</Button>

        {/* 우측 버튼들 */}
        <Box sx={{ flexGrow: 1 }} /> {/* 왼쪽 버튼들과 우측 버튼들 사이의 공간 확보 */}
        <IconButton
          color="inherit"
          href="https://github.com/HActiV00/HActiV"
          target="_blank" // 새 탭에서 열기
          sx={{ padding: 0 }} // padding 제거
        >
          <GitHubIcon /> {/* GitHub 아이콘 */}
        </IconButton>
        
        {/* 새 탭에서 Try HActiV 버튼 열기 */}
        <Button
          color="inherit"
          component="a"
          href="/dashboard"
          target="_blank" // 새 탭에서 열기
          rel="noopener noreferrer"
          sx={{ textTransform: 'none' }}
        >
          Try HActiV
        </Button>
      </Toolbar>
    </AppBar>
  );
};

export default Header;
