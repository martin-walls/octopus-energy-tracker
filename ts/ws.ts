import type { ConsumptionReading } from "./types/octopus";

let socket: WebSocket;

export async function ws(onReading: (r: ConsumptionReading) => void) {
  console.log("Connecting to websocket...");

  socket = new WebSocket("ws://localhost:9090/ws");

  socket.addEventListener("message", (e) => {
    const data: ConsumptionReading = JSON.parse(e.data);

    console.log(`Using ${data.demand}W`);

    const demandSpan = document.getElementById("demand-value");
    if (demandSpan != null) {
      demandSpan.textContent = data.demand.toString();
    }

    onReading(data);
  });

  socket.addEventListener("error", (e) => {
    console.log("Websocket error:", e);
  });

  socket.addEventListener("close", (e) => {
    console.log("Websocket closed:", e);
    console.log("Reconnecting...");
    setTimeout(ws, 1000);
  });
}
