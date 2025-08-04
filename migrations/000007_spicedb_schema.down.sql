-- Migration: Drop SpiceDB schema
-- This migration drops the SpiceDB schema tables

-- Connect to the spicedb database
\c spicedb;

-- Drop tables in reverse order
DROP TABLE IF EXISTS relation_tuple CASCADE;
DROP TABLE IF EXISTS namespace_config CASCADE;
