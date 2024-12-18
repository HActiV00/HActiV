'use client'

import React, { useState, useRef, useEffect, useMemo, useCallback } from "react";
import { Box, Grid, Typography, Card, CardContent, Button, TextField, Select, MenuItem, CircularProgress, Checkbox, Collapse, FormControlLabel, Alert, Modal } from "@mui/material";
import { Refresh, GetApp, Print, DeselectOutlined, Close, ExpandMore, ExpandLess } from "@mui/icons-material";
import { Line, Pie } from "react-chartjs-2";
import * as d3 from "d3";
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

import DashboardLayout from "examples/LayoutContainers/DashboardLayout";
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import Footer from "examples/Footer";
import DataTable from "examples/Tables/DataTable";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, ArcElement);

const buttonStyles = {
  height: 48,
  color: "#ffffff",
  borderColor: "#354e6e",
  backgroundColor: "#354e6e",
  textTransform: "none",
  "&:hover": {
    backgroundColor: "#354e6e",
    borderColor: "#354e6e",
  },
};

const printStyles = `
  @media print {
    nav, 
    aside,
    .MuiDrawer-root,
    .no-print,
    button,
    .MuiAppBar-root,
    [role="navigation"],
    #sidebar,
    .navigation-menu {
      display: none !important;
    }

    body, 
    main, 
    .MuiContainer-root,
    .dashboard-content {
      margin: 0 !important;
      padding: 0 !important;
      width: 100% !important;
      max-width: 100% !important;
    }

    .chart-container {
      break-inside: avoid;
      page-break-inside: avoid;
      margin-bottom: 20px;
    }

    .MuiCard-root {
      box-shadow: none !important;
      border: 1px solid #ddd !important;
      margin-bottom: 20px !important;
    }

    .MuiGrid-item {
      page-break-inside: avoid;
      break-inside: avoid;
    }
  }
`;

export default function FileEventDashboard() {
  const [lastUpdated, setLastUpdated] = useState(() => new Date().toLocaleString());
  const [eventFilter, setEventFilter] = useState("all");
  const [timeRange, setTimeRange] = useState({ start: "", end: "" });
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedData, setSelectedData] = useState(null);
  const [fileEventsData, setFileEventsData] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedRows, setSelectedRows] = useState([]);
  const [drawerWidth, setDrawerWidth] = useState(400);
  const [isResizing, setIsResizing] = useState(false);
  const [expandedRows, setExpandedRows] = useState({});
  const [columnSelection, setColumnSelection] = useState({
    id: true,
    event_type: true,
    timestamp: true,
    container_name: true,
    process_name: true,
    filename: true,
    file_size: true,
    mount_status: true,
  });
  const [rawDataExpanded, setRawDataExpanded] = useState({});
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedRow, setSelectedRow] = useState(null);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [openResponse, deleteResponse] = await Promise.all([
        fetch('/api/dashboard?event_type=file_open'),
        fetch('/api/dashboard?event_type=delete')
      ]);

      if (!openResponse.ok || !deleteResponse.ok) {
        throw new Error(`HTTP error! status: ${openResponse.status}, ${deleteResponse.status}`);
      }

      const openData = await openResponse.json();
      const deleteData = await deleteResponse.json();

      console.log('Open Data:', openData);
      console.log('Delete Data:', deleteData);

      const combinedData = [
        ...openData.map(item => ({ ...item, eventCategory: 'open' })),
        ...deleteData.map(item => ({ ...item, eventCategory: 'delete' }))
      ];

      console.log('Combined Data:', combinedData);
      setFileEventsData(combinedData);
      setLastUpdated(new Date().toLocaleString());
    } catch (error) {
      console.error("Failed to fetch data:", error);
      setError("Failed to fetch data. Please try again.");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    if (fileEventsData.length > 0 && selectedRows.length === 0) {
      setSelectedRows([fileEventsData[0].id]);
    }
  }, [fileEventsData, selectedRows.length]);

  useEffect(() => {
    if (selectedData) {
      const calculateWidth = (data) => {
        const stringData = JSON.stringify(data, null, 2);
        const lines = stringData.split('\n');
        const maxLineLength = Math.max(...lines.map(line => line.length));
        return Math.max(400, Math.min(1200, maxLineLength * 8));
      };

      const newWidth = Array.isArray(selectedData)
        ? calculateWidth(selectedData[0])
        : calculateWidth(selectedData);

      setDrawerWidth(newWidth);
    }
  }, [selectedData]);

  const filteredFileEventsData = useMemo(() => {
    return fileEventsData.filter((data) => {
      const matchesType = eventFilter === "all" || data.eventCategory === eventFilter;
      const matchesTimeRange = (!timeRange.start || new Date(data.timestamp) >= new Date(timeRange.start)) &&
                               (!timeRange.end || new Date(data.timestamp) <= new Date(timeRange.end));
      return matchesType && matchesTimeRange;
    });
  }, [fileEventsData, eventFilter, timeRange]);

  const lineChartData = useMemo(() => {
    if (filteredFileEventsData.length === 0) {
      return {
        labels: [],
        datasets: []
      };
    }
    const groupedData = d3.group(
      filteredFileEventsData,
      d => d3.timeHour(new Date(d.timestamp))
    );
    
    const sortedData = Array.from(groupedData)
      .sort((a, b) => a[0] - b[0])
      .map(([time, group]) => ({
        time,
        open: group.filter(d => d.eventCategory === 'open').length,
        delete: group.filter(d => d.eventCategory === 'delete').length
      }));

    return {
      labels: sortedData.map(d => d.time.toLocaleTimeString()),
      datasets: [
        {
          label: "File Open Events",
          data: sortedData.map(d => ({ x: d.time, y: d.open })),
          borderColor: "rgba(75,192,192,1)",
          backgroundColor: "rgba(75,192,192,0.2)",
          tension: 0.4,
        },
        {
          label: "File Delete Events",
          data: sortedData.map(d => ({ x: d.time, y: d.delete })),
          borderColor: "rgba(255,99,132,1)",
          backgroundColor: "rgba(255,99,132,0.2)",
          tension: 0.4,
        }
      ]
    };
  }, [filteredFileEventsData]);

  const pieChartData = useMemo(() => {
    if (filteredFileEventsData.length === 0) {
      return {
        labels: [],
        datasets: [{
          data: [],
          backgroundColor: [],
        }]
      };
    }
    const eventCounts = d3.rollup(
      filteredFileEventsData,
      v => ({ count: v.length, data: v }),
      d => d.eventCategory
    );

    const labels = Array.from(eventCounts.keys());
    const data = Array.from(eventCounts.values());
    
    return {
      labels,
      datasets: [{
        data: data.map(d => d.count),
        backgroundColor: ["#36A2EB", "#FF6384"],
      }]
    };
  }, [filteredFileEventsData]);


  const handleRefresh = () => {
    fetchData();
  };

  const handleExportCSV = () => {
    const headers = Object.keys(columnSelection).filter(key => columnSelection[key]);
    const csvContent = [
      headers.join(","),
      ...filteredFileEventsData.map(row => 
        headers.map(header => row[header]).join(",")
      )
    ].join("\n");

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement("a");
    if (link.download !== undefined) {
      const url = URL.createObjectURL(blob);
      link.setAttribute("href", url);
      link.setAttribute("download", "file_events.csv");
      link.style.visibility = 'hidden';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    }
  };

  const handlePrintPDF = () => {
    window.print();
  };

  const handleDataClick = useCallback((data) => {
    console.log('Selected data:', data); 
    setSelectedData(data);
    setDrawerOpen(true);
  }, []);


  const handleRowClick = useCallback((rowData, event) => {
    const target = event.target;
    const isCheckbox = target.type === 'checkbox' || target.closest('.MuiCheckbox-root');
    
    if (!isCheckbox) {
      setSelectedRow(rowData);
      setModalOpen(true);
    }
  }, []);

  const handleDeselectAll = () => {
    setSelectedRows([]);
  };

  const handleColumnSelectionChange = (event) => {
    setColumnSelection(prev => ({
      ...prev,
      [event.target.name]: event.target.checked,
    }));
  };

  const ExpandedRow = ({ row }) => (
    <Box sx={{ p: 2, backgroundColor: 'background.default' }}>
      <Typography><strong>Process Name:</strong> {row.original.process_name}</Typography>
      <Typography><strong>File Size:</strong> {row.original.file_size}</Typography>
      <Typography><strong>Mount Status:</strong> {row.original.mount_status}</Typography>
    </Box>
  );

  const columns = useMemo(() => [
    {
      Header: "Select",
      accessor: "select",
      Cell: ({ row }) => (
        <Checkbox
          checked={selectedRows.includes(row.original.id)}
          onChange={() => {
            setSelectedRows(prev => 
              prev.includes(row.original.id)
                ? prev.filter(id => id !== row.original.id)
                : [...prev, row.original.id]
            );
          }}
        />
      ),
    },
    ...Object.entries(columnSelection)
      .filter(([_, isSelected]) => isSelected)
      .map(([key, _]) => ({
        Header: key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase()),
        accessor: key,
      })),
    {
      Header: '',
      id: 'expander',
      Cell: ({ row }) => (
        <Button
          onClick={() => setExpandedRows(prev => ({ ...prev, [row.original.id]: !prev[row.original.id] }))}
          startIcon={expandedRows[row.original.id] ? <ExpandLess /> : <ExpandMore />}
          sx={{ fontSize: '0.75rem' }}
        >
          {expandedRows[row.original.id] ? 'Hide Details' : 'Show Details'}
        </Button>
      ),
    },
  ], [columnSelection, selectedRows, expandedRows]);

  const handleMouseDown = (e) => {
    e.preventDefault();
    setIsResizing(true);
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  };

  const handleMouseMove = useCallback((e) => {
    if (!isResizing) return;
    const newWidth = (window.innerWidth - e.clientX) * 1.5;
    setDrawerWidth(Math.max(400, Math.min(1200, newWidth)));
  }, [isResizing]);

  const handleMouseUp = useCallback(() => {
    setIsResizing(false);
    document.removeEventListener('mousemove', handleMouseMove);
    document.removeEventListener('mouseup', handleMouseUp);
  }, [handleMouseMove]);

  useEffect(() => {
    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [handleMouseMove, handleMouseUp]);

  const handleOverlayClick = useCallback((e) => {
    if (e.target === e.currentTarget) {
      setDrawerOpen(false);
    }
  }, []);

  useEffect(() => {
    if (Array.isArray(selectedData)) {
      setRawDataExpanded(selectedData.reduce((acc, item) => ({ ...acc, [item.id]: false }), {}));
    } else {
      setRawDataExpanded({ single: false });
    }
  }, [selectedData]);

  return (
    <DashboardLayout>
      <style jsx global>{`
        ::-webkit-scrollbar {
          width: 10px;
        }
        ::-webkit-scrollbar-track {
          background: #f1f1f1;
          border-radius: 5px;
        }
        ::-webkit-scrollbar-thumb {
          background: #888;
          border-radius: 5px;
        }
        ::-webkit-scrollbar-thumb:hover {
          background: #555;
        }
        @keyframes pulse {
          0% {
            background-color: rgba(0, 0, 255, 0.1);
          }
          50% {
            background-color: rgba(0, 0, 255, 0.3);
          }
          100% {
            background-color: rgba(0, 0, 255, 0.1);
          }
        }
      `}</style>
      <style>{printStyles}</style>
      <DashboardNavbar className="no-print" />
      <Box 
        sx={{ 
          display: 'flex', 
          transition: 'margin-right 0.3s ease-in-out', 
          marginRight: drawerOpen ? `${drawerWidth}px` : 0,
          position: 'relative',
          minHeight: '100vh'
        }}
        onClick={handleOverlayClick}
      >
        <Box sx={{ flexGrow: 1, overflow: 'hidden' }}>
          {isLoading && (
            <Box display="flex" justifyContent="center" alignItems="center" height="100vh">
              <CircularProgress />
            </Box>
          )}
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}
          {!isLoading && !error && (
            <>
              <Box mx={3} mt={3} display="flex" alignItems="center" className="no-print">
                <Typography variant="h4" color="textPrimary" gutterBottom>
                  File Event Dashboard
                </Typography>
                <Button
                  startIcon={<Refresh />}
                  variant="contained"
                  onClick={handleRefresh}
                  sx={{ ml: 2, ...buttonStyles }}
                >
                  Refresh
                </Button>
                <Button
                  startIcon={<GetApp />}
                  variant="contained"
                  onClick={handleExportCSV}
                  sx={{ ml: 2, ...buttonStyles }}
                >
                  Export CSV
                </Button>
                <Button
                  startIcon={<Print />}
                  variant="contained"
                  onClick={handlePrintPDF}
                  sx={{ ml: 2, ...buttonStyles }}
                >
                  Print PDF
                </Button>
                <Typography variant="body2" color="textSecondary" sx={{ ml: 2 }}>
                  Last Updated: {lastUpdated}
                </Typography>
              </Box>

              <Box mx={3} my={2} display="flex" alignItems="center" gap={1} width="100%" sx={{ height: 48 }} className="no-print">
                <TextField
                  label="Start Time"
                  type="datetime-local"
                  InputLabelProps={{ shrink: true }}
                  onChange={(e) => setTimeRange((prev) => ({ ...prev, start: e.target.value }))}
                  sx={{ width: '150px', height: '48px' }}
                />
                <TextField
                  label="End Time"
                  type="datetime-local"
                  InputLabelProps={{ shrink: true }}
                  onChange={(e) => setTimeRange((prev) => ({ ...prev, end: e.target.value }))}
                  sx={{ width: '150px', height: '48px' }}
                />
                <Select
                  value={eventFilter}
                  onChange={(e) => setEventFilter(e.target.value)}
                  displayEmpty
                  sx={{
                    width: '150px',
                    height: '48px',
                    '.MuiSelect-select': {
                      paddingTop: 1.5,
                      paddingBottom: 1.5,
                    },
                  }}
                >
                  <MenuItem value="all">All Events</MenuItem>
                  <MenuItem value="open">Open</MenuItem>
                  <MenuItem value="delete">Delete</MenuItem>
                </Select>
              </Box>

              <Grid container spacing={2} sx={{ mt: 2 }} className="chart-container">
                <Grid item xs={12} md={8}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6">File Events Over Time</Typography>
                      <Box sx={{ height: 400 }}>
                        {filteredFileEventsData.length > 0 ? (
                          <Line
                            data={lineChartData}
                            options={{
                              responsive: true,
                              maintainAspectRatio: false,
                              scales: {
                                y: {
                                  beginAtZero: true,
                                  title: {
                                    display: true,
                                    text: 'Number of Events'
                                  }
                                },
                                x: {
                                  title: {
                                    display: true,
                                    text: 'Time'
                                  }
                                }
                              },
                              onClick: (event, elements) => {
                                if (elements.length > 0) {
                                  const dataIndex = elements[0].index;
                                  const clickedData = lineChartData.datasets[0].data[dataIndex].data;
                                  handleDataClick(clickedData);
                                }
                              }
                            }}
                          />
                        ) : (
                          <Typography variant="h6" sx={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                            No data available
                          </Typography>
                        )}
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>

                <Grid item xs={12} md={4}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6">Event Type Distribution</Typography>
                      <Box sx={{ height: 400 }}>
                        {filteredFileEventsData.length > 0 ? (
                          <Pie
                            data={pieChartData}
                            options={{
                              responsive: true,
                              maintainAspectRatio: false,
                              plugins: {
                                legend: {
                                  position: 'right',
                                }
                              },
                              onClick: (event, elements) => {
                                if (elements.length > 0) {
                                  const dataIndex = elements[0].index;
                                  const clickedData = filteredFileEventsData.filter(d => d.eventCategory === pieChartData.labels[dataIndex]);
                                  handleDataClick(clickedData);
                                }
                              }
                            }}
                          />
                        ) : (
                          <Typography variant="h6" sx={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                            No data available
                          </Typography>
                        )}
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>

                <Grid item xs={12}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6">File Event Timeline</Typography>
                      <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                        <Button
                          size="small"
                          variant="contained"
                          onClick={handleDeselectAll}
                          startIcon={<DeselectOutlined />}
                          sx={{ ...buttonStyles }}
                        >
                          Deselect All
                        </Button>
                        <Typography variant="caption" color="text.secondary">
                          Click on a row to view detailed information
                        </Typography>
                      </Box>
                      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2, mb: 2, maxHeight: '150px', overflowY: 'auto' }}>
                        {Object.entries(columnSelection).map(([key, value]) => (
                          <FormControlLabel
                            key={key}
                            control={
                              <Checkbox
                                checked={value}
                                onChange={handleColumnSelectionChange}
                                name={key}
                                size="small"
                              />
                            }
                            label={key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                          />
                        ))}
                      </Box>
                      <DataTable
                        table={{
                          columns: columns,
                          rows: filteredFileEventsData,
                        }}
                        isSorted={true}
                        entriesPerPage={true}
                        showTotalEntries={true}
                        canSearch={true}
                        pagination={{
                          variant: "contained",
                          shape: "rounded"
                        }}
                        customTableContainerStyles={{
                          "& .MuiTablePagination-root": {
                            display: "flex",
                            flexDirection: "column",
                            alignItems: "center",
                            "& .MuiTablePagination-toolbar": {
                              display: "flex",
                              flexDirection: "column-reverse",
                              alignItems: "center",
                              gap: "1rem",
                            },
                            "& .MuiTablePagination-actions": {
                              marginLeft: 0,
                            },
                            "& .MuiTablePagination-displayedRows": {
                              margin: "0.5rem 0",
                              alignSelf: "flex-end",
                            },
                          },
                        }}
                        customRowStyles={(row) => ({
                          backgroundColor: selectedRows.includes(row.id) ? 'rgba(0, 0, 255, 0.1)' : 'inherit',
                          animation: selectedRows.includes(row.id) ? 'pulse 2s infinite' : 'none',
                          cursor: 'pointer',
                        })}
                        onRowClick={handleRowClick}
                        renderRowSubComponent={({ row }) => (
                          <Collapse in={expandedRows[row.original.id]}>
                            <ExpandedRow row={row} />
                          </Collapse>
                        )}
                      />
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>
            </>
          )}
          <Footer className="no-print" />
        </Box>
        <Box
          sx={{
            position: 'fixed',
            top: 0,
            right: 0,
            bottom: 0,
            width: drawerWidth,
            backgroundColor: 'background.paper',
            boxShadow: '-4px 0 8px rgba(0, 0, 0, 0.1)',
            transform: drawerOpen ? 'translateX(0)' : `translateX(${drawerWidth}px)`,
            transition: 'transform 0.3s ease-in-out',
            zIndex: 1200,
            overflowY: 'auto',
          }}
        >
          <Box
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              bottom: 0,
              width: '4px',
              cursor: 'ew-resize',
              '&:hover': {
                backgroundColor: 'rgba(0, 0, 0, 0.1)',
              },
            }}
            onMouseDown={handleMouseDown}
          />
          <Box sx={{ p: 2 }}>
            <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
              <Typography variant="h6">Detailed Information</Typography>
              <Button
                onClick={() => setDrawerOpen(false)}
                startIcon={<Close />}
                sx={{ ...buttonStyles }}
              >
                Close
              </Button>
            </Box>
            {selectedData && (
              <Box>
                {Array.isArray(selectedData) ? (
                  selectedData.map((item, index) => (
                    <Card key={index} sx={{ mb: 2 }}>
                      <CardContent>
                        <Typography variant="subtitle1" gutterBottom>
                          Event {index + 1}
                        </Typography>
                        {Object.entries(item).map(([key, value]) => (
                          <Typography key={key} variant="body2" gutterBottom>
                            <strong>{key}:</strong> {value}
                          </Typography>
                        ))}
                        <Button
                          onClick={() => setRawDataExpanded(prev => ({ ...prev, [item.id]: !prev[item.id] }))}
                          startIcon={rawDataExpanded[item.id] ? <ExpandLess /> : <ExpandMore />}
                          sx={{ mt: 1 }}
                        >
                          {rawDataExpanded[item.id] ? 'Hide' : 'Show'} Raw Data
                        </Button>
                        <Collapse in={rawDataExpanded[item.id]}>
                          <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                            {JSON.stringify(item, null, 2)}
                          </pre>
                        </Collapse>
                      </CardContent>
                    </Card>
                  ))
                ) : (
                  <Card>
                    <CardContent>
                      {Object.entries(selectedData).map(([key, value]) => (
                        <Typography key={key} variant="body2" gutterBottom>
                          <strong>{key}:</strong> {value}
                        </Typography>
                      ))}
                      <Button
                        onClick={() => setRawDataExpanded(prev => ({ ...prev, single: !prev.single }))}
                        startIcon={rawDataExpanded.single ? <ExpandLess /> : <ExpandMore />}
                        sx={{ mt: 1 }}
                      >
                        {rawDataExpanded.single ? 'Hide' : 'Show'} Raw Data
                      </Button>
                      <Collapse in={rawDataExpanded.single}>
                        <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                          {JSON.stringify(selectedData, null, 2)}
                        </pre>
                      </Collapse>
                    </CardContent>
                  </Card>
                )}
              </Box>
            )}
          </Box>
        </Box>
      </Box>
      <Modal open={modalOpen} onClose={() => setModalOpen(false)}>
        <Box sx={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          width: '80%',
          maxWidth: 600,
          bgcolor: 'background.paper',
          boxShadow: 24,
          p: 4,
          maxHeight: '80vh',
          overflowY: 'auto',
        }}>
          <Typography variant="h6" component="h2" gutterBottom>
            Event Details
          </Typography>
          {selectedRow && Object.entries(selectedRow).map(([key, value]) => (
            <Typography key={key} sx={{ mt: 2 }}>
              <strong>{key}:</strong> {value}
            </Typography>
          ))}
          <Button onClick={() => setModalOpen(false)} sx={{ mt: 2 }}>
            Close
          </Button>
        </Box>
      </Modal>
    </DashboardLayout>
  );
}