import React, { Suspense, lazy, useEffect, useMemo, useState, useRef, useCallback } from "react";
import { Grid, Skeleton, Select, MenuItem, FormControl, InputLabel, IconButton, Typography, Box, TextField, Button, useMediaQuery, Card, List, ListItem, ListItemText, LinearProgress } from "@mui/material";
import { Refresh, Block, Settings, Build, Delete, SaveAlt } from "@mui/icons-material";
import { useLocation, useNavigate } from "react-router-dom";

import MDBox from "components/MDBox";
import DashboardLayout from "examples/LayoutContainers/DashboardLayout";
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import Footer from "examples/Footer";
import GaugeChart from "./components/GaugeChart";

const ReportsLineChart = lazy(() => import("examples/Charts/LineCharts/ReportsLineChart"));

const GAUGE_COLORS = {
  cpu: "#FF6384",
  memory: "#36A2EB",
  disk: "#FFCE56",
  network_rx: "#4BC0C0",
  network_tx: "#9966FF"
};

function LazyLoadedChart({ Component, chartRef, toolName, chartId, lastUpdated, width, ...props }) {
  const navigate = useNavigate();

  const handleChartClick = () => {
    navigate(`/dashboard/detail?tool=${toolName}`);
  };

  return (
    <Suspense fallback={<Skeleton variant="rectangular" width={width} height={300} />}>
      <div onClick={handleChartClick} style={{ cursor: "pointer", height: "100%", width: width }}>
        <Component ref={chartRef} date={lastUpdated} {...props} />
      </div>
    </Suspense>
  );
}

export default function Dashboard() {
  const [dashboardData, setDashboardData] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedAgent, setSelectedAgent] = useState("default");
  const [selectedTool, setSelectedTool] = useState("all");
  const [selectedContainer, setSelectedContainer] = useState("all");
  const [filterDate, setFilterDate] = useState("");
  const [filterPeriod, setFilterPeriod] = useState("");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [lastUpdated, setLastUpdated] = useState({});
  const [hostMetrics, setHostMetrics] = useState({ cpu_cores: 0, cpu_usage: 0, memory_usage: 0, disk_usage: 0 });
  const [containerMetrics, setContainerMetrics] = useState({});
  const [inactiveContainers, setInactiveContainers] = useState({});
  const [lastUpdate, setLastUpdate] = useState({});
  const [prevEventTimes, setPrevEventTimes] = useState({});
  const [historicalData, setHistoricalData] = useState([]);
  const [isLoadingHistorical, setIsLoadingHistorical] = useState(false);
  const [viewMode, setViewMode] = useState('realtime');

  const chartRef = useRef(null);
  const location = useLocation();
  const navigate = useNavigate();
  const isSmallScreen = useMediaQuery("(max-width: 600px)");
  const wsRef = useRef(null);

  const toolTypes = useMemo(() => [
    "Systemcall",
    "Network_traffic",
    "Memory",
    "file_open",
    "delete",
    "log_file_open",
    "log_file_delete"
  ], []);

  const backgroundColors = {
    Systemcall: "rgba(75, 192, 192, 0.2)",
    "Network_traffic": "rgb(168, 74, 32, 0.2)",
    Memory: "rgba(54, 162, 235, 0.2)",
    file_open: "rgba(153, 102, 255, 0.2)",
    delete: "rgba(128, 0, 128, 0.2)",  
    log_file_open: "rgba(255, 205, 86, 0.2)",
    log_file_delete: "rgba(255, 99, 132, 0.2)"
  };

  const colorThemes = {
    Systemcall: "primary",
    "Network_traffic": "warning",
    Memory: "success",
    file_open: "info",
    delete: "secondary",  
    log_file_open: "info",
    log_file_delete: "error"
  };

  const borderColor = {
    Systemcall: "#ffffff",
    file_open: "#ffffff",
    delete: "#ffffff",
    Network_traffic: "#ffffff",
    Memory: "#ffffff",
    log_file_open: "#ffffff",
    log_file_delete: "#ffffff"
  };

  const fetchData = useCallback(async () => {
    try {
      const queryParams = new URLSearchParams(location.search);
      const tool = queryParams.get("tool") || "all";
      setSelectedTool(tool);
      
      const dashboardResponse = await fetch(`/api/dashboard?event_type=${tool}`);
      if (!dashboardResponse.ok) throw new Error(`HTTP error! status: ${dashboardResponse.status}`);
      let data = await dashboardResponse.json();

      const hostMetricsResponse = await fetch('/api/dashboard?event_type=HostMetrics');
      if (!hostMetricsResponse.ok) throw new Error(`HTTP error! status: ${hostMetricsResponse.status}`);
      const hostMetricsData = await hostMetricsResponse.json();

      const containerMetricsResponse = await fetch('/api/dashboard?event_type=ContainerMetrics');
      if (!containerMetricsResponse.ok) throw new Error(`HTTP error! status: ${containerMetricsResponse.status}`);
      const containerMetricsData = await containerMetricsResponse.json();

      const inactiveContainersResponse = await fetch('/api/dashboard?event_type=InactiveContainers');
      if (!inactiveContainersResponse.ok) throw new Error(`HTTP error! status: ${inactiveContainersResponse.status}`);
      const inactiveContainersData = await inactiveContainersResponse.json();
      
      const mostRecentTimestamp = new Date(Math.max(...data.map(event => new Date(event.timestamp).getTime())));
      const thirtyMinutesAgo = new Date(mostRecentTimestamp.getTime() - 30 * 60 * 1000);
      data = data.filter(event => 
        event.event_type !== 'HostMetrics' && 
        event.event_type !== 'ContainerMetrics' &&
        event.event_type !== 'InactiveContainers' &&
        new Date(event.timestamp) >= thirtyMinutesAgo
      );

      setDashboardData(data);
      setIsLoading(false);

      const newLastUpdated = {};
      [...toolTypes, 'all'].forEach(toolType => {
        const toolData = toolType === 'all' ? data : data.filter(item => item.event_type === toolType);
        if (toolData.length > 0) {
          const lastDataTime = new Date(Math.max(...toolData.map(event => new Date(event.timestamp).getTime())));
          newLastUpdated[toolType] = `${lastDataTime.getFullYear()}-${String(lastDataTime.getMonth() + 1).padStart(2, '0')}-${String(lastDataTime.getDate()).padStart(2, '0')} ${String(lastDataTime.getHours()).padStart(2, '0')}:${String(lastDataTime.getMinutes()).padStart(2, '0')}:${String(lastDataTime.getSeconds()).padStart(2, '0')}`;
        }
      });
      setLastUpdated(newLastUpdated);

      if (hostMetricsData.length > 0) {
        const latestHostMetrics = hostMetricsData[0];
        setHostMetrics({
          cpu_cores: parseFloat(latestHostMetrics.cpu_cores),
          cpu_usage: parseFloat(latestHostMetrics.cpu_usage),
          memory_usage: parseFloat(latestHostMetrics.memory_usage),
          disk_usage: parseFloat(latestHostMetrics.disk_usage)
        });
      }

      const newContainerMetrics = {};
      const newInactiveContainers = { ...inactiveContainersData };
      const newPrevEventTimes = { ...prevEventTimes };
      containerMetricsData.forEach(metric => {
        if (metric.container_name !== 'H') {
          const metricTime = new Date(metric.timestamp).getTime();
          const prevTime = newPrevEventTimes[metric.container_name] || 0;

          if (prevTime && metricTime - prevTime > 3 * 60 * 1000) {
            newInactiveContainers[metric.container_name] = true;
          } else {
            newContainerMetrics[metric.container_name] = {
              cpu_usage: parseFloat(metric.cpu_usage) * 100,
              memory_usage: parseFloat(metric.memory_usage),
              disk_usage: parseFloat(metric.disk_usage),
              network_rx: parseFloat(metric.rx_bytes) / (1024 * 1024),
              network_tx: parseFloat(metric.tx_bytes) / (1024 * 1024),
              lastUpdated: new Date(metric.timestamp).toLocaleString()
            };
            delete newInactiveContainers[metric.container_name];
          }

          newPrevEventTimes[metric.container_name] = metricTime;
        }
      });

      setContainerMetrics(newContainerMetrics);
      setInactiveContainers(newInactiveContainers);
      setPrevEventTimes(newPrevEventTimes);

      const newLastUpdate = {};
      containerMetricsData.forEach(metric => {
        newLastUpdate[metric.container_name] = new Date(metric.timestamp).getTime();
      });
      setLastUpdate(newLastUpdate);

      if (chartRef.current) chartRef.current.updateChartData(data);
    } catch (err) {
      console.error("Error fetching data:", err);
      setIsLoading(false);
    }
  }, [location.search, toolTypes, prevEventTimes]);

  useEffect(() => {
    fetchData();
    
    wsRef.current = new WebSocket('ws://hactiv-web-backend:8080/ws');
    
    wsRef.current.onopen = () => {
      console.log('WebSocket Connected');
    };

    wsRef.current.onmessage = (event) => {
      const newData = JSON.parse(event.data);
      console.log('Received WebSocket data:', newData);
      if (newData.event_type === 'HostMetrics') {
        setHostMetrics({
          cpu_cores: parseFloat(newData.cpu_cores),
          cpu_usage: parseFloat(newData.cpu_usage),
          memory_usage: parseFloat(newData.memory_usage),
          disk_usage: parseFloat(newData.disk_usage)
        });
        console.log('Updated HostMetrics:', newData);
      } else if (newData.event_type === 'ContainerMetrics') {
        if (newData.container_name !== 'H') {
          const metricTime = new Date(newData.timestamp).getTime();
          setPrevEventTimes(prev => {
            const prevTime = prev[newData.container_name] || 0;
            if (prevTime && metricTime - prevTime > 3 * 60 * 1000) {
              setInactiveContainers(prevInactive => ({
                ...prevInactive,
                [newData.container_name]: true
              }));
              setContainerMetrics(prevMetrics => {
                const newMetrics = { ...prevMetrics };
                delete newMetrics[newData.container_name];
                return newMetrics;
              });
            } else {
              setContainerMetrics(prevMetrics => ({
                ...prevMetrics,
                [newData.container_name]: {
                  cpu_usage: parseFloat(newData.cpu_usage) * 100,
                  memory_usage: parseFloat(newData.memory_usage),
                  disk_usage: parseFloat(newData.disk_usage),
                  network_rx: parseFloat(newData.rx_bytes) / (1024 * 1024),
                  network_tx: parseFloat(newData.tx_bytes) / (1024 * 1024),
                  lastUpdated: new Date(newData.timestamp).toLocaleString()
                }
              }));
              setInactiveContainers(prevInactive => {
                const newInactive = { ...prevInactive };
                delete newInactive[newData.container_name];
                return newInactive;
              });
            }
            return {
              ...prev,
              [newData.container_name]: metricTime
            };
          });
          setLastUpdate(prev => ({
            ...prev,
            [newData.container_name]: metricTime
          }));
        }
        console.log('Updated ContainerMetrics:', newData);
      } else if (newData.event_type === 'InactiveContainers') {
        setInactiveContainers(newData);
      } else {
        setDashboardData(prevData => {
          const updatedData = [newData, ...prevData.slice(0, 99)];
          setLastUpdated(prev => ({
            ...prev,
            [newData.event_type]: new Date(newData.timestamp).toLocaleString(),
            all: new Date(newData.timestamp).toLocaleString()
          }));
          setLastUpdate(prev => ({
            ...prev,
            [newData.container_name]: new Date(newData.timestamp).getTime()
          }));
          return updatedData;
        });
      }
    };

    wsRef.current.onerror = (error) => {
      console.error('WebSocket Error:', error);
    };

    wsRef.current.onclose = () => {
      console.log('WebSocket Disconnected');
    };

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [fetchData]);

  useEffect(() => {
    const checkInactiveContainers = () => {
      const now = new Date().getTime();
      setContainerMetrics(prevMetrics => {
        const newMetrics = { ...prevMetrics };
        Object.entries(newMetrics).forEach(([containerName, metrics]) => {
          const lastUpdateTime = new Date(metrics.lastUpdated).getTime();
          if (now - lastUpdateTime > 3 * 60 * 1000) {
            delete newMetrics[containerName];
            setInactiveContainers(prev => ({
              ...prev,
              [containerName]: true
            }));
          }
        });
        return newMetrics;
      });
    };

    checkInactiveContainers();
    const intervalId = setInterval(checkInactiveContainers, 60000);

    return () => clearInterval(intervalId);
  }, []);

  const handleAgentChange = (event) => setSelectedAgent(event.target.value);
  const handleToolChange = (event) => {
    const tool = event.target.value;
    setSelectedTool(tool);
    navigate(`?tool=${tool}`);
  };
  const handleContainerChange = (event) => setSelectedContainer(event.target.value);
  const handlePeriodChange = (event) => {
    setFilterPeriod(event.target.value);
    setFilterDate("");
  };
  const handleResetSelection = () => {
    setFilterDate("");
    setFilterPeriod("");
    setStartDate("");
    setEndDate("");
  };
  const handleStartDateChange = (event) => {
    const newStartDate = new Date(event.target.value);
    setStartDate(newStartDate);
    if (endDate) {
      fetchHistoricalData(newStartDate, endDate);
    }
  };
  const handleEndDateChange = (event) => {
    const newEndDate = new Date(event.target.value);
    setEndDate(newEndDate);
    if (startDate) {
      fetchHistoricalData(startDate, newEndDate);
    }
  };

  const fetchHistoricalData = async (startDate, endDate) => {
    setIsLoadingHistorical(true);
    try {
      const response = await fetch(`/api/dashboard?event_type=${selectedTool}&start_time=${startDate.toISOString()}&end_time=${endDate.toISOString()}`);
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      const data = await response.json();
      setHistoricalData(data);
    } catch (err) {
      console.error("Error fetching historical data:", err);
    } finally {
      setIsLoadingHistorical(false);
    }
  };

  const checkContainerActivity = useCallback((metrics) => {
    const now = new Date().getTime();
    const lastUpdateTime = new Date(metrics.lastUpdated).getTime();
    return now - lastUpdateTime <= 3 * 60 * 1000;
  }, []);

  const handleDeleteContainer = async (containerName) => {
    try {
      setContainerMetrics(prevMetrics => {
        const newMetrics = { ...prevMetrics };
        delete newMetrics[containerName];
        return newMetrics;
      });
      
      setInactiveContainers(prevInactive => {
        const newInactive = { ...prevInactive };
        delete newInactive[containerName];
        return newInactive;
      });

      setDashboardData(prevData => 
        prevData.filter(event => event.container_name !== containerName)
      );

      await fetch(`/api/dashboard/containers/${containerName}`, {
        method: 'DELETE'
      });

    } catch (error) {
      console.error('Error deleting container:', error);
    }
  };

  const handleSaveContainerData = async (containerName) => {
    try {
      const containerEvents = dashboardData.filter(
        event => event.container_name === containerName
      );

      const headers = ['event_type', 'timestamp', 'command', 'arguments'];
      const csvContent = [
        headers.join(','),
        ...containerEvents.map(event => 
          [
            event.event_type,
            event.timestamp,
            event.command,
            event.arguments
          ].join(',')
        )
      ].join('\n');

      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `${containerName}_events.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

    } catch (error) {
      console.error('Error saving container data:', error);
    }
  };

  const filteredData = useMemo(() => {
    const now = new Date().getTime();
    let data = dashboardData.filter(item => 
      item.event_type !== 'HostMetrics' && 
      item.event_type !== 'ContainerMetrics' &&
      (lastUpdate[item.container_name] && now - lastUpdate[item.container_name] <= 5 * 60 * 1000)
    );

    if (selectedTool !== "all") {
      data = data.filter((item) => item.event_type.toLowerCase().includes(selectedTool.toLowerCase()));
    }

    if (startDate && endDate) {
      const start = new Date(startDate);
      const end = new Date(endDate);
      data = data.filter((item) => {
        const itemDate = new Date(item.timestamp);
        return itemDate >= start && itemDate <= end;
      });
    } else if (filterDate) {
      const filterDateStr = filterDate.slice(0, 10);
      data = data.filter((item) => item.timestamp.slice(0, 10) === filterDateStr);
    }

    if (filterPeriod) {
      const currentTime = new Date();
      const periodInMs = {
        "30m": 30 * 60 * 1000,
        "1h": 60 * 60 * 1000,
        "6h": 6 * 60 * 60 * 1000,
        "12h": 12 * 60 * 60 * 1000,
        "24h": 24 * 60 * 60 * 1000,
        "48h": 48 * 60 * 60 * 1000,
      }[filterPeriod];

      data = data.filter((item) => {
        const eventTime = new Date(item.timestamp);
        return currentTime - eventTime <= periodInMs;
      });
    }

    return data;
  }, [dashboardData, selectedTool, startDate, endDate, filterDate, filterPeriod, lastUpdate]);

  const chartData = useMemo(() => {
    const data = [...dashboardData, ...historicalData];
    const eventCounts = {};
    const containerCounts = {};
    const timelineData = {};
    const toolTimelineData = {};
    const allEventsTimelineData = {};

    toolTypes.forEach((tool) => {
      toolTimelineData[tool] = {};
    });

    data.forEach((item) => {
      if (item.event_type !== 'HostMetrics' && item.event_type !== 'ContainerMetrics' && item.event_type !== 'InactiveContainers') {
        eventCounts[item.event_type] = (eventCounts[item.event_type] || 0) + 1;
        containerCounts[item.container_name] = (containerCounts[item.container_name] || 0) + 1;

        const date = new Date(item.timestamp);
        const timeKey = `${date.getHours().toString().padStart(2, "0")}:${date.getMinutes().toString().padStart(2, "0")}:${date.getSeconds().toString().padStart(2, "0")}`;

        timelineData[date.getHours()] = (timelineData[date.getHours()] || 0) + 1;

        // All Events Timeline data
        allEventsTimelineData[timeKey] = (allEventsTimelineData[timeKey] || 0) + 1;

        if (item.event_type && toolTimelineData[item.event_type]) {
          toolTimelineData[item.event_type][timeKey] = (toolTimelineData[item.event_type][timeKey] || 0) + 1;
        }
      }
    });

    return { eventCounts, containerCounts, timelineData, toolTimelineData, allEventsTimelineData };
  }, [dashboardData, historicalData, toolTypes]);

  if (isLoading || isLoadingHistorical) {
    return (
      <DashboardLayout>
        <MDBox display="flex" justifyContent="center" alignItems="center" height="100vh">
          <LinearProgress />
        </MDBox>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <DashboardNavbar />
      <MDBox display="flex" gap={2} mb={4} flexDirection={isSmallScreen ? "column" : "row"} sx={{ flexWrap: isSmallScreen ? "wrap" : "nowrap" }}>
        <FormControl fullWidth sx={{ maxWidth: isSmallScreen ? "48%" : "100%", height: 48 }}>
          <InputLabel id="agent-select-label">{isSmallScreen ? <Settings /> : "Select Agent"}</InputLabel>
          <Select
            labelId="agent-select-label"
            id="agent-select"
            value={selectedAgent}
            onChange={handleAgentChange}
            sx={{ height: 48, minHeight: 48 }}
          >
            <MenuItem value="default">Default</MenuItem>
            <MenuItem value="additional">Additional</MenuItem>
          </Select>
        </FormControl>

        <FormControl fullWidth sx={{ maxWidth: isSmallScreen ? "48%" : "100%", height: 48 }}>
          <InputLabel id="tool-select-label">{isSmallScreen ? <Build /> : "Select Monitoring Tool"}</InputLabel>
          <Select
            labelId="tool-select-label"
            id="tool-select"
            value={selectedTool}
            onChange={handleToolChange}
            sx={{ height: 48, minHeight: 48 }}
          >
            <MenuItem value="all">All Tools</MenuItem>
            {toolTypes.map((tool) => (
              <MenuItem key={tool} value={tool}>{tool}</MenuItem>
            ))}
          </Select>
        </FormControl>

        <FormControl fullWidth sx={{ maxWidth: isSmallScreen ? "48%" : "100%", height: 48 }}>
          <InputLabel id="container-select-label">Select Container</InputLabel>
          <Select
            labelId="container-select-label"
            id="container-select"
            value={selectedContainer}
            onChange={handleContainerChange}
            sx={{ height: 48, minHeight: 48 }}
          >
            <MenuItem value="all">All Containers</MenuItem>
            {Object.keys(containerMetrics).map((containerName) => (
              <MenuItem key={containerName} value={containerName}>{containerName}</MenuItem>
            ))}
          </Select>
        </FormControl>

        <TextField
          fullWidth
          label="Start Date"
          type="date"
          value={startDate}
          onChange={handleStartDateChange}
          InputLabelProps={{ shrink: true }}
          sx={{ maxWidth: "48%", height: 48 }}
        />
        <TextField
          fullWidth
          label="End Date"
          type="date"
          value={endDate}
          onChange={handleEndDateChange}
          InputLabelProps={{ shrink: true }}
          sx={{ maxWidth: "48%", height: 48 }}
        />

        <Box display="flex" gap={1} width="100%" sx={{ height: 48 }}>
          {["30m", "1h", "6h", "12h", "24h", "48h"].map((period) => (
            <Button
              key={period}
              fullWidth
              size="small"
              variant={filterPeriod === period ? "contained" : "outlined"}
              onClick={() => handlePeriodChange({ target: { value: period } })}
              sx={{
                height: 48,
                color: filterPeriod === period ? "#ffffff" : "#ffffff",
                borderColor: filterPeriod === period ? "#354e6e" : "#1A73E8",
                backgroundColor: filterPeriod === period ? "#354e6e" : "#1A73E8",
                textTransform: "none",
                "&:hover": {
                  backgroundColor: filterPeriod === period ? "#354e6e" : "#354e6e",
                  borderColor: "#354e6e",
                },
              }}
            >
              {period.replace("m", " M").replace("h", " H")}
            </Button>
          ))}
        </Box>

        <IconButton onClick={fetchData} sx={{ color: "primary.main", fontSize: 29, height: 48 }}>
          <Refresh />
        </IconButton>

        <IconButton onClick={handleResetSelection} sx={{ color: "primary.main", fontSize: 25, height: 48 }}>
          <Block />
        </IconButton>
      </MDBox>

      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card sx={{ p: 2, mb: 2 }}>
            <Typography variant="h6" gutterBottom>Host Metrics</Typography>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6} md={4}>
                <GaugeChart value={hostMetrics.cpu_usage} title={`CPU Usage (${hostMetrics.cpu_cores} cores)`} color={GAUGE_COLORS.cpu} unit="%" precision={2} />
              </Grid>
              <Grid item xs={12} sm={6} md={4}>
                <GaugeChart value={hostMetrics.memory_usage} title="Memory Usage" color={GAUGE_COLORS.memory} unit="%" precision={2} />
              </Grid>
              <Grid item xs={12} sm={6} md={4}>
                <GaugeChart value={hostMetrics.disk_usage} title="Disk Usage" color={GAUGE_COLORS.disk} unit="%" precision={2} />
              </Grid>
            </Grid>
          </Card>
        </Grid>

        {Object.entries(containerMetrics)
          .filter(([_, metrics]) => checkContainerActivity(metrics))
          .map(([containerName, metrics]) => (
            <Grid item xs={12} key={containerName}>
              <Card sx={{ p: 2, mb: 2 }}>
                <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                  <Typography variant="h6">
                    {containerName}
                    <Typography variant="caption" component="span" ml={2}>
                      Last updated: {metrics.lastUpdated}
                    </Typography>
                  </Typography>
                  <IconButton onClick={() => handleSaveContainerData(containerName)} title="Save container data">
                    <SaveAlt />
                  </IconButton>
                  <IconButton onClick={() => handleDeleteContainer(containerName)} title="Delete container" color="error">
                    <Delete />
                  </IconButton>
                </Box>
                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6} md={2.4}>
                    <GaugeChart value={metrics.cpu_usage} title="CPU Usage" color={GAUGE_COLORS.cpu} unit="%" precision={2} />
                  </Grid>
                  <Grid item xs={12} sm={6} md={2.4}>
                    <GaugeChart value={metrics.memory_usage} title="Memory Usage" color={GAUGE_COLORS.memory} unit="%" precision={2} />
                  </Grid>
                  <Grid item xs={12} sm={6} md={2.4}>
                    <GaugeChart value={metrics.disk_usage} title="Disk Usage" color={GAUGE_COLORS.disk} unit="%" precision={2} />
                  </Grid>
                  <Grid item xs={12} sm={6} md={2.4}>
                    <GaugeChart value={metrics.network_rx} title="Network RX" color={GAUGE_COLORS.network_rx} unit=" MB" precision={2} />
                  </Grid>
                  <Grid item xs={12} sm={6} md={2.4}>
                    <GaugeChart value={metrics.network_tx} title="Network TX" color={GAUGE_COLORS.network_tx} unit=" MB" precision={2} />
                  </Grid>
                </Grid>
              </Card>
            </Grid>
          ))}

        {/* Inactive Containers Section */}
        <Grid item xs={12}>
          <Card sx={{ p: 2, mb: 2 }}>
            <Typography variant="h6" gutterBottom>Inactive Containers</Typography>
            <List>
              {Object.entries(inactiveContainers).map(([containerName, data]) => (
                <ListItem key={containerName}>
                  <ListItemText 
                    primary={containerName}
                    secondary={`Last updated: ${data.lastUpdated}`} 
                  />
                  <IconButton onClick={() => handleSaveContainerData(containerName)} title="Save container data">
                    <SaveAlt />
                  </IconButton>
                  <IconButton onClick={() => handleDeleteContainer(containerName)} title="Delete container" color="error">
                    <Delete />
                  </IconButton>
                </ListItem>
              ))}
            </List>
          </Card>
        </Grid>

        {viewMode === 'realtime' ? (
          <>
            {/* First row */}
            <Grid container spacing={2} sx={{ paddingLeft: '32px', paddingTop: '16px', paddingBottom: '16px' }}>
              {['all', 'Systemcall', 'Network_traffic', 'Memory'].map((tool, index) => (
                <Grid item xs={12} md={3} key={tool} sx={{ 
                  paddingLeft: index === 0 ? 0 : '8px',
                  paddingRight: index === 3 ? 0 : '8px'
                }}>
                  <Box sx={{ height: '300px', marginBottom: 0, width: '100%' }}>
                    <LazyLoadedChart
                      Component={ReportsLineChart}
                      chartRef={chartRef}
                      toolName={tool === 'all' ? 'All Events' : tool}
                      chartId={index}
                      color={tool === 'all' ? 'primary' : colorThemes[tool]}
                      title={`${tool === 'all' ? 'All Events' : tool} Timeline`}
                      chart={{
                        labels: Object.keys(tool === 'all' ? chartData.allEventsTimelineData : (chartData.toolTimelineData[tool] || {})).sort(),
                        datasets: {
                          label: `${tool === 'all' ? 'All Events' : tool} Events`,
                          data: Object.values(tool === 'all' ? chartData.allEventsTimelineData : (chartData.toolTimelineData[tool] || {})),
                          backgroundColor: tool === 'all' ? "rgba(75, 192, 192, 0.2)" : backgroundColors[tool],
                          borderColor: tool === 'all' ? "#ffffff" : borderColor[tool],
                        },
                      }}
                      lastUpdated={lastUpdated[tool] || ''}
                      width="100%"
                    />
                  </Box>
                </Grid>
              ))}
            </Grid>

            {/* Second row */}
            <Grid container spacing={2} sx={{ paddingLeft: '32px', paddingTop: '16px', paddingBottom: '16px' }}>
              {['file_open', 'delete', 'log_file_open', 'log_file_delete'].map((tool, index) => (
                <Grid item xs={12} md={3} key={tool} sx={{ 
                  paddingLeft: index === 0 ? 0 : '8px',
                  paddingRight: index === 3 ? 0 : '8px'
                }}>
                  <Box sx={{ height: '300px', marginBottom: 0, width: '100%' }}>
                    <LazyLoadedChart
                      Component={ReportsLineChart}
                      chartRef={chartRef}
                      toolName={tool}
                      chartId={index + 4}
                      color={colorThemes[tool]}
                      title={`${tool} Timeline`}
                      chart={{
                        labels: Object.keys(chartData.toolTimelineData[tool] || {}).sort(),
                        datasets: {
                          label: `${tool} Events`,
                          data: Object.values(chartData.toolTimelineData[tool] || {}),
                          backgroundColor: backgroundColors[tool],
                          borderColor: borderColor[tool],
                        },
                      }}
                      lastUpdated={lastUpdated[tool] || ''}
                      width="100%"
                    />
                  </Box>
                </Grid>
              ))}
            </Grid>
          </>
        ) : (
          // Historical view remains unchanged
          <Grid item xs={12} sx={{ padding: 2 }}>
            <Box sx={{ height: '300px', marginBottom: 0 }}>
              <LazyLoadedChart
                Component={ReportsLineChart}
                chartRef={chartRef}
                toolName={selectedTool}
                chartId={1}
                color={colorThemes[selectedTool]}
                title={`${selectedTool} Event Timeline (Filtered)`}
                chart={{
                  labels: filteredData.map(item => new Date(item.timestamp).toLocaleTimeString()),
                  datasets: {
                    label: `${selectedTool} Events`,
                    data: filteredData.map(item => 1),
                    backgroundColor: backgroundColors[selectedTool],
                    borderColor: borderColor[selectedTool],
                  },
                }}
                lastUpdated={lastUpdated[selectedTool] || ''}
                width="100%"
              />
            </Box>
          </Grid>
        )}
      </Grid>

      <Footer />
    </DashboardLayout>
  );
}

