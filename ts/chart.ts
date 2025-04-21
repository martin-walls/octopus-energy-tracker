import Chart from "chart.js/auto";
import "chartjs-adapter-date-fns";
import type { ConsumptionReading } from "./types/octopus";

const container = document.getElementById("chart") as HTMLCanvasElement;

let chartData: { x: number; y: number }[] = [];

const chart = new Chart(container, {
  type: "line",
  data: {
    datasets: [
      {
        label: "Demand (W)",
        data: chartData,
      },
    ],
  },
  options: {
    scales: {
      x: {
        type: "time",
        time: {
          minUnit: "second",
        },
        grid: {
          display: false,
        },
      },
      y: {
        beginAtZero: true,
      },
    },
  },
});

export function updateChart(reading: ConsumptionReading) {
  chartData.push({
    x: new Date(reading.timestamp).getTime(),
    y: reading.demand,
  });
  chart.update();
}
