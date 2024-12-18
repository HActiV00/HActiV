// 이벤트 타임라인 하얀 박스 세로 크기 줄이기
// 타임라인 색깔 박스 세로 크기 줄이기
// pt: 하얀박스의 글자 위로 올리기, pb: 하얀박스의 아래 여백, px: 하얀박스의 좌우

import React, { useRef, useMemo, useEffect } from "react";
import PropTypes from "prop-types";
import { Line } from "react-chartjs-2";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from "chart.js";
import Card from "@mui/material/Card";
import Divider from "@mui/material/Divider";
import Icon from "@mui/material/Icon";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import configs from "examples/Charts/LineCharts/ReportsLineChart/configs";

// Register chart.js components
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler);

function ReportsLineChart({ color, title, description, date, chart, tooltipData, bgColor }) {
  const chartRef = useRef(null);
  const { data, options } = configs(chart.labels || [], chart.datasets || {}, tooltipData, bgColor); // Pass bgColor

  const updateChartData = (newData) => {
    const chartInstance = chartRef.current;

    if (chartInstance) {
      chartInstance.data.labels = newData.labels;
      chartInstance.data.datasets[0].data = newData.datasets.data;
      chartInstance.update();
    }
  };

  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const response = await fetch("/api/dashboard");
        const newData = await response.json();
        updateChartData(newData);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    }, 10000);

    return () => clearInterval(interval);
  }, []);

  return ( //이벤트 타임라인 하얀 박스 세로 크기 줄이기 {100%} -> {90%}
    <Card sx={{ height: "90%" }}> 
      <MDBox padding="1rem">
        {useMemo(
          () => (
            <MDBox
              variant="gradient"
              bgColor={color}
              borderRadius="lg"
              coloredShadow={color}
              py={2}
              pr={0.5}
              mt={-5}
              height="10rem" // 타임라인 색깔 박스 세로 크기 줄이기
            >
              <Line ref={chartRef} data={data} options={options} />
            </MDBox>
          ),
          [data, options, color] // pt: 하얀박스의 글자 위로 올리기, pb: 하얀박스의 아래 여백, px: 하얀박스의 좌우
        )}  
        <MDBox pt={1} pb={4} px={1}> 
          <MDTypography variant="h6" textTransform="capitalize">
            {title}
          </MDTypography>
          <MDTypography component="div" variant="button" color="text" fontWeight="light">
            {description}
          </MDTypography>
          <Divider />
          <MDBox display="flex" alignItems="center">
            <MDTypography variant="button" color="text" lineHeight={1} sx={{ mt: 0, mr: 0.5 }}>
              <Icon>schedule</Icon>
            </MDTypography>
            <MDTypography variant="button" color="text" fontWeight="light">
              {date}
            </MDTypography>
          </MDBox>
        </MDBox>
      </MDBox>
    </Card>
  );
}

ReportsLineChart.defaultProps = {
  color: "info",
  description: "",
  tooltipData: {},
  bgColor: "rgba(255, 255, 255, 0.8)", // Default chart background color (optional)
};

ReportsLineChart.propTypes = {
  color: PropTypes.oneOf(["primary", "secondary", "info", "success", "warning", "error", "dark"]),
  title: PropTypes.string.isRequired,
  description: PropTypes.oneOfType([PropTypes.string, PropTypes.node]),
  date: PropTypes.string.isRequired,
  chart: PropTypes.objectOf(PropTypes.oneOfType([PropTypes.array, PropTypes.object])).isRequired,
  tooltipData: PropTypes.object,
  bgColor: PropTypes.string, // Allow dynamic chart background color
};

export default ReportsLineChart;
