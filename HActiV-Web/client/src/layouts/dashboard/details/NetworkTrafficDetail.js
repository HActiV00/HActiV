'use client'

import React, { useState, useRef, useEffect, useMemo, useCallback } from "react";
import { Box, Grid, Typography, Card, CardContent, Button, TextField, Select, MenuItem, CircularProgress, Checkbox, Collapse, FormGroup, FormControlLabel } from "@mui/material";
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

export default function NetworkTrafficDetail() {
  const [lastUpdated, setLastUpdated] = useState(() => new Date().toLocaleString());
  const fgRef = useRef(null);
  const graphContainerRef = useRef(null);
  const [protocolFilter, setProtocolFilter] = useState("all");
  const [packetSizeRange, setPacketSizeRange] = useState([0, 1500]);
  const [timeRange, setTimeRange] = useState({ start: "", end: "" });
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedData, setSelectedData] = useState(null);
  const [trafficData, setTrafficData] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedRows, setSelectedRows] = useState([]);
  const [drawerWidth, setDrawerWidth] = useState(400);
  const [isResizing, setIsResizing] = useState(false);
  const [expandedRows, setExpandedRows] = useState({});
  const [columnSelection, setColumnSelection] = useState({
    id: true,
    timestamp: true,
    container_name: true,
    src_ip: true,
    dst_ip: true,
    protocol: true,
    packets: true,
    path: true,
    event_type: true,
    size: true,
    total_packets: true,
    total_size: true,
    src_ip_label: true,
    dst_ip_label: true,
  });
  const [rawDataExpanded, setRawDataExpanded] = useState({});
  const [timeInterval, setTimeInterval] = useState(1);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/dashboard?event_type=Network_traffic');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const rawData = await response.json();
      console.log("Raw API response:", rawData);

      setTrafficData(rawData);
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
    if (trafficData.length > 0 && selectedRows.length === 0) {
      setSelectedRows([trafficData[0].id]);
    }
  }, [trafficData, selectedRows.length]);

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

  const filteredTrafficData = useMemo(() => {
    return trafficData.filter((data) => {
      if (!data) return false;
      
      const matchesProtocol = protocolFilter === "all" || data.protocol === protocolFilter;
      const matchesPacketSize = data.packets >= packetSizeRange[0] && data.packets <= packetSizeRange[1];
      const matchesTimeRange = (!timeRange.start || new Date(data.timestamp) >= new Date(timeRange.start)) &&
                               (!timeRange.end || new Date(data.timestamp) <= new Date(timeRange.end));
      return matchesProtocol && matchesPacketSize && matchesTimeRange;
    });
  }, [trafficData, protocolFilter, packetSizeRange, timeRange]);

  const formattedTrafficData = useMemo(() => {
    return filteredTrafficData.map(data => ({
      ...data,
      timestamp: new Date(data.timestamp).toLocaleString(),
      path: JSON.stringify(data.path)
    }));
  }, [filteredTrafficData]);

  const lineChartData = useMemo(() => {
    if (!filteredTrafficData.length) {
      return {
        labels: [],
        datasets: [{
          label: "Packets",
          data: [],
          borderColor: "hsl(var(--chart-1))",
          backgroundColor: "hsl(var(--chart-1) / 0.1)",
          tension: 0.4,
        }]
      };
    }

    const intervalMinutes = {
      1: d3.timeMinute,
      5: d3.timeMinute.every(5),
      10: d3.timeMinute.every(10),
      30: d3.timeMinute.every(30),
      60: d3.timeHour,
    }[timeInterval];

    const groupedData = d3.group(
      filteredTrafficData,
      d => intervalMinutes(new Date(d.timestamp))
    );
    
    const sortedData = Array.from(groupedData)
      .sort((a, b) => a[0] - b[0])
      .map(([time, group]) => ({
        time,
        packets: d3.sum(group, d => Number(d.packets) || 0),
        data: group
      }));

    return {
      labels: sortedData.map(d => d.time.toLocaleString()),
      datasets: [{
        label: `Packets per ${timeInterval} ${timeInterval === 1 ? 'minute' : 'minutes'}`,
        data: sortedData.map(d => ({ x: d.time, y: d.packets, data: d.data })),
        borderColor: "hsl(var(--chart-1))",
        backgroundColor: "hsl(var(--chart-1) / 0.1)",
        tension: 0.4,
      }]
    };
  }, [filteredTrafficData, timeInterval]);

  const pieChartData = useMemo(() => {
    if (!filteredTrafficData.length) {
      return {
        labels: [],
        datasets: [{
          data: [],
          backgroundColor: [],
        }]
      };
    }

    const protocolCounts = d3.rollup(
      filteredTrafficData,
      v => ({ packets: d3.sum(v, d => Number(d.packets) || 0), data: v }),
      d => d.protocol
    );

    const labels = Array.from(protocolCounts.keys());
    const data = Array.from(protocolCounts.values());
    
    return {
      labels,
      datasets: [{
        data: data.map(d => d.packets),
        backgroundColor: ["#36A2EB", "#FF6384", "#FFCE56", "#4BC0C0", "#9966FF"].slice(0, labels.length),
      }]
    };
  }, [filteredTrafficData]);

  const networkFlowData = useMemo(() => {
    if (!filteredTrafficData.length || selectedRows.length === 0) {
      return { nodes: [], links: [] };
    }

    const nodes = new Map();
    const links = [];

    filteredTrafficData.forEach(traffic => {
      if (!traffic || !selectedRows.includes(traffic.id)) return;

      traffic.path.nodes.forEach(node => {
        if (!nodes.has(node.id)) {
          nodes.set(node.id, {
            id: node.id,
            label: `${node.id} (${node.type})`,
            group: node.type === "Docker internal" ? 1 : 2,
            packets: 0,
            eventId: traffic.id
          });
        }
      });

      traffic.path.links.forEach(link => {
        links.push({
          source: link.source,
          target: link.target,
          value: traffic.packets
        });
        nodes.get(link.source).packets += traffic.packets;
        nodes.get(link.target).packets += traffic.packets;
      });
    });

    return {
      nodes: Array.from(nodes.values()),
      links
    };
  }, [filteredTrafficData, selectedRows]);

  const handleRefresh = () => {
    fetchData();
  };

  const handleExportCSV = () => {
    const headers = Object.keys(columnSelection).filter(key => columnSelection[key]);
    const csvContent = [
      headers.join(","),
      ...filteredTrafficData.map(row => 
        headers.map(header => {
          if (header === 'path') {
            return JSON.stringify(row[header]).replace(/,/g, ';');
          }
          return row[header];
        }).join(",")
      )
    ].join("\n");

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement("a");
    if (link.download !== undefined) {
      const url = URL.createObjectURL(blob);
      link.setAttribute("href", url);
      link.setAttribute("download", "network_traffic_data.csv");
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
    const isViewPathButton = target.closest('button') && target.closest('button').textContent === 'View Path';
    
    if (!isCheckbox && !isViewPathButton) {
      handleDataClick(rowData);
    }
    
    if (isViewPathButton) {
      setExpandedRows(prev => ({ ...prev, [rowData.id]: !prev[rowData.id] }));
    }
  }, [handleDataClick]);

  const handleDeselectAll = () => {
    setSelectedRows([]);
  };

  const handleColumnSelectionChange = (event) => {
    setColumnSelection({
      ...columnSelection,
      [event.target.name]: event.target.checked,
    });
  };

  const columns = useMemo(() => [
    {
      Header: "Select",
      accessor: "select",
      Cell: ({ row }) => (
        <Checkbox
          checked={selectedRows.includes(row.original.id)}
          onChange={() => {
            const newSelectedRows = selectedRows.includes(row.original.id)
              ? selectedRows.filter(id => id !== row.original.id)
              : [...selectedRows, row.original.id];
            setSelectedRows(newSelectedRows);
          }}
        />
      ),
    },
    ...(columnSelection.id ? [{ Header: "ID", accessor: "id" }] : []),
    ...(columnSelection.timestamp ? [{ Header: "Timestamp", accessor: "timestamp" }] : []),
    ...(columnSelection.container_name ? [{ Header: "Container Name", accessor: "container_name" }] : []),
    ...(columnSelection.src_ip ? [{ 
      Header: "Source", 
      accessor: row => `${row.src_ip} (${row.src_ip_label})`,
      id: "source"
    }] : []),
    ...(columnSelection.dst_ip ? [{ 
      Header: "Destination", 
      accessor: row => `${row.dst_ip} (${row.dst_ip_label})`,
      id: "destination"
    }] : []),
    ...(columnSelection.protocol ? [{ Header: "Protocol", accessor: "protocol" }] : []),
    ...(columnSelection.packets ? [{ Header: "Packets", accessor: "packets" }] : []),
    ...(columnSelection.path ? [{
      Header: "Path",
      accessor: "path",
      Cell: ({ row }) => {
        const formatNetworkFlow = (path) => {
          if (!path || !path.nodes || path.nodes.length < 2) return "Invalid path data";
          return path.nodes.map((node, index, array) => {
            if (index === array.length - 1) return `[${node.id}(${node.label || 'Unknown'})]`;
            return `[${node.id}(${node.label || 'Unknown'})] -> `;
          }).join('');
        };

        return (
          <>
            <Button
              onClick={() => handleRowClick(row.original, { target: { closest: () => true } })}
              startIcon={expandedRows[row.original.id] ? <ExpandLess /> : <ExpandMore />}
              sx={{ fontSize: '0.75rem' }}
            >
              View Path
            </Button>
            <Collapse in={expandedRows[row.original.id]}>
              <Box sx={{ mt: 2, ml: 2, fontFamily: 'monospace', whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                {formatNetworkFlow(JSON.parse(row.original.path))}
              </Box>
            </Collapse>
          </>
        );
      }
    }] : []),
    ...(columnSelection.event_type ? [{ Header: "Event Type", accessor: "event_type" }] : []),
    ...(columnSelection.size ? [{ Header: "Size", accessor: "size" }] : []),
    ...(columnSelection.total_packets ? [{ Header: "Total Packets", accessor: "total_packets" }] : []),
    ...(columnSelection.total_size ? [{ Header: "Total Size", accessor: "total_size" }] : []),
    ...(columnSelection.src_ip_label ? [{ Header: "Source IP Label", accessor: "src_ip_label" }] : []),
    ...(columnSelection.dst_ip_label ? [{ Header: "Destination IP Label", accessor: "dst_ip_label" }] : []),
  ], [columnSelection, selectedRows, expandedRows, handleRowClick]);

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

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="100vh">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="100vh">
        <Typography color="error">{error}</Typography>
      </Box>
    );
  }

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
          <Box mx={3} mt={3} display="flex" alignItems="center" className="no-print">
            <Typography variant="h4" color="dark" gutterBottom>
              Network Traffic Detail
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
              value={protocolFilter}
              onChange={(e) => setProtocolFilter(e.target.value)}
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
              <MenuItem value="all">All Protocols</MenuItem>
              <MenuItem value="TCP">TCP</MenuItem>
              <MenuItem value="UDP">UDP</MenuItem>
              <MenuItem value="ICMP">ICMP</MenuItem>
            </Select>
            <TextField
              label="Min Packet Size"
              type="number"
              onChange={(e) => setPacketSizeRange((prev) => [Number(e.target.value), prev[1]])}
              sx={{ width: '150px', height: '48px' }}
            />
            <TextField
              label="Max Packet Size"
              type="number"
              onChange={(e) => setPacketSizeRange((prev) => [prev[0], Number(e.target.value)])}
              sx={{ width: '150px', height: '48px' }}
            />
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
                  <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                    <Typography variant="h6">Traffic Packets Over Time</Typography>
                    <Select
                      value={timeInterval}
                      onChange={(e) => setTimeInterval(Number(e.target.value))}
                      size="small"
                    >
                      <MenuItem value={1}>1 minute</MenuItem>
                      <MenuItem value={5}>5 minutes</MenuItem>
                      <MenuItem value={10}>10 minutes</MenuItem>
                      <MenuItem value={30}>30 minutes</MenuItem>
                      <MenuItem value={60}>1 hour</MenuItem>
                    </Select>
                  </Box>
                  <div style={{ height: '400px', width: '100%' }}>
                    <Line
                      data={lineChartData}
                      options={{
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                          legend: {
                            position: 'top',
                          },
                          tooltip: {
                            mode: 'index',
                            intersect: false,
                          },
                        },
                        scales: {
                          y: {
                            beginAtZero: true,
                            title: {
                              display: true,
                              text: 'Packets'
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
                  </div>
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} md={4}>
              <Card>
                <CardContent>
                  <Typography variant="h6">Protocol Distribution</Typography>
                  <div style={{ height: '400px', width: '100%' }}>
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
                            const clickedData = filteredTrafficData.filter(d => d.protocol === pieChartData.labels[dataIndex]);
                            handleDataClick(clickedData);
                          }
                        }
                      }}
                    />
                  </div>
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12}>
              <Card ref={graphContainerRef} sx={{ height: 400, overflow: "hidden" }}>
                <CardContent>
                  <Typography variant="h6">Network Flow Visualization</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
                    Drag nodes to reposition. Click a fixed node to release it. Double-click for details.
                  </Typography>
                  <ForceGraph2D
                    ref={fgRef}
                    graphData={networkFlowData}
                    nodeLabel={(node) => node.label}
                    linkDirectionalArrowLength={3}
                    linkDirectionalArrowRelPos={1}
                    linkWidth={link => Math.sqrt(link.value)}
                    nodeVal={node => Math.sqrt(node.packets) * 1.5}
                    nodeColor={node => node.group === 1 ? "#36A2EB" : "#FF6384"}
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

                      ctx.fillStyle = node.group === 1 ? "rgba(54, 162, 235, 0.2)" : "rgba(255, 99, 132, 0.2)";
                      ctx.fillRect(
                        node.x - bckgDimensions[0] / 2,
                        node.y - bckgDimensions[1] / 2,
                        ...bckgDimensions
                      );

                      ctx.strokeStyle = node.group === 1 ? "rgba(54, 162, 235, 1)" : "rgba(255, 99, 132, 1)";
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
                        const nodeData = filteredTrafficData.find(data => data.id === node.eventId);
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
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12}>
              <Card>
                <CardContent>
                  <Typography variant="h6">Network Traffic Timeline</Typography>
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
                      Click on a row to view detailed information in the side panel
                    </Typography>
                  </Box>
                  <Box mb={2}>
                    <Typography variant="subtitle2" gutterBottom>Select Columns:</Typography>
                    <FormGroup row>
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
                    </FormGroup>
                  </Box>
                  {formattedTrafficData.length > 0 ? (
                    <DataTable
                      table={{
                        columns: columns,
                        rows: formattedTrafficData,
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
                    />
                  ) : (
                    <Typography>No data available</Typography>
					)}
                </CardContent>
              </Card>
            </Grid>
          </Grid>

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
                          <Typography><strong>Total Packets:</strong> {selectedData.reduce((sum, event) => sum + Number(event.packets), 0)}</Typography>
                          <Typography variant="h6" gutterBottom sx={{ mt: 2 }}>Event List</Typography>
                          <Box sx={{ maxHeight: 'calc(100vh - 300px)', overflowY: 'auto', overflowX: 'hidden' }}>
                            {selectedData.map((event, index) => (
                              <Box key={event.id} sx={{ mb: 2, p: 1, backgroundColor: index % 2 === 0 ? 'background.default' : 'action.hover' }}>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>ID:</strong> {event.id}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Event Type:</strong> {event.event_type}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Timestamp:</strong> {new Date(event.timestamp).toLocaleString()}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Container Name:</strong> {event.container_name}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Source IP:</strong> {event.src_ip}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Source IP Label:</strong> {event.src_ip_label}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Destination IP:</strong> {event.dst_ip}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Destination IP Label:</strong> {event.dst_ip_label}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Protocol:</strong> {event.protocol}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Packets:</strong> {event.packets}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Size:</strong> {event.size}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Total Packets:</strong> {event.total_packets}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Total Size:</strong> {event.total_size}</Typography>
                                <Typography sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap' }}><strong>Path:</strong></Typography>
                                <Box sx={{ mt: 1 }}>
                                  {event.path && event.path.nodes?.map((node, index, array) => (
                                    <React.Fragment key={index}>
                                      <Typography component="span">[{node.id}]</Typography>
                                      {index < array.length - 1 && <Typography component="span"> - </Typography>}
                                    </React.Fragment>
                                  ))}
                                </Box>
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
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Source IP:</strong> {selectedData.src_ip}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Source IP Label:</strong> {selectedData.src_ip_label}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Destination IP:</strong> {selectedData.dst_ip}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Destination IP Label:</strong> {selectedData.dst_ip_label}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Protocol:</strong> {selectedData.protocol}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Packets:</strong> {selectedData.packets}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Size:</strong> {selectedData.size}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Total Packets:</strong> {selectedData.total_packets}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Total Size:</strong> {selectedData.total_size}</Typography>
                          <Typography sx={{ wordBreak: 'break-all' }}><strong>Path:</strong></Typography>
                          <Box sx={{ mt: 1 }}>
                            {selectedData.path && selectedData.path.nodes?.map((node, index, array) => (
                              <React.Fragment key={index}>
                                <Typography component="span">[{node.id}]</Typography>
                                {index < array.length - 1 && <Typography component="span"> - </Typography>}
                              </React.Fragment>
                            ))}
                          </Box>
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
    </DashboardLayout>
  );
}