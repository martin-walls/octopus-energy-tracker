import Chart from "chart.js/auto";
import "chartjs-adapter-date-fns";
import type { ConsumptionReading } from "./types/octopus";

let chartData: { x: number; y: number }[] = [];
let chart: Chart;

export function initChart() {
  const container = document.getElementById("chart") as HTMLCanvasElement;

  chart = new Chart(container, {
    type: "line",
    data: {
      datasets: [
        {
          label: "Demand",
          data: chartData,
        },
      ],
    },
    options: {
      scales: {
        x: {
          type: "time",
        },
      },
    },
  });
}

export function updateChart(reading: ConsumptionReading) {
  chartData.push({
    x: new Date(reading.timestamp).getTime(),
    y: reading.demand,
  });
  chart.update();
}
