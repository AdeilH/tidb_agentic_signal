// WebSocket test to verify both market updates and real-time state
const WebSocket = require('ws');

const ws = new WebSocket('ws://127.0.0.1:3333/ws');

let marketUpdateCount = 0;
let realtimeStateCount = 0;

ws.on('open', function open() {
    console.log('✅ WebSocket connected - monitoring for 20 seconds...');
});

ws.on('message', function message(data) {
    try {
        const parsed = JSON.parse(data);
        
        if (parsed.type === 'market_update') {
            marketUpdateCount++;
            console.log(`🔄 Market Update ${marketUpdateCount}:`, {
                dataType: parsed.data?.type,
                symbol: parsed.data?.symbol,
                price: parsed.data?.data?.price
            });
        } else if (parsed.type === 'realtime_state_update') {
            realtimeStateCount++;
            console.log(`📊 Real-time State ${realtimeStateCount}:`, {
                symbolCount: parsed.data?.length,
                timestamp: new Date(parsed.timestamp).toLocaleTimeString()
            });
        } else {
            console.log(`❓ Other message type: ${parsed.type}`);
        }
    } catch (e) {
        console.log(`❌ Parse error:`, e.message);
    }
});

ws.on('error', function error(err) {
    console.error('❌ WebSocket error:', err);
});

ws.on('close', function close() {
    console.log('🔌 WebSocket disconnected');
    console.log(`📈 Total market updates: ${marketUpdateCount}`);
    console.log(`📊 Total real-time state updates: ${realtimeStateCount}`);
});

// Keep the script running for 20 seconds
setTimeout(() => {
    ws.close();
    process.exit(0);
}, 20000);
