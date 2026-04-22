-- =============================================================================
-- Payment Gateway - PostgreSQL Initialization
-- =============================================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Schema
CREATE SCHEMA IF NOT EXISTS payment;

-- Set search path
ALTER DATABASE payment_gateway SET search_path TO payment, public;