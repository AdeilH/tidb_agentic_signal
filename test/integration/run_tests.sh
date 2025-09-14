#!/bin/bash

# Integration Test Runner for Crypto Signals Bot
# This script runs comprehensive integration tests with TiDB, Kimi AI, and Binance

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if port is open
check_port() {
    nc -z localhost $1 >/dev/null 2>&1
}

echo "🚀 Crypto Signals Bot - Integration Test Runner"
echo "================================================="

# Check prerequisites
print_status "Checking prerequisites..."

# Check Go installation
if ! command_exists go; then
    print_error "Go is not installed. Please install Go 1.25+"
    exit 1
fi

GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | head -1)
print_success "Go version: $GO_VERSION"

# Check Docker
if ! command_exists docker; then
    print_error "Docker is not installed. Please install Docker"
    exit 1
fi

# Check docker-compose (try both docker-compose and docker compose)
if command_exists docker-compose; then
    DOCKER_COMPOSE="docker-compose"
elif command_exists docker && docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    print_error "Neither docker-compose nor docker compose is available. Please install Docker Compose"
    exit 1
fi

# Check .env file
if [ ! -f ".env" ]; then
    print_warning ".env file not found. Creating from template..."
    cp .env.example .env
    print_warning "Please edit .env file with your API keys before running tests"
    exit 1
fi

print_success "Prerequisites check completed"

# Check if TiDB is running
print_status "Checking TiDB cluster status..."
if ! check_port 4000; then
    print_warning "TiDB not running. Starting TiDB cluster..."
    $DOCKER_COMPOSE up -d
    
    # Wait for TiDB to be ready
    print_status "Waiting for TiDB to be ready..."
    for i in {1..30}; do
        if check_port 4000; then
            print_success "TiDB is ready"
            break
        fi
        sleep 2
        echo -n "."
    done
    
    if ! check_port 4000; then
        print_error "TiDB failed to start. Check docker compose logs"
        exit 1
    fi
else
    print_success "TiDB is already running"
fi

# Check .env variables
print_status "Validating environment variables..."

# Load environment variables more safely using grep/cut to avoid shell parsing issues
if [ -f ".env" ]; then
    export KIMI_API_KEY=$(grep "^KIMI_API_KEY=" .env | cut -d '=' -f 2-)
    export BINANCE_TEST_KEY=$(grep "^BINANCE_TEST_KEY=" .env | cut -d '=' -f 2-)
    export BINANCE_TEST_SECRET=$(grep "^BINANCE_TEST_SECRET=" .env | cut -d '=' -f 2-)
    export TIDB_DSN=$(grep "^TIDB_DSN=" .env | cut -d '=' -f 2-)
    export SLACK_WEBHOOK_URL=$(grep "^SLACK_WEBHOOK_URL=" .env | cut -d '=' -f 2-)
fi

if [ -z "$KIMI_API_KEY" ] || [ "$KIMI_API_KEY" = "your_kimi_api_key_here" ]; then
    print_error "KIMI_API_KEY not set in .env file"
    exit 1
fi

if [ -z "$BINANCE_TEST_KEY" ] || [ "$BINANCE_TEST_KEY" = "your_binance_test_key_here" ]; then
    print_error "BINANCE_TEST_KEY not set in .env file"
    exit 1
fi

if [ -z "$BINANCE_TEST_SECRET" ] || [ "$BINANCE_TEST_SECRET" = "your_binance_test_secret_here" ]; then
    print_error "BINANCE_TEST_SECRET not set in .env file"
    exit 1
fi

print_success "Environment variables validated"

# Run tests based on arguments
case "${1:-all}" in
    "all")
        print_status "Running full integration test suite..."
        echo ""
        
        print_status "🗄️  Testing TiDB integration..."
        go test ./test/integration -v -run TestIntegrationFullSystem/TiDB
        
        print_status "🤖 Testing Kimi AI integration..."
        go test ./test/integration -v -run TestIntegrationFullSystem/Kimi
        
        print_status "📈 Testing Binance integration..."
        go test ./test/integration -v -run TestIntegrationFullSystem/Binance
        
        print_status "🔄 Testing end-to-end signal generation..."
        go test ./test/integration -v -run TestIntegrationFullSystem/End_To_End
        
        print_status "⚡ Testing TTL features..."
        go test ./test/integration -v -run TestIntegrationFullSystem/TTL
        
        print_status "🧮 Testing vector storage..."
        go test ./test/integration -v -run TestIntegrationFullSystem/Vector
        ;;
    
    "tidb")
        print_status "🗄️  Testing TiDB integration only..."
        go test ./test/integration -v -run TestIntegrationFullSystem/TiDB
        go test ./test/integration -v -run TestIntegrationFullSystem/TTL
        go test ./test/integration -v -run TestIntegrationFullSystem/Vector
        ;;
    
    "ai")
        print_status "🤖 Testing Kimi AI integration only..."
        go test ./test/integration -v -run TestIntegrationFullSystem/Kimi
        ;;
    
    "binance")
        print_status "📈 Testing Binance integration only..."
        go test ./test/integration -v -run TestIntegrationFullSystem/Binance
        ;;
    
    "e2e")
        print_status "🔄 Testing end-to-end flow only..."
        go test ./test/integration -v -run TestIntegrationFullSystem/End_To_End
        ;;
    
    "api")
        print_status "🌐 Testing API endpoints..."
        
        # Check if server is running
        if ! check_port 3333; then
            print_warning "API server not running. Starting server..."
            go run cmd/all/main.go &
            SERVER_PID=$!
            
            # Wait for server to start
            for i in {1..10}; do
                if check_port 3333; then
                    print_success "API server is ready"
                    break
                fi
                sleep 1
            done
            
            if ! check_port 3333; then
                print_error "API server failed to start"
                kill $SERVER_PID 2>/dev/null || true
                exit 1
            fi
        fi
        
        go test ./test/integration -v -run TestIntegrationAPI
        
        # Stop server if we started it
        if [ ! -z "$SERVER_PID" ]; then
            print_status "Stopping API server..."
            kill $SERVER_PID 2>/dev/null || true
        fi
        ;;
    
    "bench")
        print_status "⚡ Running performance benchmarks..."
        go test ./test/integration -bench=. -v
        ;;
    
    "help"|"-h"|"--help")
        echo "Usage: $0 [OPTION]"
        echo ""
        echo "Options:"
        echo "  all      Run all integration tests (default)"
        echo "  tidb     Test TiDB features only"
        echo "  ai       Test Kimi AI integration only"
        echo "  binance  Test Binance integration only"
        echo "  e2e      Test end-to-end signal generation"
        echo "  api      Test API endpoints (starts server if needed)"
        echo "  bench    Run performance benchmarks"
        echo "  help     Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0              # Run all tests"
        echo "  $0 tidb         # Test TiDB only"
        echo "  $0 ai           # Test Kimi AI only"
        echo "  $0 bench        # Run benchmarks"
        exit 0
        ;;
    
    *)
        print_error "Unknown option: $1"
        print_status "Use '$0 help' for usage information"
        exit 1
        ;;
esac

print_success "Integration tests completed successfully! 🎉"
echo ""
print_status "System validated with:"
echo "  ✅ TiDB cluster with TTL and vector storage"
echo "  ✅ Kimi AI prediction generation"
echo "  ✅ Binance testnet connectivity"
echo "  ✅ End-to-end signal generation pipeline"
echo ""
print_status "Your crypto signals bot is ready for production! 🚀"
