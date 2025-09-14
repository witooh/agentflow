/**
 * Test utilities for Supabase testing
 */

import { SupabaseClient } from '@supabase/supabase-js'
import type { Database } from '../supabase'

export type MockSupabaseResponse<T> = {
  data: T | null
  error: any | null
}

export type MockSupabaseClient = {
  auth: {
    getUser: jest.Mock
    setSession: jest.Mock
    signIn: jest.Mock
    signOut: jest.Mock
  }
  from: jest.Mock
}

/**
 * Creates a mock Supabase client for testing
 */
export function createMockSupabaseClient(): MockSupabaseClient {
  const mockAuth = {
    getUser: jest.fn(),
    setSession: jest.fn(),
    signIn: jest.fn(),
    signOut: jest.fn(),
  }

  const mockFrom = jest.fn(() => ({
    select: jest.fn(() => ({
      limit: jest.fn(() => Promise.resolve({ data: [], error: null })),
      eq: jest.fn(() => ({
        single: jest.fn(() => Promise.resolve({ data: null, error: null })),
      })),
      single: jest.fn(() => Promise.resolve({ data: null, error: null })),
    })),
    insert: jest.fn(() => ({
      select: jest.fn(() => ({
        single: jest.fn(() => Promise.resolve({ data: null, error: null })),
      })),
    })),
    update: jest.fn(() => ({
      eq: jest.fn(() => ({
        select: jest.fn(() => ({
          single: jest.fn(() => Promise.resolve({ data: null, error: null })),
        })),
      })),
    })),
    delete: jest.fn(() => ({
      eq: jest.fn(() => Promise.resolve({ data: null, error: null })),
    })),
  }))

  return {
    auth: mockAuth,
    from: mockFrom,
  }
}

/**
 * Creates mock project data
 */
export function createMockProject(overrides: Partial<Database['public']['Tables']['projects']['Row']> = {}) {
  return {
    id: 'test-project-id',
    name: 'Test Project',
    description: 'Test project description',
    status: 'draft' as const,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    created_by: 'test-user-id',
    ...overrides,
  }
}

/**
 * Creates mock task data
 */
export function createMockTask(overrides: Partial<Database['public']['Tables']['tasks']['Row']> = {}) {
  return {
    id: 'test-task-id',
    project_id: 'test-project-id',
    title: 'Test Task',
    description: 'Test task description',
    status: 'todo' as const,
    role: 'fe' as const,
    estimate_hour: 4,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    created_by: 'test-user-id',
    ...overrides,
  }
}

/**
 * Creates mock agent run data
 */
export function createMockAgentRun(overrides: Partial<Database['public']['Tables']['agent_runs']['Row']> = {}) {
  return {
    id: 'test-agent-run-id',
    project_id: 'test-project-id',
    agent: 'intake',
    input: { question: 'What is the project about?' },
    output: { answer: 'This is a test project' },
    status: 'completed' as const,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    completed_at: '2024-01-01T00:05:00Z',
    ...overrides,
  }
}

/**
 * Creates a successful Supabase response
 */
export function createSuccessResponse<T>(data: T): MockSupabaseResponse<T> {
  return {
    data,
    error: null,
  }
}

/**
 * Creates an error Supabase response
 */
export function createErrorResponse(message: string, code?: string): MockSupabaseResponse<null> {
  return {
    data: null,
    error: {
      message,
      code,
      details: null,
      hint: null,
    },
  }
}

/**
 * Sets up environment variables for testing
 */
export function setupTestEnv() {
  process.env.NEXT_PUBLIC_SUPABASE_URL = 'https://test.supabase.co'
  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = 'test-anon-key'
  process.env.SUPABASE_URL = 'https://test.supabase.co'
  process.env.SUPABASE_SERVICE_ROLE_KEY = 'test-service-role-key'
}

/**
 * Clears environment variables after testing
 */
export function cleanupTestEnv() {
  delete process.env.NEXT_PUBLIC_SUPABASE_URL
  delete process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY
  delete process.env.SUPABASE_URL
  delete process.env.SUPABASE_SERVICE_ROLE_KEY
}
