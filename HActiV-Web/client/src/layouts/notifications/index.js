import { useState } from "react";
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

function Notifications() {
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

  const alerts = [
    { id: 1, type: "info", agent: "Agent A", label: "Info Alert", message: "An informational alert example.", icon: "info", time: dayjs().subtract(10, 'minute').format("YYYY-MM-DD hh:mm:ss A") },
    { id: 2, type: "medium", agent: "Agent B", label: "Medium Alert", message: "A medium alert example.", icon: "priority_high", time: dayjs().subtract(20, 'minute').format("YYYY-MM-DD hh:mm:ss A") },
    { id: 3, type: "high", agent: "Agent A", label: "High Alert", message: "A high alert example.", icon: "warning", time: dayjs().subtract(45, 'minute').format("YYYY-MM-DD hh:mm:ss A") },
    { id: 4, type: "critical", agent: "Agent C", label: "Critical Alert", message: "A critical alert example.", icon: "error", time: dayjs().subtract(70, 'minute').format("YYYY-MM-DD hh:mm:ss A") },
  ];

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

  const getFilteredAlerts = () => {
    let filtered = alerts;

    if (agentType !== "All") filtered = filtered.filter((alert) => alert.agent === agentType);
    if (filter !== "All") filtered = filtered.filter((alert) => alert.type === filter.toLowerCase());
    if (timeFilter !== "All") {
      const now = dayjs();
      filtered = filtered.filter((alert) => {
        const alertTime = dayjs(alert.time);
        if (timeFilter === "15m") return now.diff(alertTime, "minute") <= 15;
        if (timeFilter === "30m") return now.diff(alertTime, "minute") <= 30;
        if (timeFilter === "1h") return now.diff(alertTime, "hour") <= 1;
        if (timeFilter === "3h") return now.diff(alertTime, "hour") <= 3;
        return true;
      });
    }
    filtered = sortOrder === "newest" ? filtered.sort((a, b) => dayjs(b.time).diff(dayjs(a.time))) : filtered.sort((a, b) => dayjs(a.time).diff(dayjs(b.time)));
    if (searchText) filtered = filtered.filter((alert) => alert.label.toLowerCase().includes(searchText.toLowerCase()) || alert.message.toLowerCase().includes(searchText.toLowerCase()));

    return filtered.slice(indexOfFirstAlert, indexOfLastAlert);
  };

  const getAlertColor = (type) => {
    switch (type) {
      case "info":
        return "success"; // 초록색
      case "medium":
        return "warning"; // 노란색
      case "high":
        return "orange"; // 주황색
      case "critical":
        return "error"; // 빨간색
      default:
        return "inherit";
    }
  };

  const filteredAlerts = getFilteredAlerts();
  const totalPages = Math.ceil(alerts.length / alertsPerPage);

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

              {/* All Alerts Counter */}
              <MDBox display="flex" alignItems="center" mb={2}>
                <MDTypography variant="h3" fontWeight="bold" color="textSecondary">10</MDTypography>
                <MDTypography variant="subtitle1" color="textSecondary" sx={{ ml: 1 }}>All Alerts</MDTypography>
                <MDTypography variant="caption" color="textSecondary" sx={{ ml: 'auto' }}>
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
                      <MenuItem value="Agent A">Agent A</MenuItem>
                      <MenuItem value="Agent B">Agent B</MenuItem>
                      <MenuItem value="Agent C">Agent C</MenuItem>
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
                      <MenuItem value="Info">Info</MenuItem>
                      <MenuItem value="Medium">Medium</MenuItem>
                      <MenuItem value="High">High</MenuItem>
                      <MenuItem value="Critical">Critical</MenuItem>
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
                      {dayjs(alert.time).format("YYYY-MM-DD hh:mm:ss A")}
                    </MDTypography>
                    <MDBox position="relative" mr={2} zIndex={1}>
                      <Icon color={getAlertColor(alert.type)} sx={{ fontSize: "1.5rem", position: "relative" }}>{alert.icon}</Icon>
                    </MDBox>
                    <MDBox flex={1} onClick={() => handleAlertClick(alert)}>
                      <MDTypography variant="h6" fontWeight="bold">{alert.label}</MDTypography>
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
              <MDTypography variant="h6" fontWeight="bold">
                {selectedAlert.label}
              </MDTypography>
              <MDTypography variant="body2" mt={2}>
                {selectedAlert.message}
              </MDTypography>
              <MDTypography variant="caption" color="textSecondary" mt={1}>
                {selectedAlert.time}
              </MDTypography>
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
      <Footer />
    </DashboardLayout>
  );
}

export default Notifications;
