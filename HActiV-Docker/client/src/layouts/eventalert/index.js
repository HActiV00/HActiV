import { useState, useEffect, useCallback } from "react";
import Drawer from "@mui/material/Drawer";
import Grid from "@mui/material/Grid";
import Card from "@mui/material/Card";
import Icon from "@mui/material/Icon";
import IconButton from "@mui/material/IconButton";
import Checkbox from "@mui/material/Checkbox";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDButton from "components/MDButton";
import Footer from "examples/Footer";
import DashboardLayout from "examples/LayoutContainers/DashboardLayout";
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import TextField from "@mui/material/TextField";
import Pagination from "@mui/material/Pagination";
import dayjs from "dayjs";
import { Snackbar, Alert } from '@mui/material';

function EventAlert() {
  const [selectedAlert, setSelectedAlert] = useState(null);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [agentType, setAgentType] = useState("All");
  const [filter, setFilter] = useState("All");
  const [timeFilter, setTimeFilter] = useState("All");
  const [sortOrder, setSortOrder] = useState("newest");
  const [searchText, setSearchText] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [selectedAlerts, setSelectedAlerts] = useState([]);
  const [selectAll, setSelectAll] = useState(false);
  const [favorites, setFavorites] = useState(new Set());
  const [endDate, setEndDate] = useState("");
  const [alerts, setAlerts] = useState({ critical: [], high: [], medium: [], low: [] });
  const [alertOpen, setAlertOpen] = useState(false);
  const [alertSeverity, setAlertSeverity] = useState('info');
  const [alertMessage, setAlertMessage] = useState('');

  const alertsPerPage = 10;
  const indexOfLastAlert = currentPage * alertsPerPage;
  const indexOfFirstAlert = indexOfLastAlert - alertsPerPage;

  const handleAlertClick = (alert) => {
    setSelectedAlert(alert);
    setIsDrawerOpen(true);
  };

  const closeDrawer = () => {
    setIsDrawerOpen(false);
    setSelectedAlert(null);
  };

  const handleAgentTypeChange = (event) => setAgentType(event.target.value);
  const handleFilterChange = (event) => setFilter(event.target.value);
  const handleTimeFilterChange = (event) => setTimeFilter(event.target.value);
  const handleSortOrderChange = (event) => setSortOrder(event.target.value);
  const handleSearchChange = (event) => setSearchText(event.target.value);
  const handlePageChange = (event, page) => setCurrentPage(page);
  const handleEndDateChange = (event) => setEndDate(event.target.value);

  const handleSelectAllChange = (event) => {
    setSelectAll(event.target.checked);
    setSelectedAlerts(event.target.checked ? getFilteredAlerts().map(alert => alert.id) : []);
  };

  const handleSelectAlertChange = (id) => {
    setSelectedAlerts((prevSelected) =>
      prevSelected.includes(id) ? prevSelected.filter((alertId) => alertId !== id) : [...prevSelected, id]
    );
  };

  const handleFavoriteToggle = (id) => {
    setFavorites((prevFavorites) => {
      const updatedFavorites = new Set(prevFavorites);
      updatedFavorites.has(id) ? updatedFavorites.delete(id) : updatedFavorites.add(id);
      return updatedFavorites;
    });
  };

  const getAlertPriority = (type) => {
    switch (type) {
      case "low":
        return 1;
      case "medium":
        return 2;
      case "high":
        return 3;
      case "critical":
        return 4;
      default:
        return 0;
    }
  };

  const getFilteredAlerts = () => {
    let filtered = [...alerts.critical, ...alerts.high, ...alerts.medium, ...alerts.low];

    if (agentType !== "All") filtered = filtered.filter((alert) => alert.container_name === agentType);
    if (filter !== "All") filtered = filtered.filter((alert) => alert.severity.toLowerCase() === filter.toLowerCase());
    if (timeFilter !== "All") {
      const now = dayjs();
      filtered = filtered.filter((alert) => {
        const alertTime = dayjs(alert.timestamp);
        if (timeFilter === "15m") return now.diff(alertTime, "minute") <= 15;
        if (timeFilter === "30m") return now.diff(alertTime, "minute") <= 30;
        if (timeFilter === "1h") return now.diff(alertTime, "hour") <= 1;
        if (timeFilter === "3h") return now.diff(alertTime, "hour") <= 3;
        return true;
      });
    }

    if (endDate) {
      const end = dayjs(endDate);
      filtered = filtered.filter((alert) => dayjs(alert.timestamp).isBefore(end));
    }

    if (sortOrder === "newest") {
      filtered.sort((a, b) => dayjs(b.timestamp).diff(dayjs(a.timestamp)));
    } else if (sortOrder === "oldest") {
      filtered.sort((a, b) => dayjs(a.timestamp).diff(dayjs(b.timestamp)));
    } else if (sortOrder === "lowest_priority") {
      filtered.sort((a, b) => getAlertPriority(a.severity) - getAlertPriority(b.severity));
    } else if (sortOrder === "highest_priority") {
      filtered.sort((a, b) => getAlertPriority(b.severity) - getAlertPriority(a.severity));
    }

    if (searchText) {
      filtered = filtered.filter((alert) =>
        alert.event_type.toLowerCase().includes(searchText.toLowerCase()) || 
        alert.message.toLowerCase().includes(searchText.toLowerCase())
      );
    }

    return filtered.slice(indexOfFirstAlert, indexOfLastAlert);
  };

  const getAlertColor = (type) => {
    switch (type) {
      case "medium":
        return "#FFEB3B"; // Yellow
      case "high":
        return "#FB8C00"; // Orange
      case "critical":
        return "#F44336"; // Red
      default:
        return "#757575"; // Default gray
    }
  };

  const checkForAlerts = useCallback((data) => {
    if (data.critical.length > 0) {
      setAlertSeverity('error');
      setAlertMessage(`${data.critical.length} critical alert(s) detected. Immediate attention required!`);
      setAlertOpen(true);
    } else if (data.high.length > 0) {
      setAlertSeverity('warning');
      setAlertMessage(`${data.high.length} high severity alert(s) detected. Please investigate.`);
      setAlertOpen(true);
    } else if (data.medium.length > 0) {
      setAlertSeverity('info');
      setAlertMessage(`${data.medium.length} medium severity alert(s) detected.`);
      setAlertOpen(true);
    }
  }, []);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch('/api/alert?event_type=all');
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        const data = await response.json();
        setAlerts(data);
        checkForAlerts(data);
      } catch (error) {
        console.error("Failed to fetch data:", error);
      }
    };

    fetchData();
  }, [checkForAlerts]);

  const filteredAlerts = getFilteredAlerts();
  const totalAlerts = alerts.critical.length + alerts.high.length + alerts.medium.length + alerts.low.length;
  const totalPages = Math.ceil(totalAlerts / alertsPerPage);

  const maxDate = dayjs().format("YYYY-MM-DD");
  const minDate = dayjs().subtract(3, "year").format("YYYY-MM-DD");

  return (
    <DashboardLayout>
      <DashboardNavbar />
      <MDBox mt={2} mb={2} display="flex" justifyContent="center">
        <Grid container justifyContent="center">
          <Grid item xs={12} sm={10} md={10} lg={10}>
            <Card sx={{ padding: 2 }}>
              {/* Alert Timeline Title */}
              <MDBox display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                <MDTypography variant="h5" fontWeight="bold">Alert Timeline</MDTypography>
              </MDBox>

              {/* All Alerts Counter and Alert Counts by Severity */}
              <MDBox display="flex" alignItems="center" mb={2}>
                <MDTypography variant="h3" fontWeight="bold" color="textSecondary">{totalAlerts}</MDTypography>
                <MDTypography variant="subtitle1" color="textSecondary" sx={{ ml: 1 }}>All Alerts</MDTypography>
                
                {/* Risk Level Counts */}
                <MDBox display="flex" alignItems="center" ml={2}>
                  {['medium', 'high', 'critical'].map((severity) => {
                    const count = alerts[severity].length;
                    return (
                      <MDBox key={severity} display="flex" alignItems="center" mr={2}>
                        <MDTypography variant="body2" fontWeight="bold" color={getAlertColor(severity)} mr={0.5}>
                          {severity.charAt(0).toUpperCase() + severity.slice(1)}:
                        </MDTypography>
                        <MDTypography variant="body2" color="textSecondary">
                          {count}
                        </MDTypography>
                      </MDBox>
                    );
                  })}
                </MDBox>

                {/* Date Filter */}
                <TextField
                  label="Date Filter"
                  type="date"
                  value={endDate}
                  onChange={handleEndDateChange}
                  InputLabelProps={{ shrink: true }}
                  inputProps={{
                    min: minDate,
                    max: maxDate,
                  }}
                  sx={{ ml: 'auto', mr: 3 }}
                />

                <MDTypography variant="caption" color="textSecondary">
                  {dayjs().format("YYYY-MM-DD HH:mm:ss")}
                </MDTypography>
              </MDBox>

              {/* Filter Controls */}
              <Grid container spacing={2} mb={2}>
                <Grid item xs={12} sm={2}>
                  <FormControl fullWidth>
                    <InputLabel style={{ fontSize: "0.9rem", textAlign: "center" }}>Agent Type</InputLabel>
                    <Select
                      value={agentType}
                      onChange={handleAgentTypeChange}
                      label="Agent Type"
                      IconComponent={() => <Icon fontSize="large">arrow_drop_down</Icon>}
                      sx={{ textAlign: "center", fontSize: "0.9rem" }}
                      MenuProps={{ PaperProps: { sx: { textAlign: "center" } } }}
                    >
                      <MenuItem value="All">All</MenuItem>
                      {[...new Set([...alerts.critical, ...alerts.high, ...alerts.medium, ...alerts.low].map(alert => alert.container_name))].map(agent => (
                        <MenuItem key={agent} value={agent}>{agent}</MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Grid>
                <Grid item xs={12} sm={2}>
                  <FormControl fullWidth>
                    <InputLabel style={{ fontSize: "0.9rem", textAlign: "center" }}>Alert Type</InputLabel>
                    <Select
                      value={filter}
                      onChange={handleFilterChange}
                      label="Alert Type"
                      IconComponent={() => <Icon fontSize="large">arrow_drop_down</Icon>}
                      sx={{ textAlign: "center", fontSize: "0.9rem" }}
                      MenuProps={{ PaperProps: { sx: { textAlign: "center" } } }}
                    >
                      <MenuItem value="All">All</MenuItem>
                      {['critical', 'high', 'medium', 'low'].map(type => (
                        <MenuItem key={type} value={type}>{type}</MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Grid>
                <Grid item xs={12} sm={2}>
                  <FormControl fullWidth>
                    <InputLabel style={{ fontSize: "0.9rem", textAlign: "center" }}>Time Filter</InputLabel>
                    <Select
                      value={timeFilter}
                      onChange={handleTimeFilterChange}
                      label="Time Filter"
                      IconComponent={() => <Icon fontSize="large">arrow_drop_down</Icon>}
                      sx={{ textAlign: "center", fontSize: "0.9rem" }}
                      MenuProps={{ PaperProps: { sx: { textAlign: "center" } } }}
                    >
                      <MenuItem value="All">All</MenuItem>
                      <MenuItem value="15m">Last 15 minutes</MenuItem>
                      <MenuItem value="30m">Last 30 minutes</MenuItem>
                      <MenuItem value="1h">Last 1 hour</MenuItem>
                      <MenuItem value="3h">Last 3 hours</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
                <Grid item xs={12} sm={2}>
                  <FormControl fullWidth>
                    <InputLabel style={{ fontSize: "0.9rem", textAlign: "center" }}>Sort By</InputLabel>
                    <Select
                      value={sortOrder}
                      onChange={handleSortOrderChange}
                      label="Sort By"
                      IconComponent={() => <Icon fontSize="large">arrow_drop_down</Icon>}
                      sx={{ textAlign: "center", fontSize: "0.9rem" }}
                      MenuProps={{ PaperProps: { sx: { textAlign: "center" } } }}
                    >
                      <MenuItem value="newest">Newest First</MenuItem>
                      <MenuItem value="oldest">Oldest First</MenuItem>
                      <MenuItem value="lowest_priority">Lowest Priority</MenuItem>
                      <MenuItem value="highest_priority">Highest Priority</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
                <Grid item xs={12} sm={4} display="flex" alignItems="center">
                  <TextField
                    label="Search"
                    variant="outlined"
                    size="small"
                    value={searchText}
                    onChange={handleSearchChange}
                    fullWidth
                  />
                  <IconButton sx={{ ml: 1 }}>
                    <Icon>settings</Icon>
                  </IconButton>
                </Grid>
              </Grid>

              {/* Select All and Buttons */}
              <MDBox display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                <MDBox display="flex" alignItems="center">
                  <Checkbox
                    checked={selectAll}
                    onChange={handleSelectAllChange}
                    color="primary"
                  />
                  <MDTypography variant="body2" color="textSecondary" sx={{ ml: 1, minWidth: '75px' }}>
                    {selectedAlerts.length > 0 ? `${selectedAlerts.length} Selected` : ""}
                  </MDTypography>
                </MDBox>
                <MDBox display="flex" sx={{ ml: 'auto' }}>
                  <MDButton
                    variant="contained"
                    color="light"
                    sx={{ ml: 2 }}
                    startIcon={<Icon>notifications_active</Icon>}
                  >
                    Enable Selected
                  </MDButton>
                  <MDButton
                    variant="contained"
                    color="light"
                    sx={{ ml: 2 }}
                    startIcon={<Icon>notifications_off</Icon>}
                  >
                    Disable Selected
                  </MDButton>
                  <MDButton
                    variant="contained"
                    color="light"
                    sx={{ ml: 2 }}
                    startIcon={<Icon>check_box</Icon>}
                  >
                    Resolve
                  </MDButton>
                  <MDButton
                    variant="contained"
                    color="light"
                    sx={{ ml: 2 }}
                    startIcon={<Icon>delete</Icon>}
                  >
                    Delete
                  </MDButton>
                  <MDButton
                    variant="contained"
                    color="light"
                    sx={{ ml: 2 }}
                    startIcon={<Icon>delete_outline</Icon>}
                  >
                    Trash
                  </MDButton>
                </MDBox>
              </MDBox>

              <MDBox p={2} position="relative">
                <MDBox position="absolute" top={0} bottom={0} left="48px" width="2px" bgcolor="grey.700" />
                {filteredAlerts.map((alert) => (
                  <MDBox key={alert.id} display="flex" alignItems="center" mb={2} sx={{ cursor: "pointer", borderBottom: "1px solid #e0e0e0", pb: 2 }}>
                    <Checkbox
                      checked={selectedAlerts.includes(alert.id)}
                      onChange={() => handleSelectAlertChange(alert.id)}
                      sx={{ marginRight: 1 }}
                    />
                    <MDTypography variant="caption" sx={{ width: "100px", textAlign: "right", marginRight: "16px" }}>
                      {dayjs(alert.timestamp).format("YYYY-MM-DD hh:mm:ss A")}
                    </MDTypography>
                    <MDBox position="relative" mr={2} zIndex={1}>
                      <Icon sx={{ color: getAlertColor(alert.severity), fontSize: "1.5rem" }}>{alert.severity === 'low' ? 'info' : 'warning'}</Icon>
                    </MDBox>
                    <MDBox flex={1} onClick={() => handleAlertClick(alert)}>
                      <MDTypography variant="h6" fontWeight="bold">{alert.event_type}</MDTypography>
                      <MDTypography variant="body2" color="text">{alert.message}</MDTypography>
                    </MDBox>
                    <Icon
                      color={favorites.has(alert.id) ? "primary" : "action"}
                      onClick={() => handleFavoriteToggle(alert.id)}
                      sx={{ fontSize: "1.5rem", cursor: "pointer", ml: 1 }}
                    >
                      {favorites.has(alert.id) ? "star" : "star_border"}
                    </Icon>
                  </MDBox>
                ))}
                <MDBox display="flex" justifyContent="center" mt={2}>
                  <Pagination
                    count={totalPages}
                    page={currentPage}
                    onChange={handlePageChange}
                    variant="outlined"
                    shape="rounded"
                    showFirstButton
                    showLastButton
                  />
                </MDBox>
              </MDBox>
            </Card>
          </Grid>
        </Grid>
      </MDBox>
      <Drawer
        anchor="right"
        open={isDrawerOpen}
        onClose={closeDrawer}
        sx={{
          "& .MuiDrawer-paper": {
            width: { xs: "100%", sm: "500px", md: "600px" },
            maxWidth: "100vw",
            height: "100vh",
            padding: "20px",
            overflowY: "auto",
          },
        }}
      >
        <MDBox display="flex" flexDirection="column" justifyContent="space-between" height="100%">
          {selectedAlert && (
            <MDBox flex={1}>
              <MDTypography variant="h6" fontWeight="bold" mb={2}>
                Alert Details
              </MDTypography>
              {Object.entries(selectedAlert).map(([key, value]) => (
                <MDBox key={key} mb={1}>
                  <MDTypography variant="body2" fontWeight="bold">
                    {key}:
                  </MDTypography>
                  <MDTypography variant="body2">
                    {typeof value === 'object' ? JSON.stringify(value, null, 2) : value.toString()}
                  </MDTypography>
                </MDBox>
              ))}
            </MDBox>
          )}
          <MDButton
            onClick={closeDrawer}
            color="primary"
            variant="contained"
            fullWidth
            sx={{ mt: 2 }}
          >
            Close
          </MDButton>
        </MDBox>
      </Drawer>
      <Snackbar
        open={alertOpen}
        autoHideDuration={6000}
        onClose={() => setAlertOpen(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
      >
        <Alert onClose={() => setAlertOpen(false)} severity={alertSeverity} sx={{ width: '100%' }}>
          {alertMessage}
        </Alert>
      </Snackbar>
      <Footer />
    </DashboardLayout>
  );
}

export default EventAlert;