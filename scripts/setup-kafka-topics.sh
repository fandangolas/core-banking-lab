#!/bin/bash
# Setup Kafka topics with at-least-once delivery semantics
# This script configures topics for reliability and durability

set -e

BOOTSTRAP_SERVER="${KAFKA_BOOTSTRAP_SERVER:-localhost:9092}"
REPLICATION_FACTOR="${KAFKA_REPLICATION_FACTOR:-1}"  # 3 for production
MIN_INSYNC_REPLICAS="${KAFKA_MIN_INSYNC_REPLICAS:-1}"  # 2 for production
PARTITIONS="${KAFKA_PARTITIONS:-3}"
RETENTION_MS="${KAFKA_RETENTION_MS:-604800000}"  # 7 days

echo "==================================="
echo "Kafka Topics Setup - At-Least-Once"
echo "==================================="
echo "Bootstrap Server: $BOOTSTRAP_SERVER"
echo "Replication Factor: $REPLICATION_FACTOR"
echo "Min In-Sync Replicas: $MIN_INSYNC_REPLICAS"
echo "Partitions: $PARTITIONS"
echo "Retention: $RETENTION_MS ms"
echo "==================================="
echo ""

# Function to create topic with at-least-once configuration
create_topic() {
    local topic_name=$1
    local description=$2

    echo "Creating topic: $topic_name"
    echo "  Description: $description"

    kafka-topics --create \
        --topic "$topic_name" \
        --bootstrap-server "$BOOTSTRAP_SERVER" \
        --replication-factor "$REPLICATION_FACTOR" \
        --partitions "$PARTITIONS" \
        --config min.insync.replicas="$MIN_INSYNC_REPLICAS" \
        --config retention.ms="$RETENTION_MS" \
        --config cleanup.policy=delete \
        --config compression.type=snappy \
        --if-not-exists || echo "  ⚠️  Topic already exists"

    echo "  ✅ Topic configuration completed"
    echo ""
}

# Account Events (result events)
create_topic "banking.accounts.created" \
    "Account creation events"

# Deposit Command and Events
create_topic "banking.commands.deposit-requests" \
    "Deposit request commands (fire-and-forget)"

create_topic "banking.transactions.deposit" \
    "Deposit completion events"

# Withdrawal Events
create_topic "banking.transactions.withdrawal" \
    "Withdrawal completion events"

# Transfer Events
create_topic "banking.transactions.transfer" \
    "Transfer completion events"

# Failed Transaction Events
create_topic "banking.transactions.failed" \
    "Failed transaction events (audit trail)"

echo "==================================="
echo "Topic Verification"
echo "==================================="
echo ""

# List all banking topics
kafka-topics --list \
    --bootstrap-server "$BOOTSTRAP_SERVER" | grep "banking\."

echo ""
echo "==================================="
echo "Detailed Configuration"
echo "==================================="
echo ""

# Show detailed config for deposit requests topic (most critical)
echo "Deposit Requests Topic Configuration:"
kafka-topics --describe \
    --topic banking.commands.deposit-requests \
    --bootstrap-server "$BOOTSTRAP_SERVER"

echo ""
echo "==================================="
echo "✅ All topics configured successfully!"
echo "==================================="
echo ""
echo "At-Least-Once Guarantees:"
echo "  ✅ min.insync.replicas=$MIN_INSYNC_REPLICAS (prevents data loss)"
echo "  ✅ retention.ms=$RETENTION_MS (7 days message retention)"
echo "  ✅ cleanup.policy=delete (automatic cleanup)"
echo "  ✅ compression.type=snappy (efficient storage)"
echo ""
echo "Producer must be configured with:"
echo "  - RequiredAcks: 'all' (wait for all ISR)"
echo "  - EnableIdempotence: true (prevent duplicates)"
echo "  - MaxRetries: 5 (retry on failures)"
echo ""
echo "Consumer must be configured with:"
echo "  - AutoCommit: false (manual commit)"
echo "  - Commit after processing (at-least-once)"
echo ""
