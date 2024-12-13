import React, { useEffect, useState, useRef } from "react";
import { Dialog, DialogTitle, DialogContent, DialogActions } from "@mui/material";
import Card from "@mui/material/Card";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import GaugeChart from "react-gauge-chart";
import MDButton from "components/MDButton";
import useMediaQuery from "@mui/material/useMediaQuery";
import { Doughnut } from "react-chartjs-2";
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from "chart.js";

// Chart.js 요소 등록
ChartJS.register(ArcElement, Tooltip, Legend);

function EventsAndAlertsCard({ events }) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState(null);
  const isLargeScreen = useMediaQuery("(min-width:1200px)");
  const [chartWidth, setChartWidth] = useState(isLargeScreen ? 400 : 300);
  const chartContainerRef = useRef(null);

  // 상태 결정 함수
  const determineBoundaryLevel = () => {
    const highestSeverity = events.reduce((max, event) => 
      Math.max(max, getSeverityLevel(event.severity)), 0);
    if (highestSeverity === 0.25) return "정보";
    if (highestSeverity === 0.5) return "주의";
    if (highestSeverity === 0.75) return "위험";
    return "심각";
  };

  const getSeverityLevel = (severity) => {
    switch (severity) {
      case "INFO":
        return 0.25;
      case "MEDIUM":
        return 0.5;
      case "HIGH":
        return 0.75;
      case "CRITICAL":
        return 1;
      default:
        return 0.5;
    }
  };

  const relatedEvents = [...events]
    .sort((a, b) => b.count - a.count)
    .slice(0, 3);

  while (relatedEvents.length < 3) {
    relatedEvents.push({ title: "No Event", count: "-" });
  }

  const openModal = (event) => {
    setSelectedEvent(event);
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setSelectedEvent(null);
  };

  useEffect(() => {
    const observer = new ResizeObserver(() => {
      const newWidth = isLargeScreen ? 400 : 300;
      if (chartWidth !== newWidth) {
        setChartWidth(newWidth);
      }
    });
    if (chartContainerRef.current) {
      observer.observe(chartContainerRef.current);
    }

    return () => {
      if (chartContainerRef.current) observer.unobserve(chartContainerRef.current);
      observer.disconnect();
    };
  }, [isLargeScreen, chartWidth]);

  const data = {
    labels: ["Info", "Medium", "High", "Critical"],
    datasets: [
      {
        data: [10, 20, 30, 40],
        backgroundColor: ["#36A2EB", "#FFCE56", "#FF6384", "#FF4500"],
      },
    ],
  };

  const options = {
    responsive: true,
    plugins: {
      legend: {
        display: true,
        position: "left",
        align: "start",
      },
    },
  };

  return (
    <>
      <Card
        ref={chartContainerRef}
        sx={{
          position: "relative",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
        }}
      >
        <MDBox position="absolute" top={16} right={16}>
          <MDButton variant="text" color="dark" size="small" onClick={() => openModal(events)}>
            상세보기
          </MDButton>
        </MDBox>

        <MDBox p={2} flexGrow={1}>
          <MDTypography variant="h6" fontWeight="medium">
            이벤트 및 경고
          </MDTypography>
          {events && events.length > 0 ? (
            events.map((event, index) => (
              <MDBox key={index} my={2}>
                <MDTypography variant="subtitle2">
                  {event.title} ({event.severity})
                </MDTypography>
                <GaugeChart
                  id={`gauge-chart-${index}`}
                  nrOfLevels={4}
                  colors={["#00FF00", "#FFDD00", "#FF9900", "#FF0000"]}
                  arcWidth={0.3}
                  percent={getSeverityLevel(event.severity)}
                  needleColor="#345243"
                  textColor="#000000"
                  style={{
                    width: `${chartWidth}px`,
                    margin: "0 auto",
                  }}
                />
                <MDTypography variant="body2" mt={1}>
                  {event.description}
                </MDTypography>
              </MDBox>
            ))
          ) : (
            <MDTypography variant="body2" color="textSecondary">
              표시할 이벤트가 없습니다.
            </MDTypography>
          )}
        </MDBox>

        <MDBox p={2} mt={-2} borderTop="1px solid #ddd" />

        <MDBox display="flex" flexDirection="column" alignItems="center" gap={1} pb={2}>
          {relatedEvents.map((event, index) => (
            <MDButton
              key={index}
              variant="outlined"
              color="primary"
              size="small"
              sx={{
                width: "90%",
                margin: "4px 0",
              }}
              onClick={() => openModal(event)}
            >
              {event.title} - {event.count}
            </MDButton>
          ))}
        </MDBox>
      </Card>

      {/* 이벤트 상세 모달 */}
      <Dialog open={isModalOpen} onClose={closeModal} maxWidth="lg" fullWidth>
        <DialogTitle>
          <MDBox
            sx={{
              backgroundColor: "#2c3e50",
              color: "#fff",
              padding: "10px",
              borderRadius: "5px",
              textAlign: "center",
              fontSize: "1.5rem",
              fontWeight: "bold",
              animation: "fadeIn 0.5s ease-in-out",
            }}
          >
            경계 단계: {determineBoundaryLevel()}
          </MDBox>
        </DialogTitle>
        <DialogContent>
          <MDBox display="flex" flexDirection="column" alignItems="start">
            <MDTypography variant="h6" component="div" sx={{ marginBottom: "10px" }}>
              지난 48시간 동안의 이벤트
            </MDTypography>
            <Doughnut data={data} options={options} style={{ maxWidth: "400px" }} />
          </MDBox>
        </DialogContent>
        <DialogActions>
          <MDButton onClick={closeModal} color="primary">
            닫기
          </MDButton>
        </DialogActions>
      </Dialog>
    </>
  );
}

export default EventsAndAlertsCard;
