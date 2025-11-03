#!/bin/bash
# Native API Performance Test
# Bypasses Docker bridge networking by running API natively on macOS
# while keeping infrastructure (Postgres + Kafka) in Docker

set -e

echo "🚀 Starting native API performance test..."
echo ""

# 1. Stop existing API container if running
echo "1️⃣  Stopping Docker API container..."
docker-compose stop api 2>/dev/null || true
echo ""

# 2. Start infrastructure only
echo "2️⃣  Starting infrastructure (Postgres + Kafka)..."
docker-compose up -d postgres kafka kafka-ui prometheus grafana
echo ""

# 3. Wait for services
echo "3️⃣  Waiting for services to be ready..."
sleep 15
echo ""

# 4. Verify database is ready
echo "4️⃣  Checking PostgreSQL..."
until docker exec banking-postgres pg_isready -U banking >/dev/null 2>&1; do
  echo "   ⏳ Waiting for PostgreSQL..."
  sleep 2
done
echo "   ✅ PostgreSQL ready!"
echo ""

# 5. Verify Kafka is ready
echo "5️⃣  Checking Kafka..."
until docker exec banking-kafka kafka-broker-api-versions --bootstrap-server kafka:9092 >/dev/null 2>&1; do
  echo "   ⏳ Waiting for Kafka..."
  sleep 2
done
echo "   ✅ Kafka ready!"
echo ""

# 6. Build API
echo "6️⃣  Building API..."
go build -o /tmp/banking-api cmd/api/main.go
echo "   ✅ API built!"
echo ""

# 7. Export environment
echo "7️⃣  Configuring environment..."
export SERVER_PORT=8080
export SERVER_HOST=localhost
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=banking
export DB_USER=banking
export DB_PASSWORD=banking_secure_pass_2024
export DB_SSLMODE=disable
export DB_MAX_OPEN_CONNS=100
export DB_MAX_IDLE_CONNS=25
export KAFKA_BROKERS=localhost:9092
export KAFKA_ENABLED=true
export KAFKA_ENABLE_IDEMPOTENCE=false
export KAFKA_REQUIRED_ACKS=all
export LOG_LEVEL=warn
export LOG_FORMAT=json
echo "   ✅ Environment configured!"
echo ""

# 8. Start API in background
echo "8️⃣  Starting API natively..."
/tmp/banking-api > /tmp/banking-api.log 2>&1 &
API_PID=$!
echo "   ✅ API started (PID: $API_PID)"
echo ""

# 9. Wait for API to be ready
echo "9️⃣  Waiting for API to be ready..."
for i in {1..10}; do
  if curl -sf http://localhost:8080/metrics >/dev/null 2>&1; then
    echo "   ✅ API is ready!"
    break
  fi
  if [ $i -eq 10 ]; then
    echo "   ❌ API failed to start. Check logs at /tmp/banking-api.log"
    kill $API_PID 2>/dev/null || true
    exit 1
  fi
  echo "   ⏳ Waiting for API... ($i/10)"
  sleep 2
done
echo ""

# 10. Show architecture
echo "📊 Performance Test Architecture:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  k6 Client → macOS Network → Native Go Process"
echo "                                  ↓"
echo "                          PostgreSQL (Docker)"
echo "                          Kafka (Docker)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ Zero Docker bridge overhead"
echo "  ✅ Native network performance"
echo "  ✅ Maximum throughput potential"
echo ""

# 11. Run k6 test
echo "🔥 Running k6 extreme test (5000 VUs)..."
echo ""
k6 run k6-tests/scenarios/extreme.js 2>&1 | tee /tmp/k6-native-extreme.log

# 12. Cleanup
echo ""
echo "🧹 Cleaning up..."
echo "   Stopping native API (PID: $API_PID)..."
kill $API_PID 2>/dev/null || true
echo "   ✅ API stopped"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Test completed!"
echo ""
echo "📁 Logs available at:"
echo "   • k6 results: /tmp/k6-native-extreme.log"
echo "   • API logs: /tmp/banking-api.log"
echo ""
echo "📊 Compare results with Docker bridge test:"
echo "   • Docker: /tmp/k6-extreme-after-fix.log"
echo "   • Native: /tmp/k6-native-extreme.log"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
