import React from 'react';
import { Box, Typography, CircularProgress } from '@mui/material';

interface GaugeChartProps {
  value: number;
  title: string;
  color: string;
  unit?: string;
  precision?: number;
  lastUpdated?: string;
}

const GaugeChart: React.FC<GaugeChartProps> = ({ value, title, color, unit = '%', precision = 2, lastUpdated }) => {
  const displayValue = value.toFixed(precision);
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
      <Box sx={{ position: 'relative', display: 'inline-flex' }}>
        <CircularProgress
          variant="determinate"
          value={value > 100 ? 100 : value}
          size={80}
          thickness={8}
          style={{ color: color }}
        />
        <Box
          sx={{
            top: 0,
            left: 0,
            bottom: 0,
            right: 0,
            position: 'absolute',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <Typography variant="caption" component="div" style={{ color: color, fontWeight: 'bold' }}>
            {`${displayValue}${unit}`}
          </Typography>
        </Box>
      </Box>
      <Typography variant="body2" component="div" mt={1} textAlign="center">
        {title}
      </Typography>
      {lastUpdated && (
        <Typography variant="caption" component="div" mt={0.5} textAlign="center">
          Last updated: {lastUpdated}
        </Typography>
      )}
    </Box>
  );
};

export default React.memo(GaugeChart);

