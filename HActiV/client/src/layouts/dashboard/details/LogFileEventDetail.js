'use client'

import React, { useState, useRef, useEffect, useMemo, useCallback } from "react";
import { Box, Grid, Typography, Card, CardContent, Button, TextField, Select, MenuItem, CircularProgress, Checkbox, Collapse, FormControlLabel, Alert, Modal } from "@mui/material";
import { Refresh, GetApp, Print, DeselectOutlined, Close, ExpandMore, ExpandLess } from "@mui/icons-material";
import { Line, Pie } from "react-chartjs-2";
import { ForceGraph2D } from "react-force-graph";
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

export default function LogFileEventDetail() {
  const [lastUpdated, setLastUpdated] = useState(() => new Date().toLocaleString());
  const fgRef = useRef();
  const graphContainerRef = useRef();
  const [eventFilter, setEventFilter] = useState("all");
  const [timeRange, setTimeRange] = useState({ start: "", end: "" });
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedData, setSelectedData] = useState(null);
  const [logEventsData, setLogEventsData] = useState([]);
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
      const [accessResponse, deleteResponse] = await Promise.all([
        fetch('/api/dashboard?event_type=log_file_open'),
        fetch('/api/dashboard?event_type=log_file_delete')
      ]);

      if (!accessResponse.ok || !deleteResponse.ok) {
        throw new Error(`HTTP error! status: ${accessResponse.status}, ${deleteResponse.status}`);
      }

      const accessData = await accessResponse.json();
      const deleteData = await deleteResponse.json();

      console.log('Access Data:', accessData);
      console.log('Delete Data:', deleteData);

      const combinedData = [
        ...accessData.map(item => ({ ...item, eventCategory: 'access' })),
        ...deleteData.map(item => ({ ...item, eventCategory: 'delete' }))
      ];

      console.log('Combined Data:', combinedData);
      setLogEventsData(combinedData);
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
    if (logEventsData.length > 0 && selectedRows.length === 0) {
      setSelectedRows([logEventsData[0].id]);
    }
  }, [logEventsData, selectedRows.length]);

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

  const filteredLogEventsData = useMemo(() => {
    return logEventsData.filter((data) => {
      const matchesType = eventFilter === "all" || data.eventCategory === eventFilter;
      const matchesTimeRange = (!timeRange.start || new Date(data.timestamp) >= new Date(timeRange.start)) &&
                               (!timeRange.end || new Date(data.timestamp) <= new Date(timeRange.end));
      return matchesType && matchesTimeRange;
    });
  }, [logEventsData, eventFilter, timeRange]);

  const lineChartData = useMemo(() => {
    if (filteredLogEventsData.length === 0) {
      return {
        labels: [],
        datasets: []
      };
    }
    const groupedData = d3.group(
      filteredLogEventsData,
      d => d3.timeHour(new Date(d.timestamp))
    );
    
    const sortedData = Array.from(groupedData)
      .sort((a, b) => a[0] - b[0])
      .map(([time, group]) => ({
        time,
        access: group.filter(d => d.eventCategory === 'access').length,
        delete: group.filter(d => d.eventCategory === 'delete').length
      }));

    return {
      labels: sortedData.map(d => d.time.toLocaleTimeString()),
      datasets: [
        {
          label: "Log File Access Events",
          data: sortedData.map(d => ({ x: d.time, y: d.access })),
          borderColor: "rgba(75,192,192,1)",
          backgroundColor: "rgba(75,192,192,0.2)",
          tension: 0.4,
        },
        {
          label: "Log File Delete Events",
          data: sortedData.map(d => ({ x: d.time, y: d.delete })),
          borderColor: "rgba(255,99,132,1)",
          backgroundColor: "rgba(255,99,132,0.2)",
          tension: 0.4,
        }
      ]
    };
  }, [filteredLogEventsData]);

  const pieChartData = useMemo(() => {
    if (filteredLogEventsData.length === 0) {
      return {
        labels: [],
        datasets: [{
          data: [],
          backgroundColor: [],
        }]
      };
    }
    const eventCounts = d3.rollup(
      filteredLogEventsData,
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
  }, [filteredLogEventsData]);

  const networkFlowData = useMemo(() => {
    if (!filteredLogEventsData.length || selectedRows.length === 0) {
      return { nodes: [], links: [] };
    }

    const nodes = new Map();
    const links = new Map();

    filteredLogEventsData.forEach(event => {
      if (!event || !selectedRows.includes(event.id)) return;

      if (!nodes.has(event.process_name)) {
        nodes.set(event.process_name, {
          id: event.process_name,
          label: event.process_name,
          group: event.eventCategory === 'access' ? 1 : 2,
          eventId: event.id
        });
      }

      if (!nodes.has(event.filename)) {
        nodes.set(event.filename, {
          id: event.filename,
          label: event.filename,
          group: 3,
          eventId: event.id
        });
      }

      const linkId = `${event.process_name}-${event.filename}`;
      if (!links.has(linkId)) {
        links.set(linkId, {
          source: event.process_name,
          target: event.filename,
          value: 1
        });
      } else {
        links.get(linkId).value++;
      }
    });

    return {
      nodes: Array.from(nodes.values()),
      links: Array.from(links.values())
    };
  }, [filteredLogEventsData, selectedRows]);

  const handleRefresh = () => {
    fetchData();
  };

  const handleExportCSV = () => {
    const headers = Object.keys(columnSelection).filter(key => columnSelection[key]);
    const csvContent = [
      headers.join(","),
      ...filteredLogEventsData.map(row => 
        headers.map(header => row[header]).join(",")
      )
    ].join("\n");

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement("a");
    if (link.download !== undefined) {
      const url = URL.createObjectURL(blob);
      link.setAttribute("href", url);
      link.setAttribute("download", "log_file_events.csv");
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

  const handleResetGraph = () => {
    if (fgRef.current) {
      networkFlowData.nodes.forEach(node => {
        node.fx = null;
        node.fy = null;
      });
      fgRef.current.zoomToFit(400);
    }
  };

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
                  Log File Event Detail
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
                  <MenuItem value="access">Access</MenuItem>
                  <MenuItem value="delete">Delete</MenuItem>
                </Select>
                <Button
                  size="small"
                  variant="contained"
                  onClick={handleResetGraph}
                  sx={{ ...buttonStyles, width: '150px' }}
                >
                  Reset Graph
                </Button>
              </Box>

              <Grid container spacing={2} sx={{ mt: 2 }} className="chart-container">
                <Grid item xs={12} md={8}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6">Log File Events Over Time</Typography>
                      <Box sx={{ height: 400 }}>
                        {filteredLogEventsData.length > 0 ? (
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
                        {filteredLogEventsData.length > 0 ? (
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
                                  const clickedData = filteredLogEventsData.filter(d => d.eventCategory === pieChartData.labels[dataIndex]);
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
                  <Card ref={graphContainerRef} sx={{ height: 400, overflow: "hidden" }}>
                    <CardContent>
                      <Typography variant="h6">Log File Event Flow Visualization</Typography>
                      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
                        Drag nodes to reposition. Click a fixed node to release it. Double-click for details.
                      </Typography>
                      <Box sx={{ height: 400 }}>
                        {networkFlowData.nodes.length > 0 ? (
                          <ForceGraph2D
                            ref={fgRef}
                            graphData={networkFlowData}
                            nodeLabel={(node) => node.label}
                            linkDirectionalArrowLength={3}
                            linkDirectionalArrowRelPos={1}
                            linkWidth={link => Math.sqrt(link.value)}
                            nodeVal={node => 5}
                            nodeColor={node => node.group === 1 ? "#36A2EB" : (node.group === 2 ? "#FF6384" : "#4BC0C0")}
                            width={graphContainerRef.current ? graphContainerRef.current.offsetWidth : 600}
                            height={graphContainerRef.current ? graphContainerRef.current.offsetHeight : 400}
                            nodeCanvasObject={(node, ctx, globalScale) => {
                              const label = node.label;
                              const fontSize = 14 / globalScale;
                              ctx.font = `${fontSize}px Arial, sans-serif`;
                              ctx.textAlign = "center";
                              ctx.textBaseline = "middle";
                              const textWidth = ctx.measureText(label).width;
                              const bckgDimensions = [textWidth, fontSize].map(n => n + fontSize * 0.8);

                              ctx.fillStyle = node.group === 1 ? "rgba(54, 162, 235, 0.2)" : (node.group === 2 ? "rgba(255, 99, 132, 0.2)" : "rgba(75, 192, 192, 0.2)");
                              ctx.fillRect(
                                node.x - bckgDimensions[0] / 2,
                                node.y - bckgDimensions[1] / 2,
                                ...bckgDimensions
                              );

                              ctx.strokeStyle = node.group === 1 ? "rgba(54, 162, 235, 1)" : (node.group === 2 ? "rgba(255, 99, 132, 1)" : "rgba(75, 192, 192, 1)");
                              ctx.lineWidth = 2 / globalScale;
                              ctx.strokeRect(
                                node.x - bckgDimensions[0] / 2,
                                node.y - bckgDimensions[1] / 2,
                                ...bckgDimensions
                              );

                              ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
                              ctx.fillText(label, node.x, node.y);

                              node.__bckgDimensions = bckgDimensions;
                            }}
                            nodePointerAreaPaint={(node, color, ctx) => {
                              ctx.fillStyle = color;
                              const bckgDimensions = node.__bckgDimensions;
                              bckgDimensions && ctx.fillRect(
                                node.x - bckgDimensions[0] / 2,
                                node.y - bckgDimensions[1] / 2,
                                ...bckgDimensions
                              );
                            }}
                            onNodeDragStart={node => {
                              node.fx = node.x;
                              node.fy = node.y;
                            }}
                            onNodeDrag={(node, translate) => {
                              node.fx = node.x + translate.x;
                              node.fy = node.y + translate.y;
                            }}
                            onNodeDragEnd={node => {
                              node.fx = node.x;
                              node.fy = node.y;
                            }}
                            onNodeClick={(node) => {
                              if (node.fx !== undefined && node.fy !== undefined) {
                                node.fx = undefined;
                                node.fy = undefined;
                              } else {
                                const nodeData = filteredLogEventsData.find(data => data.id === node.eventId);
                                handleDataClick(nodeData);
                              }
                            }}
                            cooldownTicks={200}
                            cooldownTime={5000}
                            d3AlphaDecay={0.01}
                            d3VelocityDecay={0.6}
                            warmupTicks={100}
                            linkStrength={0.3}
                            nodeRelSize={6}
                            onEngineStop={() => fgRef.current.zoomToFit(400)}
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
                      <Typography variant="h6">Log File Event Timeline</Typography>
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
                          rows: filteredLogEventsData,
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
                          cursor: 'pointer',
                          '&:hover': {
                            backgroundColor: 'rgba(0, 0, 255, 0.05)',
                          },
                          '&.selected-row': {
                            animation: 'pulse 0.3s',
                          },
                        })}
                        onRowClick={(rowData, event) => {
                          handleRowClick(rowData, event);
                          const rowElement = document.querySelector(`[data-row-id="${rowData.id}"]`);
                          if (rowElement) {
                            rowElement.classList.add('selected-row');
                            setTimeout(() => rowElement.classList.remove('selected-row'), 300);
                          }
                        }}
                        renderRowSubComponent={({ row }) => <ExpandedRow row={row} />}
                        expandedRows={expandedRows}
                      />
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>
            </>
          )}
          <Footer className="no-print" />
        </Box>
        {drawerOpen && (
          <Box
            data-testid="sliding-panel"
            onClick={(e) => e.stopPropagation()}
            sx={{
              width: `${drawerWidth}px`,
              flexShrink: 0,
              height: 'calc(100vh - 69px)', // Reduced height by 5px
              position: 'fixed',
              top: 64,
              right: 0,
              backgroundColor: 'background.paper',
              boxShadow: 3,
              overflowY: 'auto',
              overflowX: 'hidden',
              transition: 'width 0.3s ease-in-out',
              borderTopLeftRadius: 16,
              borderBottomLeftRadius: 16,
              display: 'flex',
              flexDirection: 'column',
              zIndex: 1000, // Ensure it's above other content
            }}
          >
            <Box
              sx={{
                position: 'absolute',
                top: 0,
                left: -10,
                bottom: 0,
                width: '20px',
                cursor: 'ew-resize',
                backgroundColor: 'transparent',
                '&:hover': {
                  backgroundColor: 'rgba(0, 0, 0, 0.1)',
                },
                zIndex: 1,
              }}
              onMouseDown={handleMouseDown}
            />
            <Box sx={{ 
              p: 2, 
              position: 'sticky', 
              top: 0, 
              backgroundColor: 'background.paper', 
              zIndex: 2,
              borderBottom: '1px solid',
              borderColor: 'divider',
            }}>
              <Box display="flex" justifyContent="space-between" alignItems="center">
                <Typography variant="h6">Detailed Information</Typography>
                <Button
                  onClick={() => setDrawerOpen(false)}
                  sx={{ minWidth: 'auto', p: 0.5 }}
                >
                  <Close />
                </Button>
              </Box>
            </Box>
            <Box sx={{ flexGrow: 1, overflowY: 'auto', p: 2 }}>
              {selectedData && (
                <Box>
                  <Card sx={{ mb: 2, backgroundColor: 'grey.100' }}>
                    <CardContent>
                      <Typography variant="h6" gutterBottom>Event Detail</Typography>
                      {Array.isArray(selectedData) ? (
                        <>
                          <Typography><strong>Number of Events:</strong> {selectedData.length}</Typography>
                          <Typography><strong>Time Range:</strong> {new Date(selectedData[0].timestamp).toLocaleString()} - {new Date(selectedData[selectedData.length - 1].timestamp).toLocaleString()}</Typography>
                          <Typography variant="h6" gutterBottom sx={{ mt: 2 }}>Event List</Typography>
                          <Box sx={{ maxHeight: 'calc(100vh - 300px)', overflowY: 'auto', overflowX: 'hidden' }}>
                            {selectedData.map((event, index) => (
                              <Box key={event.id} sx={{ mb: 2, p: 1, backgroundColor: index % 2 === 0 ? 'background.default' : 'action.hover' }}>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>ID:</strong> {event.id}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Event Type:</strong> {event.event_type}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Timestamp:</strong> {new Date(event.timestamp).toLocaleString()}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Container Name:</strong> {event.container_name}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Process Name:</strong> {event.process_name}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Filename:</strong> {event.filename}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>File Size:</strong> {event.file_size}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Mount Status:</strong> {event.mount_status}</Typography>
                                <Box
                                  sx={{
                                    display: 'flex',
                                    justifyContent: 'space-between',
                                    alignItems: 'center',
                                    cursor: 'pointer',
                                    mt: 2,
                                  }}
                                  onClick={() => setRawDataExpanded(prev => ({ ...prev, [event.id]: !prev[event.id] }))}
                                >
                                  <Typography variant="subtitle2">Raw Data</Typography>
                                  {rawDataExpanded[event.id] ? <ExpandLess /> : <ExpandMore />}
                                </Box>
                                <Collapse in={rawDataExpanded[event.id]}>
                                  <Box sx={{ overflowX: 'auto', mt: 1 }}>
                                    <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                                      {JSON.stringify(event, null, 2)}
                                    </pre>
                                  </Box>
                                </Collapse>
                              </Box>
                            ))}
                          </Box>
                        </>
                      ) : (
                        <>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>ID:</strong> {selectedData.id}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Event Type:</strong> {selectedData.event_type}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Timestamp:</strong> {new Date(selectedData.timestamp).toLocaleString()}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Container Name:</strong> {selectedData.container_name}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Process Name:</strong> {selectedData.process_name}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Filename:</strong> {selectedData.filename}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>File Size:</strong> {selectedData.file_size}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Mount Status:</strong> {selectedData.mount_status}</Typography>
                          <Box
                            sx={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              alignItems: 'center',
                              cursor: 'pointer',
                              mt: 2,
                            }}
                            onClick={() => setRawDataExpanded(prev => ({ ...prev, single: !prev.single }))}
                          >
                            <Typography variant="subtitle2">Raw Data</Typography>
                            {rawDataExpanded.single ? <ExpandLess /> : <ExpandMore />}
                          </Box>
                          <Collapse in={rawDataExpanded.single}>
                            <Box sx={{ overflowX: 'auto', mt: 1 }}>
                              <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                                {JSON.stringify(selectedData, null, 2)}
                              </pre>
                            </Box>
                          </Collapse>
                        </>
                      )}
                    </CardContent>
                  </Card>
                </Box>
              )}
            </Box>
          </Box>
        )}
      </Box>
      <Modal open={modalOpen} onClose={() => setModalOpen(false)}>
        <Box sx={{ position: "absolute", top: "50%", left: "50%", transform: "translate(-50%, -50%)", width: 400, bgcolor: "background.paper", p: 4 }}>
          {selectedRow && (
            <>
              <Typography variant="h6" gutterBottom>Log File Event Details</Typography>
              <Typography>ID: {selectedRow.id}</Typography>
              <Typography>Event Type: {selectedRow.event_type}</Typography>
              <Typography>Timestamp: {new Date(selectedRow.timestamp).toLocaleString()}</Typography>
              <Typography>Container Name: {selectedRow.container_name}</Typography>
              <Typography>Process Name: {selectedRow.process_name}</Typography>
              <Typography>Filename: {selectedRow.filename}</Typography>
              <Typography>File Size: {selectedRow.file_size}</Typography>
              <Typography>Mount Status: {selectedRow.mount_status}</Typography>
            </>
          )}
        </Box>
      </Modal>
    </DashboardLayout>
  );
}