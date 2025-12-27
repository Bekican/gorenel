const WebSocket = require('ws');
const wss = new WebSocket.Server({ port: 3000 });

wss.on('connection', ws => {
  console.log('Client connected');
  ws.on('message', msg => {
    console.log('Received:', msg);
    ws.send(`Echo: ${msg}`);
  });
});

console.log('WebSocket server running on ws://localhost:3000');