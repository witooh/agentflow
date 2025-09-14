-- Enable pgvector (extension name: vector)
-- This migration is idempotent and safe to run multiple times.
create extension if not exists vector;

