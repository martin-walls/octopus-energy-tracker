import type { ConsumptionReading } from "./types/octopus";

let socket: WebSocket;

async function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function ws() {
  console.log("Connecting to websocket...");

  socket = new WebSocket("ws://localhost:9090/ws");

  // socket.addEventListener("open", () => {
  //   socket.send("Hello server!");
  // });

  socket.addEventListener("message", (e) => {
    const data: ConsumptionReading = JSON.parse(e.data);

    console.log(`Using ${data.demand}W`);

    const demandSpan = document.getElementById("demand-value");
    if (demandSpan != null) {
      demandSpan.textContent = data.demand.toString();
    }
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

document.getElementById("btn-connect")?.addEventListener("click", ws);
