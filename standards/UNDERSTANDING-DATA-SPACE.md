# Understanding Data Space - AI Scaffolder Database Architecture

## 📊 Overview

This guide explains how to choose the right database technology for different types of IoT data. Using a smart commercial building management system as our example, we'll explore how different data types flow through our polyglot architecture.

---

## 🏢 Use Case: Smart Building Management System (SmartBuild)

**Scenario**: A multi-building enterprise with 50,000+ sensors across 100 buildings in US East, US West, and Europe.

- **Building A (NYC)**: 15,000 sensors (temperature, occupancy, energy meters, access logs)
- **Building B (LA)**: 12,000 sensors
- **Building C (London)**: 8,000 sensors
- Plus 65 other buildings worldwide

---

## 1. **RELATIONAL DATA (SQL Server / Azure SQL MI)**

### What Goes Here?
Structured, normalized data with relationships - data that needs ACID transactions and complex joins.

### SmartBuild Examples:

```sql
-- User Management (Normalized Structure)
CREATE TABLE users (
    user_id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    building_id UUID NOT NULL,
    role ENUM('admin', 'manager', 'occupant', 'maintenance'),
    created_at TIMESTAMP,
    FOREIGN KEY (building_id) REFERENCES buildings(building_id)
);

-- Building Registry
CREATE TABLE buildings (
    building_id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    region ENUM('US-EAST', 'US-WEST', 'EU'),
    total_floors INT,
    operational_since DATE,
    main_contact_id UUID,
    FOREIGN KEY (main_contact_id) REFERENCES users(user_id)
);

-- Device Master Data
CREATE TABLE devices (
    device_id UUID PRIMARY KEY,
    building_id UUID NOT NULL,
    device_type VARCHAR(50),  -- 'temp_sensor', 'occupancy', 'access_point'
    zone_id UUID,
    serial_number VARCHAR(100) UNIQUE,
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    installation_date DATE,
    last_maintenance DATE,
    status ENUM('active', 'inactive', 'maintenance'),
    coordinates POINT,  -- (lat, lon) for geospatial queries
    FOREIGN KEY (building_id) REFERENCES buildings(building_id),
    FOREIGN KEY (zone_id) REFERENCES zones(zone_id)
);

-- Zone/Area Definition
CREATE TABLE zones (
    zone_id UUID PRIMARY KEY,
    building_id UUID NOT NULL,
    zone_name VARCHAR(255),  -- "Floor 3 - Conference Wing"
    zone_type ENUM('floor', 'wing', 'room', 'datacenter'),
    responsible_manager_id UUID,
    priority_level INT,  -- 1=critical, 5=routine
    FOREIGN KEY (building_id) REFERENCES buildings(building_id),
    FOREIGN KEY (responsible_manager_id) REFERENCES users(user_id)
);

-- Access Control Rules
CREATE TABLE access_rules (
    rule_id UUID PRIMARY KEY,
    building_id UUID NOT NULL,
    door_id UUID NOT NULL,
    user_id UUID NOT NULL,
    access_level INT,
    valid_from DATE,
    valid_until DATE,
    days_of_week VARCHAR(10),  -- '1,2,3,4,5' for Mon-Fri
    time_from TIME,
    time_until TIME,
    FOREIGN KEY (building_id) REFERENCES buildings(building_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Maintenance Schedules (Business Logic)
CREATE TABLE maintenance_schedules (
    schedule_id UUID PRIMARY KEY,
    device_id UUID NOT NULL,
    schedule_type ENUM('preventive', 'reactive'),
    frequency_days INT,
    last_maintenance TIMESTAMP,
    next_scheduled TIMESTAMP,
    assigned_technician_id UUID,
    status ENUM('pending', 'in_progress', 'completed'),
    FOREIGN KEY (device_id) REFERENCES devices(device_id),
    FOREIGN KEY (assigned_technician_id) REFERENCES users(user_id)
);
```

### Key Characteristics:
- ✅ **Structured schema** with relationships
- ✅ **Transaction safety** - ACID compliance
- ✅ **Real-time consistency** - immediate updates
- ✅ **Complex queries** - JOINs across multiple tables
- ✅ **Business logic** - access rules, user relationships

### Regional Considerations:
```
US-EAST Region:
- Primary: Azure SQL MI in East US
- Replicas: Read replicas for analytics
- Backup: Cross-region backup to West US

US-WEST Region:
- Autonomous: Local Azure SQL MI
- Sync: Replication from US-EAST for reference data
- Failover: Auto-failover groups (30-second RTO)

EU Region:
- GDPR compliant: EU data residency required
- Encryption: Always TLS in transit + TDE at rest
- Segregation: Separate EU-only Azure SQL MI
- Audit: Azure SQL Auditing to Log Analytics
```

---

## 2. **HIERARCHICAL/DOCUMENT DATA (MongoDB Atlas)**

### What Goes Here?
Semi-structured data, hierarchical configurations, flexible schemas, device profiles, and settings.

### SmartBuild Examples:

```json
// Device Profile - Flexible Configuration
{
  "_id": ObjectId("507f1f77bcf86cd799439011"),
  "device_id": "uuid-123",
  "building_id": "uuid-building-456",
  "zone_id": "uuid-zone-789",
  "device_type": "temperature_sensor",
  "vendor": "SensorTech",
  "model": "ST-500-Pro",
  "metadata": {
    "serial": "ST500-000123",
    "firmware_version": "2.4.1",
    "hardware_version": "Rev-B",
    "installed_date": ISODate("2023-06-15"),
    "warranty_expiry": ISODate("2025-06-15")
  },
  "capabilities": {
    "sensors": [
      {
        "sensor_type": "temperature",
        "unit": "celsius",
        "accuracy": 0.5,
        "range": { "min": -40, "max": 80 },
        "polling_interval_ms": 5000,
        "transmission_interval_ms": 30000
      },
      {
        "sensor_type": "humidity",
        "unit": "percent_rh",
        "accuracy": 3,
        "range": { "min": 0, "max": 100 },
        "polling_interval_ms": 5000
      }
    ]
  },
  "communication": {
    "protocol": "mqtt",
    "broker_url": "mqtt.building-a.internal:1883",
    "topics": {
      "publish": "building/nyc/floor3/zone-a/sensor-123/telemetry",
      "subscribe": "building/nyc/floor3/zone-a/sensor-123/commands"
    },
    "update_frequency_minutes": 5
  },
  "configuration": {
    "calibration": {
      "temp_offset": 0.2,
      "humidity_offset": -2,
      "last_calibrated": ISODate("2025-01-10")
    },
    "alerts": {
      "high_temp_threshold": 28,
      "low_temp_threshold": 16,
      "high_humidity_threshold": 70,
      "low_humidity_threshold": 30
    },
    "features": {
      "self_healing": true,
      "mesh_enabled": true,
      "battery_monitoring": true,
      "local_processing": true
    }
  },
  "location": {
    "building": "NYC-Office",
    "floor": 3,
    "zone": "Conference-Wing-A",
    "room": "Meeting-Room-301",
    "gps": { "type": "Point", "coordinates": [-74.0060, 40.7128] },
    "indoor_map": { "zone_id": "uuid-123", "coordinates": [42.5, 87.3] }
  },
  "performance_profile": {
    "battery_capacity_mah": 5000,
    "expected_battery_life_months": 24,
    "low_battery_warning_percent": 15,
    "power_consumption_ma": 15
  },
  "last_seen": ISODate("2025-01-15T14:32:00Z"),
  "status": "active",
  "version": 3
}

// Building Configuration - Hierarchical
{
  "_id": ObjectId("507f1f77bcf86cd799439012"),
  "building_id": "uuid-building-456",
  "name": "SmartBuild HQ - NYC",
  "address": "123 Tech Avenue, New York, NY 10001",
  "region": "US-EAST",
  "floors": [
    {
      "floor_number": 1,
      "floor_name": "Ground Floor - Lobby & Retail",
      "total_area_sqft": 50000,
      "zones": [
        {
          "zone_id": "uuid-zone-lobby",
          "zone_name": "Main Lobby",
          "area_sqft": 5000,
          "occupancy_capacity": 200,
          "devices": [
            { "device_id": "uuid-temp-001", "type": "temperature_sensor" },
            { "device_id": "uuid-occ-001", "type": "occupancy_sensor" },
            { "device_id": "uuid-hvac-001", "type": "hvac_controller" }
          ]
        },
        {
          "zone_id": "uuid-zone-retail",
          "zone_name": "Retail Space",
          "area_sqft": 45000,
          "occupancy_capacity": 500,
          "devices": [...]
        }
      ]
    },
    {
      "floor_number": 2,
      "floor_name": "First Floor - Offices",
      "zones": [...]
    }
  ],
  "hvac_system": {
    "type": "VAV",
    "zones_served": 45,
    "chillers": [
      {
        "chiller_id": "ch-001",
        "model": "Carrier-19DV",
        "capacity_tons": 250,
        "region_served": "US-EAST",
        "redundancy": "dual"
      }
    ]
  },
  "energy_management": {
    "meter_count": 150,
    "sub_meters": true,
    "demand_response_capable": true,
    "renewable_integration": {
      "solar_panels": 500,
      "kw_capacity": 125,
      "battery_storage_kwh": 300
    }
  }
}

// Automation Rules - Dynamic Configuration
{
  "_id": ObjectId("507f1f77bcf86cd799439013"),
  "rule_id": "uuid-rule-123",
  "building_id": "uuid-building-456",
  "rule_name": "Conference Room Energy Optimization",
  "description": "Reduce HVAC when room is unoccupied",
  "enabled": true,
  "priority": 2,
  "triggers": [
    {
      "type": "sensor_reading",
      "sensor_type": "occupancy",
      "condition": "no_motion",
      "duration_minutes": 15
    }
  ],
  "conditions": [
    {
      "type": "time_based",
      "business_hours": true,
      "working_days": "monday-friday"
    },
    {
      "type": "threshold",
      "parameter": "outside_temperature",
      "operator": "less_than",
      "value": 50  // Fahrenheit
    }
  ],
  "actions": [
    {
      "type": "hvac_adjust",
      "target_devices": ["uuid-hvac-001", "uuid-hvac-002"],
      "action": "reduce_setpoint",
      "setpoint_adjustment": -3
    },
    {
      "type": "notification",
      "recipients": ["manager@smartbuild.com"],
      "message": "Conference room unoccupied, HVAC reduced"
    }
  ],
  "created_by": "admin@smartbuild.com",
  "created_at": ISODate("2025-01-10"),
  "modified_at": ISODate("2025-01-15"),
  "version": 5
}
```

### Key Characteristics:
- ✅ **Flexible schema** - add fields anytime
- ✅ **Nested structures** - hierarchical data natural
- ✅ **Versioning** - track configuration changes
- ✅ **Easy to scale** - horizontal sharding
- ✅ **Embedded arrays** - relationships without JOINs

### Regional Considerations:
```
MongoDB Atlas Global Cluster Topology:

US-EAST (Primary - Virginia):
- Cluster Tier: M30 or higher (dedicated)
- Replica Set: 3 nodes (Primary + 2 Secondaries)
- Shard Key: building_id (ensures data locality)
- Read Preference: primaryPreferred
- Write Concern: majority
- Backup: Continuous cloud backup with point-in-time recovery
- Oplog: 24-hour window for change streams

US-WEST (Secondary - Oregon):
- Cluster Tier: M30 (matched to US-EAST)
- Global Cluster Zone: us-west-2
- Zone Sharding: building_id prefix routing
- Local Reads: Enabled for low-latency queries
- Cross-region replication lag: < 100ms typical

EU (Isolated - Frankfurt):
- Cluster Tier: M30 (dedicated for GDPR)
- Isolated Cluster: Separate Atlas project for EU data
- Encryption: AES-256 at rest + TLS 1.3 in transit
- GDPR Compliance: EU-only data residency
- No cross-region replication to non-EU zones
- TTL Indexes: Document expiry for right-to-be-forgotten
- Audit Logging: Enabled (Atlas audit logs)

MongoDB Atlas Features Used:
- Change Streams: Real-time event notifications
- Atlas Search: Full-text search on device profiles
- Atlas Triggers: Serverless functions for automation
- Atlas Data Federation: Query across clusters
- Online Archive: Automatic archival to cheaper storage
```

---

## 3. **TIME-SERIES DATA (ScyllaDB)**

### What Goes Here?
Sensor readings, measurements, metrics - high-volume data with time dimension.

### SmartBuild Examples:

```cql
-- ScyllaDB Time-Series Table for Sensor Telemetry
-- Designed for high-throughput writes (3B readings/day)
CREATE TABLE IF NOT EXISTS device_telemetry (
    building_id UUID,
    date DATE,  -- Partition key for time-based bucketing
    time TIMESTAMP,
    device_id UUID,
    metric_type TEXT,  -- 'temperature', 'humidity', 'power'
    value DECIMAL,
    unit TEXT,
    raw_value INT,
    quality_score INT,
    PRIMARY KEY ((building_id, date), time, device_id, metric_type)
) WITH 
    CLUSTERING ORDER BY (time DESC) 
    AND compaction = { 'class': 'TimeWindowCompactionStrategy', 'compaction_window_size': 1, 'compaction_window_unit': 'DAYS' }
    AND default_time_to_live = 7776000;  -- 90-day retention

-- Insert telemetry data from 50,000 sensors
INSERT INTO device_telemetry (building_id, date, time, device_id, metric_type, value, unit, quality_score)
VALUES 
    (550e8400-e29b-41d4-a716-446655440000, 2025-01-15, 2025-01-15T14:32:00Z, 60e8400-e29b-41d4-a716-446655440000, 'temperature', 21.5, 'celsius', 95) USING TTL 7776000,
    (550e8400-e29b-41d4-a716-446655440000, 2025-01-15, 2025-01-15T14:32:00Z, 61e8400-e29b-41d4-a716-446655440001, 'temperature', 22.1, 'celsius', 98) USING TTL 7776000,
    (550e8400-e29b-41d4-a716-446655440000, 2025-01-15, 2025-01-15T14:32:00Z, 60e8400-e29b-41d4-a716-446655440000, 'humidity', 45.2, 'percent', 92) USING TTL 7776000;

-- Query latest readings (ScyllaDB clustering order DESC)
SELECT * FROM device_telemetry 
WHERE building_id = 550e8400-e29b-41d4-a716-446655440000 
  AND date = 2025-01-15 
  AND device_id = 60e8400-e29b-41d4-a716-446655440000
ORDER BY time DESC LIMIT 10;

-- Create materialized view for hourly aggregations
CREATE MATERIALIZED VIEW hourly_temperature_avg AS
SELECT
    building_id,
    date,
    dateOf(time) as day,
    hourOf(time) as hour,
    device_id,
    avg(value) as avg_temperature,
    min(value) as min_temperature,
    max(value) as max_temperature,
    count(*) as reading_count,
    time
FROM device_telemetry
WHERE metric_type = 'temperature'
PRIMARY KEY ((building_id, day, hour), device_id, time)
WITH CLUSTERING ORDER BY (device_id ASC, time DESC);

-- Compaction strategy tuned for time-series (TWCS)
-- Automatically merges data from same time window
-- Old data compacted separately, ready for archival
ALTER TABLE device_telemetry WITH
    compaction = {
        'class': 'TimeWindowCompactionStrategy',
        'compaction_window_size': 1,
        'compaction_window_unit': 'DAYS'
    };

-- Query last hour temperature in NYC building
SELECT time, device_id, value
FROM device_telemetry
WHERE building_id = 550e8400-e29b-41d4-a716-446655440000
  AND date = 2025-01-15 
  AND metric_type = 'temperature'
ORDER BY time DESC;

-- Anomaly Detection: Temperature deviation from hourly average
SELECT 
    dt.time,
    dt.device_id,
    dt.value as current_temp,
    hta.avg_temperature,
    abs(dt.value - hta.avg_temperature) as deviation
FROM device_telemetry dt
WHERE dt.building_id = 550e8400-e29b-41d4-a716-446655440000
  AND abs(dt.value - hta.avg_temperature) > 5  -- Threshold
  AND dt.date = 2025-01-15;

-- 3. Daily Summary for Energy Analysis
SELECT
    DATE(time) as day,
    building_id,
    AVG(CASE WHEN metric_type = 'temperature' THEN value END) as avg_temp,
    AVG(CASE WHEN metric_type = 'power_consumption' THEN value END) as avg_power_kw,
    SUM(CASE WHEN metric_type = 'power_consumption' THEN value END) * 0.04167 as energy_kwh
FROM device_telemetry
WHERE time > NOW() - INTERVAL '30 days'
GROUP BY day, building_id
ORDER BY day DESC;
```

### Key Characteristics:
- ✅ **Optimized for time-series** - automatic partitioning by time
- ✅ **Massive scale** - handles trillions of data points
- ✅ **Compression** - reduces storage by 90%+ over time
- ✅ **Continuous aggregates** - pre-computed summaries
- ✅ **Real-time queries** - millisecond response times

### Regional Considerations:
```
TimescaleDB Deployment:

US-EAST Region:
- Hypertable chunks: 1 day chunks for hot data, 7-day for warm
- Compression: After 30 days (older data 90% smaller)
- Retention: Keep raw data 1 year, compressed 3 years
- Backup: Hourly incremental backups, daily full backups
- Scale: Auto-scaling horizontal read replicas for queries

US-WEST Region:
- Local TimescaleDB: Independent hot data storage
- Streaming replication: To US-EAST for historical archive
- Failover: RTO 5 minutes if primary fails

EU Region:
- GDPR: Separate EU-only cluster
- Right to erasure: TTL based cleanup jobs
- Data sovereignty: No sync to non-EU regions
```

---

## 4. **EVENT DATA (ScyllaDB - Time-Series Event Store)**

### What Goes Here?
Immutable event log - what happened, when, and why. Perfect for audit trails, state changes, and event sourcing.

### SmartBuild Examples:

```cql
-- ScyllaDB Event Store with Time-Series Partitioning
CREATE KEYSPACE IF NOT EXISTS events_store
WITH replication = {'class': 'NetworkTopologyStrategy', 'us_east': 3, 'us_west': 3, 'eu': 3};

-- Event Log Table (append-only)
CREATE TABLE events_store.device_events (
    event_date DATE,                    -- Partition key (daily partitions)
    event_time TIMESTAMP,               -- Clustering key (time order)
    event_id UUID,                      -- Clustering key (uniqueness)
    device_id UUID,
    building_id UUID,
    region TEXT,
    event_type TEXT,                    -- 'status_change', 'alert', 'reading_anomaly'
    event_category TEXT,                -- 'telemetry', 'maintenance', 'security'
    severity TEXT,                      -- 'info', 'warning', 'critical'
    source TEXT,                        -- What generated event
    old_value TEXT,
    new_value TEXT,
    metadata MAP<TEXT, TEXT>,           -- Additional context
    user_id UUID,                       -- Who/what triggered it
    PRIMARY KEY ((event_date, building_id), event_time DESC, event_id)
) WITH CLUSTERING ORDER BY (event_time DESC);

-- Index for queries by device
CREATE INDEX idx_device_events ON events_store.device_events (device_id);
CREATE INDEX idx_event_type ON events_store.device_events (event_type);
CREATE INDEX idx_severity ON events_store.device_events (severity);

-- CQL Examples - Event Insertion and Querying

-- 1. Device comes online (Status Change Event)
INSERT INTO events_store.device_events (
    event_date, event_time, event_id, device_id, building_id, region,
    event_type, event_category, severity, source, old_value, new_value, user_id
) VALUES (
    '2025-01-15'::date,
    '2025-01-15T14:32:00Z'::timestamp,
    uuid(),
    'uuid-sensor-001'::uuid,
    'uuid-building-456'::uuid,
    'US-EAST',
    'status_change',
    'telemetry',
    'info',
    'device_heartbeat',
    'offline',
    'online',
    null
);

-- 2. Temperature Anomaly Alert
INSERT INTO events_store.device_events (
    event_date, event_time, event_id, device_id, building_id, region,
    event_type, event_category, severity, source, new_value,
    metadata, user_id
) VALUES (
    '2025-01-15'::date,
    '2025-01-15T14:35:22Z'::timestamp,
    uuid(),
    'uuid-sensor-temp-042'::uuid,
    'uuid-building-456'::uuid,
    'US-EAST',
    'reading_anomaly',
    'telemetry',
    'warning',
    'anomaly_detector',
    '45.2',
    {
        'expected_range': '18-26C',
        'zone': 'Floor-3-Conference',
        'threshold_exceeded': 'high',
        'deviation_percent': '73.8',
        'trend': 'rising_rapidly'
    },
    null
);

-- 3. Maintenance Task Assigned (Maintenance Event)
INSERT INTO events_store.device_events (
    event_date, event_time, event_id, device_id, building_id, region,
    event_type, event_category, severity, source, new_value,
    metadata, user_id
) VALUES (
    '2025-01-15'::date,
    '2025-01-15T09:00:00Z'::timestamp,
    uuid(),
    'uuid-hvac-001'::uuid,
    'uuid-building-456'::uuid,
    'US-EAST',
    'maintenance_assigned',
    'maintenance',
    'info',
    'maintenance_scheduler',
    'SCHEDULED',
    {
        'maintenance_type': 'preventive',
        'assigned_technician': 'tech-john-smith',
        'estimated_duration_hours': '2',
        'parts_needed': 'filter_kit_hc250,thermal_paste',
        'priority': 'high'
    },
    'admin@smartbuild.com'::uuid
);

-- Query Examples:

-- 1. All events for a device in last 24 hours
SELECT event_time, event_type, severity, new_value
FROM events_store.device_events
WHERE event_date = '2025-01-15'
  AND device_id = 'uuid-sensor-001'::uuid
ALLOW FILTERING;

-- 2. All critical events in building today
SELECT event_time, device_id, event_type, severity
FROM events_store.device_events
WHERE event_date = '2025-01-15'
  AND building_id = 'uuid-building-456'::uuid
  AND severity = 'critical'
ALLOW FILTERING
ORDER BY event_time DESC
LIMIT 100;

-- 3. Event timeline for debugging (event sourcing replay)
SELECT event_time, event_id, old_value, new_value, user_id
FROM events_store.device_events
WHERE event_date IN ('2025-01-14'::date, '2025-01-15'::date)
  AND device_id = 'uuid-hvac-001'::uuid
ORDER BY event_time DESC
LIMIT 50;

-- 4. Audit trail for access control changes
SELECT event_time, event_type, user_id, metadata
FROM events_store.device_events
WHERE event_date = '2025-01-15'
  AND event_type = 'access_rule_modified'
ORDER BY event_time DESC;
```

### Key Characteristics:
- ✅ **Append-only** - never overwrite, perfect for compliance
- ✅ **Event sourcing** - replay history to any point in time
- ✅ **Immutable audit trail** - meet regulatory requirements
- ✅ **Time-range queries** - efficient daily partitions
- ✅ **Distributed** - data replication across regions
- ✅ **Search patterns** - flexible indexes for forensics

### Regional Considerations:
```
ScyllaDB Cluster Topology:

US-EAST:
- Replication Factor: 3 (consistent quorum)
- Data Centers: NYC, Boston, Philadelphia
- Consistency Level: QUORUM (strong consistency)
- Event Retention: 7 years (compliance requirement)

US-WEST:
- Replication Factor: 3
- Data Centers: SF, LA, Seattle
- Autonomous cluster: No sync with US-EAST
- Local events: Only US-WEST region events

EU:
- Replication Factor: 3
- Data Centers: Frankfurt, Amsterdam, Dublin
- GDPR mode: Right to erasure via TTL
- Encryption: TDE at rest, TLS in transit
- Compliance: ISO 27001, SOC 2
```

---

## 5. **MESSAGE STREAMING (Apache Kafka)**

### What Goes Here?
High-throughput real-time data streams - decoupling producers from consumers.

### SmartBuild Kafka Topics:

```yaml
Topics Architecture:

Topic: device-telemetry
  Partitions: 200 (based on 50,000 devices / 250 per partition)
  Replication Factor: 3
  Retention: 24 hours (1 TB per day at 10K msgs/sec)
  Schema:
    {
      "device_id": "uuid",
      "timestamp": "2025-01-15T14:32:00Z",
      "metric_type": "temperature",
      "value": 21.5,
      "unit": "celsius",
      "building_id": "uuid",
      "region": "US-EAST"
    }
  Consumers:
    - TimescaleDB Writer (bulk ingest 60-second batches)
    - Real-time Alerting Service
    - Stream Processor (anomaly detection)

Topic: device-events
  Partitions: 50 (events are lower volume than telemetry)
  Replication Factor: 3
  Retention: 30 days
  Schema:
    {
      "event_id": "uuid",
      "device_id": "uuid",
      "event_type": "status_change|alert|configuration",
      "severity": "info|warning|critical",
      "timestamp": "2025-01-15T14:32:00Z",
      "old_value": "...",
      "new_value": "...",
      "building_id": "uuid"
    }
  Consumers:
    - ScyllaDB Event Store Writer
    - Event Processor (rules engine)
    - WebSocket Broadcast (real-time UI updates)

Topic: access-events
  Partitions: 25 (security events - moderate volume)
  Replication Factor: 3
  Retention: 90 days (compliance)
  Schema:
    {
      "event_id": "uuid",
      "access_point_id": "uuid",
      "user_id": "uuid",
      "timestamp": "2025-01-15T14:32:00Z",
      "access_granted": true|false,
      "deny_reason": "expired_credentials|no_access_right",
      "building_id": "uuid",
      "region": "US-EAST"
    }
  Consumers:
    - Access Audit Logger
    - Security Team Alerts
    - ScyllaDB Event Store

Topic: maintenance-tasks
  Partitions: 10 (low volume)
  Replication Factor: 3
  Retention: 365 days (historical records)
  Schema:
    {
      "task_id": "uuid",
      "device_id": "uuid",
      "task_type": "preventive|corrective|inspection",
      "assigned_technician": "uuid",
      "status": "pending|in_progress|completed|failed",
      "timestamp": "2025-01-15T14:32:00Z",
      "estimated_duration_minutes": 120,
      "parts_needed": ["part1", "part2"]
    }
  Consumers:
    - Maintenance Dashboard
    - Technician Mobile App
    - MongoDB Device Profile Updater

Topic: notifications (outbound)
  Partitions: 50 (fanout - many subscribers)
  Replication Factor: 3
  Retention: 48 hours
  Schema:
    {
      "notification_id": "uuid",
      "recipient_id": "uuid",
      "channel": "email|sms|push|webhook",
      "subject": "...",
      "body": "...",
      "priority": "low|normal|high|critical",
      "timestamp": "2025-01-15T14:32:00Z",
      "building_id": "uuid"
    }
  Consumers:
    - Email Service
    - SMS Gateway
    - Push Notification Service
    - Webhook Dispatcher
```

### Consumer Groups & Partitioning Strategy:

```
Consumer Group: timeseries-writers
  Topic: device-telemetry
  Instances: 4 (one per US region partition)
  Processing: Bulk insert to TimescaleDB every 60 seconds
  Max Records: 50,000 per batch
  Lag SLA: < 2 minutes

Consumer Group: real-time-alerts
  Topic: device-telemetry, device-events
  Instances: 8 (horizontal scaling for latency)
  Processing: Check thresholds, trigger alerts < 500ms
  Lag SLA: < 5 seconds
  Parallelism: Thread pool per partition

Consumer Group: event-store-writers
  Topic: device-events, access-events
  Instances: 2 (ScyllaDB batch writing)
  Processing: Write to event store
  Batch Size: 1,000 events
  Lag SLA: < 30 seconds

Consumer Group: ui-broadcasters
  Topic: device-events, device-telemetry (filtered)
  Instances: 3 (WebSocket connections)
  Processing: Send to connected clients via WebSocket
  Lag SLA: < 1 second
  Fan-out: 1,000+ concurrent connections per instance
```

### Partition Key Strategy:

```csharp
// Producer Example: SmartBuild Device Ingest Service
public class KafkaDeviceTelemetryProducer
{
    private readonly IProducer<string, string> _producer;

    public async Task ProduceTelemetryAsync(DeviceReading reading)
    {
        // Partition Key = building_id ensures:
        // 1. All events for a building go to same partition
        // 2. Ordering guaranteed per building
        // 3. Easy consumer scaling per building
        var message = new Message<string, string>
        {
            Key = reading.BuildingId.ToString(),  // Partition key
            Value = JsonConvert.SerializeObject(new
            {
                device_id = reading.DeviceId,
                timestamp = reading.Timestamp,
                value = reading.Value,
                metric_type = reading.MetricType,
                building_id = reading.BuildingId,
                region = GetRegion(reading.BuildingId)
            })
        };

        await _producer.ProduceAsync("device-telemetry", message);
    }
}

// Consumer Example: Real-time Alerts
public class AlertingConsumer
{
    private readonly IConsumer<string, string> _consumer;
    private readonly AlertService _alertService;

    public async Task ConsumeAsync(CancellationToken ct)
    {
        _consumer.Subscribe("device-telemetry");

        while (!ct.IsCancellationRequested)
        {
            var consumeResult = _consumer.Consume(TimeSpan.FromSeconds(1));
            if (consumeResult == null) continue;

            var reading = JsonConvert.DeserializeObject<DeviceReading>(consumeResult.Message.Value);
            
            // Check threshold (building-specific rules)
            if (ShouldAlert(reading))
            {
                await _alertService.SendAlertAsync(reading);
            }

            _consumer.CommitAsync(consumeResult);
        }
    }

    private bool ShouldAlert(DeviceReading reading)
    {
        return reading.MetricType == "temperature" && reading.Value > 30;
    }
}
```

### Topic Decision Matrix:

| Data Type | Topic | Volume | Consumers | Retention |
|-----------|-------|--------|-----------|-----------|
| Telemetry | device-telemetry | 10K msgs/sec | 2-3 | 24 hours |
| Events | device-events | 500 msgs/sec | 3-4 | 30 days |
| Access Logs | access-events | 200 msgs/sec | 2 | 90 days |
| Maintenance | maintenance-tasks | 50 msgs/sec | 1-2 | 365 days |
| Notifications | notifications | 1K msgs/sec | 4-5 | 48 hours |

---

## 6. **CACHING LAYER (Redis)**

### What Goes Here?
Hot data, session state, frequently accessed lookups - reduce database load.

### SmartBuild Redis Usage:

```
Redis Data Structures & TTLs:

1. SESSION CACHE (TTL: 1 hour)
   Key: session:{session_id}
   Type: Hash
   Purpose: Store user session data
   Value:
   {
       "user_id": "uuid-123",
       "building_id": "uuid-456",
       "last_activity": "2025-01-15T14:32:00Z",
       "permissions": ["read_telemetry", "manage_devices"],
       "region": "US-EAST"
   }
   Strategy:
   - Set on login
   - Update on_activity
   - Delete on logout
   - Evict after 1 hour inactivity

2. DEVICE CONFIG CACHE (TTL: 6 hours)
   Key: device:config:{device_id}
   Type: String (JSON)
   Purpose: Cache device profile from MongoDB
   Size per device: ~2-5 KB
   Total: 50,000 devices × 3 KB = 150 MB
   Strategy:
   - Lazy load on first read
   - Invalidate on config change
   - Refresh before TTL on update
   
   Example:
   Redis GET device:config:uuid-sensor-001
   Returns: {
       "polling_interval_ms": 5000,
       "thresholds": {"high": 30, "low": 10},
       "alerts_enabled": true,
       "location": "Floor-3-Conference"
   }

3. LATEST TELEMETRY (TTL: 15 minutes)
   Key: telemetry:latest:{device_id}
   Type: Sorted Set or Stream
   Purpose: Real-time dashboards, last-known-value queries
   
   Example (Sorted Set by timestamp):
   ZADD telemetry:latest:uuid-sensor-001 1705340000 "21.5|95"
   Score = Unix timestamp, Value = temperature|quality
   
   Query: Get latest reading for device
   ZRANGE telemetry:latest:uuid-sensor-001 -1 -1 WITHSCORES
   
   Benefits:
   - O(1) for latest reading
   - No database round trip
   - Perfect for live dashboards

4. BUILDING OCCUPANCY (TTL: 5 minutes, Real-time)
   Key: occupancy:{building_id}:{floor}
   Type: Hash
   Purpose: Real-time occupancy tracking
   
   Value:
   {
       "current_occupants": 234,
       "capacity": 500,
       "trend": "increasing",
       "last_update": "2025-01-15T14:32:45Z",
       "zones": {
           "floor-1": 45,
           "floor-2": 67,
           "floor-3": 122
       }
   }
   
   Updates:
   - When occupancy sensor fires
   - Push to Redis immediately
   - WebSocket subscribers notified
   - Database update batched every 5 minutes

5. ALERT STATUS (TTL: varies by severity)
   Key: alert:{alert_id}
   Type: Hash
   Purpose: Track active alerts
   
   TTL Strategy:
   - Info: 1 hour
   - Warning: 4 hours
   - Critical: 24 hours (require manual ack)
   
   Value:
   {
       "device_id": "uuid-123",
       "type": "temperature_high",
       "value": 45.2,
       "threshold": 30,
       "created_at": "2025-01-15T14:32:00Z",
       "acknowledged": false,
       "acknowledged_by": null
   }

6. LOOKUP/REFERENCE DATA (TTL: 12 hours)
   Key: lookup:building_zones:{building_id}
   Type: Set or Sorted Set
   Purpose: Cache hierarchical lookups
   
   Example: Get all devices in a building
   SMEMBERS lookup:building_zones:uuid-building-456
   Returns: ["zone-1", "zone-2", "zone-3", ...]
   
   Then: Get devices in zone
   SMEMBERS lookup:zone_devices:zone-1
   Returns: ["device-1", "device-2", ...]

7. RATE LIMITING (TTL: 1 second to 1 minute)
   Key: ratelimit:{user_id}:{api_endpoint}
   Type: String (counter)
   Purpose: API throttling
   
   Example:
   INCR ratelimit:user-123:api/telemetry
   EXPIRE ratelimit:user-123:api/telemetry 60
   
   Limits:
   - 100 requests/minute for standard users
   - 1000 requests/minute for admin
   - Sliding window implementation

8. GEO-SPATIAL QUERIES (TTL: 6 hours)
   Key: geo:building:{building_id}
   Type: Geospatial Set (using GEOADD)
   Purpose: Find devices near location
   
   Example:
   GEOADD geo:building:uuid-456 -74.0060 40.7128 "device-1"
   GEOADD geo:building:uuid-456 -74.0065 40.7135 "device-2"
   
   Query: Find devices within 100 meters
   GEORADIUS geo:building:uuid-456 -74.0060 40.7128 100 m
   Returns: ["device-1", "device-2"]
```

### Cache Invalidation Strategy:

```
Invalidation Patterns:

1. TTL-based (Automatic)
   - Session data: 1 hour
   - Device config: 6 hours
   - Reference data: 12 hours
   - Perfect for: Read-mostly data

2. Event-driven (Publish-Subscribe)
   When device config changes in MongoDB:
   
   MongoDB → Change Stream Listener
   ↓
   PUBLISH cache:invalidate {"type": "device_config", "device_id": "uuid-123"}
   ↓
   All services subscribe to "cache:invalidate"
   ↓
   DEL device:config:uuid-123
   ↓
   Next read triggers lazy reload from MongoDB

3. Pattern-based (Wildcard)
   Invalidate all related data:
   DEL lookup:building_zones:uuid-456   # Delete building lookup
   DEL lookup:zone_devices:zone-*       # Delete all zones in building
   DEL telemetry:latest:device-*        # Delete all device telemetry

4. Proactive refresh (Before TTL)
   For critical data:
   PEXPIRE session:{session_id} 3600000 NX  # Set 1-hour TTL if not exists
   If key will expire soon:
   Check expiration time (TTL command)
   Proactively refresh in background thread
   Reset TTL

Key Eviction Policy:
Redis maxmemory-policy: allkeys-lru
- When memory full: Remove least recently used keys
- Protects critical keys (session, alerts)
- Automatically evicts old telemetry readings
```

### Regional Redis Deployment:

```
US-EAST Redis Cluster:
- 3-node cluster (primary + 2 replicas)
- 256 GB total capacity
- Handles: Sessions, occupancy, active alerts
- Persistence: AOF (append-only file) + RDB snapshots
- Replication lag: < 1ms

US-WEST Redis Cluster:
- Independent cluster (no sync with US-EAST)
- Local sessions and cache
- Failover: Automatic with Sentinel

EU Redis Cluster:
- GDPR: Separate cluster (EU data only)
- Encryption: TLS + Redis ACL
- No sync with non-EU clusters
```

---

## 7. **MESSAGE QUEUE (Apache ActiveMQ Artemis) - Reliable Task Distribution**

### What Goes Here?
Guaranteed delivery tasks - notifications, reports, non-time-critical operations.

### Why Apache ActiveMQ Artemis?
- **Multi-Protocol Support**: AMQP 1.0, STOMP, OpenWire, MQTT, and HornetQ native
- **High Performance**: Non-blocking architecture with persistent journaling
- **JMS 2.0 Compliant**: Full Java Message Service specification support
- **Flexible Addressing**: Address model with anycast (point-to-point) and multicast (pub-sub)
- **Kubernetes Native**: Excellent integration with cloud-native deployments

### Artemis Core Concepts:

```
Address Model:

┌─────────────────────────────────────────────────────────────────┐
│                         ADDRESS                                  │
│  (logical endpoint where producers send messages)               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│   │   QUEUE 1   │    │   QUEUE 2   │    │   QUEUE 3   │        │
│   │  (anycast)  │    │ (multicast) │    │ (multicast) │        │
│   └─────────────┘    └─────────────┘    └─────────────┘        │
│                                                                  │
│   ANYCAST: Point-to-point (one consumer gets message)           │
│   MULTICAST: Pub-sub (all subscribed queues get copy)           │
└─────────────────────────────────────────────────────────────────┘
```

### SmartBuild Artemis Queues:

```
Address/Queue Architecture:

1. EMAIL NOTIFICATIONS
   Address: notifications.email
   Queue: notifications.email
   Routing Type: ANYCAST (competing consumers)
   Purpose: Send alert emails to facility managers
   
   Message Schema:
   {
       "type": "alert",
       "to": "manager@smartbuild.com",
       "subject": "CRITICAL: Temperature alert in Building A",
       "body": "Floor 3 temperature reached 45C",
       "priority": 9,
       "building_id": "uuid-456",
       "timestamp": "2025-01-15T14:32:00Z",
       "retry_count": 0,
       "max_retries": 3,
       "_AMQ_SCHED_DELIVERY": null
   }
   
   Characteristics:
   - Message Expiry: 86400000ms (24 hours)
   - Durability: DURABLE queue persisted to journal
   - Consumers: 3 email worker instances (load balanced)
   - Processing: Send email via SMTP
   - Retry: Redelivery policy with exponential backoff
   - Dead Letter: messages → DLA.notifications.email after 3 attempts
   - Max Delivery Attempts: 3

2. SMS ALERTS
   Address: notifications.sms
   Queue: notifications.sms
   Routing Type: ANYCAST
   Purpose: Send critical SMS to on-call staff
   
   Message:
   {
       "type": "critical_alert",
       "phone": "+1-555-0123",
       "message": "CRITICAL: Building A loss of power detected",
       "priority": 9,
       "timestamp": "2025-01-15T14:32:00Z"
   }
   
   Characteristics:
   - Message Expiry: 300000ms (5 minutes - time-sensitive)
   - Consumer Window Size: 1 (ensure one SMS at a time)
   - Consumers: 2 SMS gateway workers
   - Rate limit: 100 SMS/minute per carrier (application-level)

3. REPORT GENERATION
   Address: reports.generation
   Queue: reports.generation
   Routing Type: ANYCAST
   Purpose: Generate daily energy reports
   
   Message:
   {
       "report_type": "daily_energy_summary",
       "building_id": "uuid-456",
       "date": "2025-01-15",
       "recipient": "admin@smartbuild.com",
       "priority": 1,
       "scheduled_time": "2025-01-16T06:00:00Z"
   }
   
   Characteristics:
   - Message Expiry: -1 (no expiry - can wait indefinitely)
   - Priority: Enabled (process urgent buildings first)
   - Consumers: 5 report generator workers
   - Processing time: 2-5 minutes per report
   - Output: Store in Azure Blob + email link to user

4. DEVICE FIRMWARE UPDATES
   Address: device.firmware
   Queue: device.firmware
   Routing Type: ANYCAST
   Purpose: Distribute firmware updates to devices
   
   Message:
   {
       "type": "firmware_update",
       "device_id": "uuid-sensor-001",
       "building_id": "uuid-456",
       "firmware_version": "2.5.0",
       "download_url": "https://cdn.smartbuild.com/fw/2.5.0.bin",
       "checksum": "sha256:abc123...",
       "priority": 5,
       "scheduled_window": {
           "start": "2025-01-16T02:00:00Z",
           "end": "2025-01-16T04:00:00Z"
       }
   }
   
   Characteristics:
   - Message Expiry: 604800000ms (7 days retry window)
   - Batch processing: Slow consumer rate (10 devices/minute)
   - Consumers: 2 device update coordinators
   - Max Delivery Attempts: 5 over 7 days

5. BROADCAST ALERTS (Multicast Example)
   Address: alerts.broadcast
   Queues: 
     - alerts.broadcast.email (multicast)
     - alerts.broadcast.sms (multicast)
     - alerts.broadcast.webhook (multicast)
   Routing Type: MULTICAST
   Purpose: Fan-out critical alerts to all channels
   
   Message:
   {
       "type": "critical_system_alert",
       "severity": "critical",
       "building_id": "uuid-456",
       "message": "Fire alarm triggered in Building A",
       "timestamp": "2025-01-15T14:32:00Z"
   }
   
   Characteristics:
   - All subscribed queues receive a copy
   - Each channel processes independently
   - Used for emergency notifications

6. MAINTENANCE TASK NOTIFICATIONS
   Address: maintenance.tasks
   Queue: maintenance.tasks
   Routing Type: ANYCAST
   Purpose: Notify technicians of assigned tasks
   
   Message:
   {
       "type": "task_assigned",
       "task_id": "uuid-task-789",
       "technician_id": "tech-john-smith",
       "building_id": "uuid-456",
       "device_id": "uuid-hvac-001",
       "task_type": "preventive_maintenance",
       "urgency": "medium",
       "estimated_duration": 120,
       "parts_kit": "HX-500-KIT"
   }
   
   Characteristics:
   - Message Expiry: 2592000000ms (30 days)
   - Priority: Enabled (high priority first)
   - Consumers: 1 (mobile app push service)
   - Delivery: Push notification + app badge
```

### Address/Queue Design Patterns:

```
Point-to-Point (Anycast):
- Address: tasks.device_config
- Queue: tasks.device_config
- Routing: ANYCAST
- Pattern: Competing consumers - one consumer gets each message
- Processing: 1 message = 1 operation

Publish-Subscribe (Multicast):
- Address: alerts.system
- Queues: alerts.email, alerts.sms, alerts.webhook (all multicast)
- Pattern: All subscribed queues get a copy of each message
- Use case: Fan-out notifications to multiple channels

Filtered Subscription (Selector):
- Address: maintenance.all
- Queue: maintenance.critical
- Filter: "priority >= 8"
- Pattern: Only receive messages matching filter criteria
- Use case: Priority-based routing

Priority Queue:
- Address: reports.generation
- Queue: reports.generation (with priority enabled)
- Priority Range: 0-9 (9 = highest)
- Processing: Higher priority messages delivered first
- Configuration: default-max-priority="9" in address settings

Scheduled Delivery:
- Property: _AMQ_SCHED_DELIVERY (epoch timestamp)
- Use case: Schedule firmware updates for maintenance windows
- Example: Send at 2AM local time
```

### Reliability Guarantees:

```
Message Durability:

1. Producer Acknowledgments
   Producer → Artemis Broker ← Acknowledgment
   
   AMQP: Sender settle mode = UNSETTLED (wait for disposition)
   JMS: Session.CLIENT_ACKNOWLEDGE or DUPS_OK_ACKNOWLEDGE
   Every message confirmed persisted to journal before ACK

2. Consumer Acknowledgments
   Consumer receives message
   Process message
   If success: ACK → message removed from queue
   If failure: No ACK → message redelivered after timeout
   
   Consumer Window Size: 10 (flow control)
   Acknowledgment Batch: 1 (for critical queues)
   Redelivery Delay: Exponential (1s, 5s, 30s, 2m)

3. Dead Letter Address (DLA)
   Message exceeds max-delivery-attempts → DLA
   Purpose: Manual inspection and handling
   Retention: 30 days (configurable per address)
   Alert: Monitor DLA queue depth via Prometheus
   
   Configuration:
   <address-setting match="notifications.#">
       <dead-letter-address>DLA.notifications</dead-letter-address>
       <max-delivery-attempts>3</max-delivery-attempts>
       <redelivery-delay>5000</redelivery-delay>
       <redelivery-multiplier>2.0</redelivery-multiplier>
   </address-setting>

4. Message Persistence
   Journal: Append-only log with fsync
   Storage: AIO (Linux) or NIO (cross-platform)
   Paging: Large queues page to disk automatically
   Backup: Journal files replicated to standby broker
```

### Clustering & High Availability:

```
Artemis HA Architecture:

┌─────────────────────────────────────────────────────────────────┐
│                      LIVE-BACKUP PAIR                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌─────────────┐         ┌─────────────┐                       │
│   │    LIVE     │ ──────► │   BACKUP    │                       │
│   │   BROKER    │  sync   │   BROKER    │                       │
│   │  (active)   │         │ (standby)   │                       │
│   └─────────────┘         └─────────────┘                       │
│         │                       │                                │
│         ▼                       ▼                                │
│   ┌─────────────────────────────────────┐                       │
│   │        SHARED JOURNAL DATA          │                       │
│   │    (or replicated journal sync)     │                       │
│   └─────────────────────────────────────┘                       │
└─────────────────────────────────────────────────────────────────┘

HA Policies:
- Replication: Journal replicated to backup (network-based)
- Shared Store: NFS or Azure Files shared journal
- Failover: Automatic when live broker fails
- Failback: Configurable (live can reclaim after recovery)
```

### Regional Queue Topology:

```
US-EAST Artemis Cluster:
- Live-Backup pair (2 brokers)
- Addresses: High-priority notifications (email, SMS)
- Throughput: 50K msgs/second
- Journal: SSD-backed with replication
- Protocol: AMQP 1.0 (primary), STOMP (webhooks)

US-WEST Artemis Cluster:
- Live-Backup pair
- Addresses: Reports, firmware updates
- Bridge: Message bridge to US-EAST for critical alerts
- Configuration:
  <bridge name="us-west-to-us-east">
      <queue-name>alerts.critical</queue-name>
      <forwarding-address>alerts.critical</forwarding-address>
      <static-connectors>
          <connector-ref>us-east-connector</connector-ref>
      </static-connectors>
  </bridge>

EU Artemis Cluster:
- Separate isolated cluster (GDPR requirement)
- GDPR compliance: No message bridging to non-EU clusters
- Encryption: TLS 1.3 + message-level encryption
- Journal encryption: Enabled for data-at-rest
- Audit logging: All message operations logged

Kubernetes Deployment:
- ArtemisCloud Operator for automated management
- StatefulSet with persistent volumes
- Service mesh integration (Istio) for mTLS
- Prometheus metrics via JMX exporter
```

### Protocol Access:

```
Multi-Protocol Acceptors:

AMQP 1.0 (Primary - for Go/Python services):
- Port: 5672 (plain), 5671 (TLS)
- Use case: High-performance messaging
- Libraries: go-amqp, python-qpid-proton

STOMP (For web/webhook integrations):
- Port: 61613 (plain), 61614 (TLS)
- Use case: Simple text-based protocol, webhooks
- Libraries: stomp.py, stomp.js

MQTT (For IoT devices):
- Port: 1883 (plain), 8883 (TLS)
- Use case: Lightweight device messaging
- Note: Can bridge MQTT → Artemis queues

Core Protocol (Internal):
- Port: 61616
- Use case: Artemis-to-Artemis bridges, clustering
```

---

## 8. **SECRETS MANAGEMENT (Azure Key Vault)**

### What Goes Here?
Sensitive configuration data - API keys, connection strings, certificates, encryption keys, and credentials that must never be stored in code or config files.

### SmartBuild Key Vault Structure:

```
Vault Organization:

SmartBuild Key Vault Hierarchy:
├── smartbuild-kv-dev          (Development environment)
├── smartbuild-kv-staging      (Staging/QA environment)
├── smartbuild-kv-prod-useast  (Production US-EAST)
├── smartbuild-kv-prod-uswest  (Production US-WEST)
└── smartbuild-kv-prod-eu      (Production EU - GDPR isolated)

Secret Naming Convention:
{service}--{category}--{name}
Examples:
- api-gateway--jwt--secret
- device-service--mongodb--connection-string
- notification-service--twilio--api-key
- analytics-ingestion--scylladb--password
```

### SmartBuild Examples:

```yaml
# 1. DATABASE CONNECTION STRINGS
Secrets:
  device-service--mongodb--connection-string:
    value: "mongodb+srv://smartbuild:${password}@cluster0.mongodb.net/devices?retryWrites=true"
    content_type: "connection-string"
    tags:
      service: device-service
      database: mongodb
      environment: production
    expiration: null  # No expiration for connection strings

  user-service--azuresql--connection-string:
    value: "Server=tcp:smartbuild.database.windows.net;Database=users;User ID=app_user;Password=${password};Encrypt=true;"
    content_type: "connection-string"
    tags:
      service: user-service
      database: azure-sql
      region: us-east

  analytics-ingestion--scylladb--hosts:
    value: "scylla-node1.smartbuild.internal,scylla-node2.smartbuild.internal,scylla-node3.smartbuild.internal"
    content_type: "text/plain"
    tags:
      service: analytics-ingestion
      database: scylladb

# 2. API KEYS & TOKENS
Secrets:
  api-gateway--jwt--secret:
    value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  # 256-bit key
    content_type: "jwt-secret"
    tags:
      service: api-gateway
      purpose: authentication
    rotation_policy:
      automatic_rotation: true
      rotation_interval_days: 90
      notify_before_expiry_days: 30

  notification-service--twilio--api-key:
    value: "SK1234567890abcdef..."
    content_type: "api-key"
    tags:
      service: notification-service
      provider: twilio
      purpose: sms-alerts
    expiration: "2025-12-31T23:59:59Z"

  notification-service--sendgrid--api-key:
    value: "SG.xxxxxxxxxxxx..."
    content_type: "api-key"
    tags:
      service: notification-service
      provider: sendgrid
      purpose: email-delivery

  agentic-ai--google--gemini-api-key:
    value: "AIzaSy..."
    content_type: "api-key"
    tags:
      service: agentic-ai
      provider: google
      model: gemini-2.0-flash

# 3. CERTIFICATES
Certificates:
  smartbuild--tls--wildcard:
    type: "certificate"
    subject: "CN=*.smartbuild.com"
    issuer: "DigiCert"
    validity_months: 12
    auto_renew: true
    key_type: "RSA"
    key_size: 2048
    tags:
      purpose: tls-termination
      domains: "*.smartbuild.com"

  mqtt-adapter--client--cert:
    type: "certificate"
    subject: "CN=mqtt-adapter.smartbuild.internal"
    purpose: "MQTT broker mutual TLS"
    key_type: "EC"
    key_curve: "P-256"
    tags:
      service: mqtt-adapter
      protocol: mqtt
      auth: mtls

# 4. ENCRYPTION KEYS
Keys:
  smartbuild--data--encryption-key:
    type: "RSA"
    key_size: 4096
    operations: ["encrypt", "decrypt", "wrapKey", "unwrapKey"]
    tags:
      purpose: data-at-rest-encryption
      compliance: gdpr

  analytics--pii--encryption-key:
    type: "RSA"
    key_size: 2048
    operations: ["encrypt", "decrypt"]
    tags:
      purpose: pii-field-encryption
      services: "anomaly-detection,business-analytics"

# 5. INFRASTRUCTURE SECRETS
Secrets:
  kafka--cluster--sasl-password:
    value: "${complex_password}"
    content_type: "password"
    tags:
      service: kafka
      auth: sasl-scram

  redis--cluster--auth-token:
    value: "${redis_auth_token}"
    content_type: "auth-token"
    tags:
      service: redis
      purpose: cluster-auth

  prometheus--basic--auth-password:
    value: "${prometheus_password}"
    content_type: "password"
    tags:
      service: prometheus
      purpose: scrape-auth
```

### Access Patterns & SDK Usage:

```go
// Go Service - Key Vault Integration
package config

import (
    "context"
    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    "github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

type KeyVaultClient struct {
    client   *azsecrets.Client
    vaultURL string
    cache    map[string]*CachedSecret
}

// CachedSecret with TTL for performance
type CachedSecret struct {
    Value     string
    ExpiresAt time.Time
}

func NewKeyVaultClient(vaultURL string) (*KeyVaultClient, error) {
    // Use Managed Identity in production (no credentials in code)
    cred, err := azidentity.NewDefaultAzureCredential(nil)
    if err != nil {
        return nil, fmt.Errorf("failed to get Azure credential: %w", err)
    }

    client, err := azsecrets.NewClient(vaultURL, cred, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create Key Vault client: %w", err)
    }

    return &KeyVaultClient{
        client:   client,
        vaultURL: vaultURL,
        cache:    make(map[string]*CachedSecret),
    }, nil
}

// GetSecret retrieves a secret with caching (5-minute TTL)
func (kv *KeyVaultClient) GetSecret(ctx context.Context, name string) (string, error) {
    // Check cache first
    if cached, ok := kv.cache[name]; ok {
        if time.Now().Before(cached.ExpiresAt) {
            return cached.Value, nil
        }
    }

    // Fetch from Key Vault
    resp, err := kv.client.GetSecret(ctx, name, "", nil)
    if err != nil {
        return "", fmt.Errorf("failed to get secret %s: %w", name, err)
    }

    // Cache for 5 minutes
    kv.cache[name] = &CachedSecret{
        Value:     *resp.Value,
        ExpiresAt: time.Now().Add(5 * time.Minute),
    }

    return *resp.Value, nil
}

// Usage in service initialization
func LoadConfig() (*ServiceConfig, error) {
    kvClient, err := NewKeyVaultClient(os.Getenv("AZURE_KEYVAULT_URL"))
    if err != nil {
        return nil, err
    }

    ctx := context.Background()

    // Load secrets at startup
    mongoConn, _ := kvClient.GetSecret(ctx, "device-service--mongodb--connection-string")
    jwtSecret, _ := kvClient.GetSecret(ctx, "api-gateway--jwt--secret")
    kafkaPassword, _ := kvClient.GetSecret(ctx, "kafka--cluster--sasl-password")

    return &ServiceConfig{
        MongoDBConnectionString: mongoConn,
        JWTSecret:              jwtSecret,
        KafkaSASLPassword:      kafkaPassword,
    }, nil
}
```

```python
# Python Service - Key Vault Integration
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient
from functools import lru_cache
from datetime import datetime, timedelta
import os

class KeyVaultConfig:
    """
    Key Vault client with caching for Python services
    """
    
    def __init__(self, vault_url: str = None):
        self.vault_url = vault_url or os.getenv("AZURE_KEYVAULT_URL")
        
        # Use Managed Identity (workload identity in K8s)
        credential = DefaultAzureCredential()
        self.client = SecretClient(vault_url=self.vault_url, credential=credential)
        
        self._cache = {}
        self._cache_ttl = timedelta(minutes=5)
    
    def get_secret(self, name: str) -> str:
        """Get secret with 5-minute cache TTL"""
        cache_key = name
        
        # Check cache
        if cache_key in self._cache:
            value, expires_at = self._cache[cache_key]
            if datetime.now() < expires_at:
                return value
        
        # Fetch from Key Vault
        secret = self.client.get_secret(name)
        
        # Cache result
        self._cache[cache_key] = (
            secret.value,
            datetime.now() + self._cache_ttl
        )
        
        return secret.value
    
    def get_connection_string(self, service: str, database: str) -> str:
        """Convenience method for database connections"""
        secret_name = f"{service}--{database}--connection-string"
        return self.get_secret(secret_name)
    
    def get_api_key(self, service: str, provider: str) -> str:
        """Convenience method for API keys"""
        secret_name = f"{service}--{provider}--api-key"
        return self.get_secret(secret_name)

# Usage in anomaly-detection service
kv = KeyVaultConfig()

config = ServiceConfig(
    scylladb_password=kv.get_secret("analytics-ingestion--scylladb--password"),
    kafka_sasl_password=kv.get_secret("kafka--cluster--sasl-password"),
    redis_auth_token=kv.get_secret("redis--cluster--auth-token"),
    gemini_api_key=kv.get_api_key("agentic-ai", "google"),
)
```

### Kubernetes Integration:

```yaml
# Service Account with Workload Identity
apiVersion: v1
kind: ServiceAccount
metadata:
  name: device-service
  namespace: smartbuild
  annotations:
    azure.workload.identity/client-id: "12345678-1234-1234-1234-123456789abc"
---
# Pod with Key Vault CSI Driver (secrets as files)
apiVersion: v1
kind: Pod
metadata:
  name: device-service
  labels:
    azure.workload.identity/use: "true"
spec:
  serviceAccountName: device-service
  containers:
    - name: device-service
      image: smartbuild/device-service:latest
      env:
        - name: AZURE_KEYVAULT_URL
          value: "https://smartbuild-kv-prod-useast.vault.azure.net/"
      volumeMounts:
        - name: secrets-store
          mountPath: "/mnt/secrets"
          readOnly: true
  volumes:
    - name: secrets-store
      csi:
        driver: secrets-store.csi.k8s.io
        readOnly: true
        volumeAttributes:
          secretProviderClass: "smartbuild-keyvault"
---
# SecretProviderClass for CSI Driver
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: smartbuild-keyvault
  namespace: smartbuild
spec:
  provider: azure
  parameters:
    usePodIdentity: "false"
    useVMManagedIdentity: "false"
    clientID: "12345678-1234-1234-1234-123456789abc"
    keyvaultName: "smartbuild-kv-prod-useast"
    tenantId: "87654321-4321-4321-4321-cba987654321"
    objects: |
      array:
        - |
          objectName: device-service--mongodb--connection-string
          objectType: secret
        - |
          objectName: api-gateway--jwt--secret
          objectType: secret
        - |
          objectName: kafka--cluster--sasl-password
          objectType: secret
```

### Key Characteristics:
- ✅ **Zero secrets in code** - All sensitive data externalized
- ✅ **RBAC access control** - Fine-grained permissions per service
- ✅ **Audit logging** - Every secret access logged for compliance
- ✅ **Automatic rotation** - Keys and secrets rotate without downtime
- ✅ **Managed Identity** - No credentials to manage (workload identity)
- ✅ **Encryption at rest** - HSM-backed key storage
- ✅ **Versioning** - Secret history maintained for rollback

### Regional Considerations:

```
Key Vault Regional Strategy:

US-EAST (Primary Production):
- Vault: smartbuild-kv-prod-useast
- Contains: All US service secrets
- Redundancy: Zone-redundant storage
- Access: US services only (network rules)
- Backup: Daily to Azure Storage (encrypted)

US-WEST (Secondary Production):
- Vault: smartbuild-kv-prod-uswest
- Contains: Replicated critical secrets
- Sync: Manual sync for critical rotation
- Purpose: Disaster recovery failover
- Access: US-WEST services only

EU (GDPR Isolated):
- Vault: smartbuild-kv-prod-eu
- Contains: EU-specific secrets ONLY
- Location: West Europe (Amsterdam)
- Compliance: No replication outside EU
- Access: EU services only (private endpoint)
- Audit: 2-year retention for GDPR

Secret Categories by Vault:

Common Secrets (all regions):
- JWT signing keys (rotated quarterly)
- API Gateway secrets
- Kafka SASL credentials

Region-Specific:
- Database connection strings (region-local)
- Redis auth tokens (region-local)
- Third-party API keys (may vary by region)

EU-Only (GDPR):
- EU user encryption keys
- EU-specific API credentials
- GDPR audit encryption key
```

### Access Control Matrix:

```
RBAC Permissions per Service:

Service              | Secrets Access                           | Keys Access
---------------------|------------------------------------------|------------------
api-gateway          | jwt-secret, rate-limit-config            | None
device-service       | mongodb-conn, redis-auth                 | None
user-service         | azuresql-conn, jwt-secret                | pii-encryption
notification-service | twilio-key, sendgrid-key, redis-auth     | None
analytics-ingestion  | scylladb-pass, kafka-sasl, redis-auth    | None
anomaly-detection    | scylladb-pass, kafka-sasl, model-server  | pii-encryption
agentic-ai           | gemini-key, mongodb-conn                 | None
event-processor      | kafka-sasl, notification-service-url     | None

Access Policies:
- GET only: Services can read, never write
- No LIST: Services cannot enumerate secrets
- Audit: All access logged to Log Analytics
- Network: Private endpoint access only (no public)
```

### Secret Rotation Strategy:

```
Rotation Policies:

1. Database Passwords (90 days)
   - Azure SQL: Automatic rotation via Key Vault
   - MongoDB: Manual rotation with dual-credential pattern
   - ScyllaDB: Manual rotation during maintenance window
   
   Dual-Credential Pattern:
   Day 0: password-v1 active, password-v2 created
   Day 1: Both passwords valid (transition period)
   Day 2: Services switch to password-v2
   Day 3: password-v1 disabled
   Day 4: password-v1 deleted

2. JWT Signing Keys (90 days)
   - Asymmetric rotation
   - Old key valid for verification (30 days overlap)
   - New key used for signing immediately
   
3. API Keys (varies by provider)
   - Twilio: 180 days
   - SendGrid: 365 days
   - Gemini: No expiration, rotate on security events

4. Certificates (Auto-renew 30 days before expiry)
   - TLS certificates: Auto-renewed via Key Vault
   - mTLS client certs: Auto-renewed, services reload

5. Encryption Keys (Annually)
   - Data encryption keys: Rotate with re-encryption
   - Key wrapping: New version, old versions decrypt-only
```

### Disaster Recovery:

```
Key Vault Backup Strategy:

Daily Backup:
- All secrets exported (encrypted)
- Stored in geo-redundant Azure Storage
- Retention: 90 days
- Tested monthly (restore to test vault)

Recovery Scenarios:

1. Accidental Secret Deletion:
   - Soft delete enabled (90-day retention)
   - Recovery: az keyvault secret recover --name <secret>
   - RTO: < 1 minute

2. Vault Corruption/Loss:
   - Restore from daily backup
   - Re-create vault in same region
   - Update service configurations
   - RTO: 30 minutes
   - RPO: 24 hours (last backup)

3. Regional Outage:
   - Failover to secondary vault
   - Services configured with failover URL
   - DNS switch to secondary region
   - RTO: 5 minutes (if pre-configured)

Backup Command:
az keyvault secret backup --vault-name smartbuild-kv-prod-useast \
    --name device-service--mongodb--connection-string \
    --file ./backups/mongodb-conn.bak
```

---

## 9. **DECISION MATRIX: Choosing the Right Database**

```
Data Type          → SQL        → Document       → Time-Series  → Events      → Cache        → Queue        → Secrets
                   (SQL Server) (MongoDB Atlas)  (ScyllaDB)      (ScyllaDB)    (Redis)        (Artemis)      (Key Vault)

User/Auth Data     ✅✅✅                                                    ✅             
Device Registry    ✅✅        ✅✅
Device Config                   ✅✅✅        
Sensor Readings                              ✅✅✅
Telemetry Streams                            ✅✅✅           ✅            ✅ (latest)
Event Audit Log                                              ✅✅✅         
Access Logs                                                 ✅✅✅
Alerts                                                      ✅             ✅✅
Sessions                                                                   ✅✅✅         
Rate Limiting                                                             ✅✅✅
Real-time Data                                                           ✅✅✅
Notifications                                                                           ✅✅✅
Reports                                                                               ✅✅
Maintenance                                   ✅
API Keys                                                                                             ✅✅✅
Connection Strings                                                                                   ✅✅✅
Certificates                                                                                         ✅✅✅
Encryption Keys                                                                                      ✅✅✅
JWT Secrets                                                                                          ✅✅✅
```

---

## 10. **SCHEMA DESIGN BEST PRACTICES**

### Partitioning Strategy by Region:

```
Data Distribution Rules:

1. Building-First Partitioning
   Every table/collection partitioned by building_id
   Reason: Query locality + easier disaster recovery
   
   Example:
   SELECT * FROM device_telemetry WHERE building_id = 'NYC' AND time > NOW() - 1h;
   → Uses partition for NYC, no cross-region scan

2. Time-Based Partitioning (Time-Series)
   Partition by day/week/month
   Recent data in fast storage
   Historical data compressed/archived
   
   Example:
   device_telemetry_2025_01_15 (hot, SSD)
   device_telemetry_2025_01_10 (warm, compressed)
   device_telemetry_2024_q4 (cold, S3 archive)

3. Geographic Replication
   Each region has complete copy
   Local queries use local region database
   No cross-region latency for reads
   Async replication for updates
```

---

## 11. **DISASTER RECOVERY & BACKUP STRATEGY**

```
RTO/RPO by Data Type:

Azure SQL MI (User/Config Data):
- RTO: 5 minutes (auto-failover groups)
- RPO: 5 seconds (synchronous replication)
- Backup: Automated backups + point-in-time restore (35 days)
- Replication: Auto-failover group to secondary region

TimescaleDB (Time-Series):
- RTO: 30 minutes
- RPO: 1 hour
- Backup: Daily snapshots (compressed)
- Older data: Accept loss if < 24 hours old

MongoDB Atlas (Configs):
- RTO: 10 minutes
- RPO: 5 minutes (point-in-time recovery)
- Backup: Continuous cloud backup + snapshots
- Replication: 3-node replica set across AZs
- Restore: Point-in-time to any second in oplog window

ScyllaDB (Events):
- RTO: 1 minute (quorum replication)
- RPO: < 1 second
- Backup: Incremental daily
- 7-year retention for compliance

Redis (Cache):
- RTO: 0 minutes (no data loss acceptable)
- RPO: N/A (recreate on demand)
- Strategy: Ephemeral, not critical

Artemis (Tasks):
- RTO: 5 minutes
- RPO: 0 (durable journal on disk)
- Replication: Live-backup pairs
```

---

## Summary: Data Flow in SmartBuild

```
                              Azure Key Vault
                                    ↓
                    (secrets, keys, certificates)
                                    ↓
IoT Devices
    ↓
[MQTT/UDP/HTTP] ← Protocol Gateway
    ↓
Device Ingest (normalization) 
    ↓
┌─→ Kafka: device-telemetry ──→ TimescaleDB (time-series)
│                           ──→ Redis: latest readings
│                           ──→ Stream Processors
├─→ Kafka: device-events ──→ ScyllaDB (immutable audit)
│                       ──→ Rules Engine
├─→ Azure SQL MI ← Device Registry, Users, Config (relational)
├─→ MongoDB Atlas ← Device Profiles, Building Config (hierarchical)
├─→ Redis ← Sessions, Cache, Live Data
└─→ Artemis ← Notifications (reliable delivery)

Dashboard/API Query:
Service startup → Key Vault: Load secrets (connection strings, API keys)
User logged in → Session: Redis (fast)
Device config needed → Cache: Redis (2KB, < 1ms)
                    → MongoDB (if cache miss)
Last reading → Redis telemetry:latest (< 1ms)
Historical data → ScyllaDB (pre-aggregated)
Audit trail → ScyllaDB (immutable records)
```

---

## Conclusion

By understanding your data characteristics, you can make informed decisions:

1. **RELATIONAL** (SQL Server / Azure SQL MI): When you need ACID transactions, relationships, and complex queries
2. **HIERARCHICAL/DOCUMENT** (MongoDB Atlas): When structure is flexible, nested, and semi-structured
3. **TIME-SERIES** (ScyllaDB): When time is the primary dimension and you need high-throughput writes (3B reads/day)
4. **EVENTS** (ScyllaDB): When you need immutable, replicated audit trails with strong consistency
5. **REAL-TIME** (Redis): When you need sub-millisecond access for sessions, caching, and live data
6. **STREAMING** (Kafka): When you need decoupled, high-throughput producers/consumers with event sourcing
7. **RELIABILITY** (Apache ActiveMQ Artemis): When guaranteed delivery matters more than latency (email, SMS, reports)
8. **SECRETS** (Azure Key Vault): When you need secure storage for API keys, connection strings, certificates, and encryption keys with audit logging and automatic rotation

The polyglot approach lets you use the **right tool for each job** rather than forcing all data into one database. For SmartBuild's 50,000 sensors, this architecture handles billions of telemetry points daily while maintaining ACID compliance for critical user/building data.

