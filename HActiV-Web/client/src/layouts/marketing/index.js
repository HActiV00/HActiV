import React from 'react';
import { Container, Grid, Box, Typography, Button } from '@mui/material';
import Header from './Header'; // 새로 만든 Header 컴포넌트 import
import Footer from "examples/Footer";
import { Link } from 'react-router-dom';

const HomePage = () => {
  return (
    <div>
      <Header /> {/* 상단바 컴포넌트 사용 */}

      <Container>
        {/* Hero Section */}
        <Box sx={{ my: 4, textAlign: 'center' }}>
          <Typography variant="h2">Welcome to Your Monitoring Dashboard</Typography>
          <Typography variant="h5">Real-time insights into your cloud infrastructure</Typography>
          <Button variant="contained" sx={{ mt: 2 }} component={Link} to="/dashboard">Get Started</Button>
        </Box>

        {/* Features Section */}
        <Grid container spacing={4}>
          {["Feature 1", "Feature 2", "Feature 3"].map((feature, index) => (
            <Grid item xs={12} sm={4} key={index}>
              <Box sx={{ border: '1px solid #ccc', borderRadius: 2, p: 2 }}>
                <Typography variant="h6">{feature}</Typography>
                <Typography>Short description of {feature}.</Typography>
                <Button variant="outlined">Learn More</Button>
              </Box>
            </Grid>
          ))}
        </Grid>

        {/* Footer */}
        <Footer />
      </Container>
    </div>
  );
};

export default HomePage;
