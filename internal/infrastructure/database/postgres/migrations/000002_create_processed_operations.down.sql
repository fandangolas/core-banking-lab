-- Migration: Drop processed_operations table
-- Version: 000002
-- Description: Rollback migration for processed_operations table

DROP TABLE IF EXISTS processed_operations;
