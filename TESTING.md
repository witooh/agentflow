# Testing Guide

## Overview

This project uses Jest with React Testing Library for comprehensive unit
testing. All tests must pass before any commit is allowed.

## Test Framework

- **Jest**: JavaScript testing framework
- **React Testing Library**: React component testing utilities
- **@testing-library/jest-dom**: Custom Jest matchers for DOM elements
- **@testing-library/user-event**: User interaction simulation

## Running Tests

### Run all tests

```bash
pnpm test
```

### Run tests in watch mode

```bash
pnpm test:watch
```

### Run tests with coverage

```bash
pnpm test:coverage
```

### Run tests for specific workspace

```bash
cd apps/web && pnpm test
```

## Test Structure

```
src/
├── components/
│   └── ui/
│       └── __tests__/
│           └── button.test.tsx
├── lib/
│   └── __tests__/
│       └── utils.test.ts
└── stores/
    └── __tests__/
        ├── useProjectStore.test.ts
        └── useTaskStore.test.ts
```

## Test Coverage

The project enforces minimum test coverage thresholds:

- **Branches**: 70%
- **Functions**: 70%
- **Lines**: 70%
- **Statements**: 70%

## Pre-commit Hooks

Every commit automatically runs:

1. **TypeScript type checking** (`pnpm typecheck`)
2. **Unit tests** (`pnpm test --passWithNoTests`)
3. **Linting** (`pnpm lint --fix`)
4. **Code formatting** (`prettier --write`)

If any of these checks fail, the commit will be blocked.

## Writing Tests

### Component Tests

```tsx
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from '../button';

describe('Button', () => {
  it('should render with default props', () => {
    render(<Button>Click me</Button>);

    const button = screen.getByRole('button', { name: /click me/i });
    expect(button).toBeInTheDocument();
  });
});
```

### Store Tests

```tsx
import { act, renderHook } from '@testing-library/react';
import { useProjectStore } from '../useProjectStore';

describe('useProjectStore', () => {
  beforeEach(() => {
    useProjectStore.getState().reset();
  });

  it('should add a new project', () => {
    const { result } = renderHook(() => useProjectStore());

    const newProject = {
      id: '1',
      name: 'Test Project',
      // ... other properties
    };

    act(() => {
      result.current.addProject(newProject);
    });

    expect(result.current.projects).toHaveLength(1);
  });
});
```

### Utility Tests

```tsx
import { cn } from '../utils';

describe('utils', () => {
  describe('cn', () => {
    it('should merge class names correctly', () => {
      expect(cn('class1', 'class2')).toBe('class1 class2');
    });
  });
});
```

## Test Configuration

### Jest Configuration (`jest.config.js`)

- Uses Next.js Jest configuration
- JSdom environment for React testing
- Module name mapping for `@/` imports
- Coverage collection from `src/**/*.{js,jsx,ts,tsx}`
- Excludes `.d.ts` and `.stories` files

### Jest Setup (`jest.setup.js`)

- Imports `@testing-library/jest-dom` matchers
- Mocks Next.js router and image components
- Clears all mocks before each test

## Best Practices

1. **Test Behavior, Not Implementation**: Focus on what the component does, not
   how it does it
2. **Use Semantic Queries**: Prefer `getByRole`, `getByLabelText` over
   `getByTestId`
3. **Test User Interactions**: Use `userEvent` for realistic user interactions
4. **Reset State**: Always reset store state in `beforeEach` for store tests
5. **Mock External Dependencies**: Mock API calls, external libraries, and
   Next.js features
6. **Write Descriptive Test Names**: Use clear, descriptive test descriptions
7. **Keep Tests Simple**: One assertion per test when possible
8. **Test Edge Cases**: Include tests for error states, empty states, and edge
   cases

## Debugging Tests

### Run specific test file

```bash
pnpm test button.test.tsx
```

### Run tests matching pattern

```bash
pnpm test --testNamePattern="should render"
```

### Debug mode

```bash
pnpm test --detectOpenHandles --forceExit
```

## Continuous Integration

The pre-commit hooks ensure that:

- All TypeScript code compiles without errors
- All tests pass
- Code follows linting rules
- Code is properly formatted

This guarantees code quality and prevents broken code from being committed to
the repository.

# Test commit
