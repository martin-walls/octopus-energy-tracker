import { initChart, updateChart } from "./chart.js";
import { ws } from "./ws.js";

initChart();

ws(updateChart);
