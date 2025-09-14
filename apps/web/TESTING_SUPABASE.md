# Supabase Testing Guide

This document describes the testing setup for Supabase integration in the
AgentFlow web application.

## Test Structure

### Unit Tests

Located in `src/lib/__tests__/`:

- `supabase.test.ts` - Tests for client-side Supabase configuration
- `supabase-admin.test.ts` - Tests for server-side Supabase admin configuration

### Integration Tests

Located in `src/lib/__tests__/`:

- `supabase-integration.test.ts` - End-to-end tests that require a running
  Supabase instance

### Test Utilities

Located in `src/lib/test-utils.ts`:

- Mock factories for creating test data
- Helper functions for setting up test environments
- Shared utilities for Supabase testing

## Running Tests

### Unit Tests Only

```bash
npm run test:unit
```

This runs all tests except integration tests. These tests use mocks and don't
require a live database.

### Integration Tests

```bash
npm run test:integration
```

This runs only the integration tests. **Requires a running Supabase instance.**

### All Tests

```bash
npm test
```

Runs all tests including unit and integration tests.

### Test Coverage

```bash
npm run test:coverage
```

Runs all tests and generates a coverage report.

## Integration Test Setup

Integration tests require environment variables to be set:

### Required Environment Variables

```bash
RUN_INTEGRATION_TESTS=true
NEXT_PUBLIC_SUPABASE_URL=http://127.0.0.1:54321
NEXT_PUBLIC_SUPABASE_ANON_KEY=your_anon_key
SUPABASE_URL=http://127.0.0.1:54321
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
```

### Using Local Supabase

1. Start local Supabase instance:

   ```bash
   npx supabase start
   ```

2. Run the integration test script:
   ```bash
   ./scripts/test-integration-example.sh
   ```

## Test Coverage

### Current Coverage (Supabase modules)

- **supabase.ts**: 100% coverage
- **supabase-admin.ts**: 100% coverage

### What's Tested

#### Client Configuration (`supabase.test.ts`)

- Environment variable validation
- Client creation with correct configuration
- Error handling for missing environment variables

#### Admin Configuration (`supabase-admin.test.ts`)

- Admin client creation with service role
- getSupabaseClient function with/without access tokens
- Session management for authenticated requests

#### Integration Tests (`supabase-integration.test.ts`)

- Live database connections
- CRUD operations on all main tables
- Database schema validation
- Error handling for invalid queries
- Authentication state management

## Test Data Factories

The `test-utils.ts` file provides factories for creating mock data:

```typescript
import {
  createMockAgentRun,
  createMockProject,
  createMockTask,
} from '@/lib/test-utils';

// Create mock project
const project = createMockProject({ name: 'Custom Project' });

// Create mock task
const task = createMockTask({ status: 'in_progress' });

// Create mock agent run
const agentRun = createMockAgentRun({ agent: 'sa' });
```

## Best Practices

### Unit Tests

1. Always mock external dependencies
2. Test both success and error scenarios
3. Validate input parameters and return values
4. Use descriptive test names

### Integration Tests

1. Clean up test data after each test
2. Use unique identifiers to avoid conflicts
3. Test realistic scenarios
4. Verify database constraints and relationships

### Test Utilities

1. Create reusable mock factories
2. Provide sensible defaults
3. Allow easy customization through overrides
4. Keep utilities simple and focused

## Troubleshooting

### Common Issues

#### "Missing environment variable" errors

- Ensure all required environment variables are set
- Check that .env file exists and is properly formatted

#### Integration tests failing

- Verify Supabase instance is running (`npx supabase status`)
- Check database migrations are applied
- Ensure test user has proper permissions

#### Jest configuration warnings

- Jest should use `moduleNameMapping` for path aliases
- Ensure test files are in `__tests__` directories
- Exclude utility files from test runs

### Debug Tips

1. Use `console.log` in tests to debug issues
2. Run single test files: `npx jest supabase.test.ts`
3. Use `--verbose` flag for detailed output
4. Check Supabase logs for database errors

## CI/CD Considerations

For continuous integration:

1. **Unit Tests**: Should run on every commit
2. **Integration Tests**: Can be run separately or with test database
3. **Coverage**: Enforce minimum thresholds
4. **Environment**: Use test-specific environment variables

Example GitHub Actions workflow:

```yaml
- name: Run Unit Tests
  run: npm run test:unit

- name: Run Integration Tests
  run: npm run test:integration
  env:
    RUN_INTEGRATION_TESTS: true
    NEXT_PUBLIC_SUPABASE_URL: ${{ secrets.TEST_SUPABASE_URL }}
    # ... other environment variables
```
