// Simple WebSocket test to verify data flow
const WebSocket = require('ws');

const ws = new WebSocket('ws://127.0.0.1:3333/ws');

let messageCount = 0;

ws.on('open', function open() {
    console.log('✅ WebSocket connected successfully');
});

ws.on('message', function message(data) {
    messageCount++;
    try {
        const parsed = JSON.parse(data);
        console.log(`📨 Message ${messageCount}:`, {
            type: parsed.type,
            dataType: parsed.data?.type,
            symbol: parsed.data?.symbol || parsed.symbol,
            timestamp: parsed.timestamp || parsed.data?.timestamp
        });
    } catch (e) {
        console.log(`📨 Message ${messageCount} (raw):`, data.toString().substring(0, 100));
    }
});

ws.on('error', function error(err) {
    console.error('❌ WebSocket error:', err);
});

ws.on('close', function close() {
    console.log('🔌 WebSocket disconnected');
    console.log(`Total messages received: ${messageCount}`);
});

// Keep the script running for 30 seconds
setTimeout(() => {
    console.log('⏰ Test timeout - closing connection');
    ws.close();
    process.exit(0);
}, 30000);

console.log('🚀 Starting WebSocket test...');
