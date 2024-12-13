function configs(labels, datasets, tooltipData, bgColor) {
  return {
    data: {
      labels,
      datasets: [
        {
          label: datasets.label,
          tension: 0.4,
          pointRadius: 0,
          pointBorderColor: "transparent",
          pointBackgroundColor: "rgba(255, 255, 255, .8)",
          borderColor: datasets.borderColor || "rgba(255, 255, 255, 0.8)",
          borderWidth: datasets.borderWidth || 4,
          backgroundColor: bgColor || "transparent",
          fill: true,
          data: datasets.data,
          maxBarThickness: 6,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: {
        duration: 1000,
        easing: "easeInOutQuad",
      },
      plugins: {
        legend: {
          display: false,
        },
        tooltip: {
          callbacks: {
            title: (tooltipItems) => {
              const label = tooltipItems[0].label;
              return tooltipData[label] || label;
            },
            label: (tooltipItems) => {
              const dataPoint = datasets.data[tooltipItems.dataIndex];
              return [`Value: ${dataPoint}`];
            },
          },
        },
      },
      interaction: {
        intersect: false,
        mode: "nearest",
      },
      scales: {
        y: {
          grid: {
            drawBorder: false,
            display: true,
            drawOnChartArea: true,
            drawTicks: false,
            borderDash: [5, 5],
            color: "rgba(255, 255, 255, .2)",
          },
          ticks: {
            display: true,
            color: "#f8f9fa",
            padding: 10,
            font: {
              size: 14,
              weight: 300,
              family: "Roboto",
              style: "normal",
              lineHeight: 2,
            },
          },
        },
        x: {
          grid: {
            drawBorder: false,
            display: false,
            drawOnChartArea: false,
            drawTicks: false,
            borderDash: [5, 5],
          },
          ticks: {
            display: true,
            color: "#f8f9fa",
            padding: 10,
            font: {
              size: 14,
              weight: 300,
              family: "Roboto",
              style: "normal",
              lineHeight: 2,
            },
          },
        },
      },
      onClick: () => {
        // ��릭 시 새 탭에서 지정된 URL 열기
        window.open("https://hactiv-web.run.goorm.site/dashboard/eventalert", "_blank");
      },
    },
  };
}

export default configs;

