import { renderHook, act } from '@testing-library/react'
import { useTaskStore } from '../useTaskStore'

describe('useTaskStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useTaskStore.getState().reset()
  })

  it('should initialize with empty state', () => {
    const { result } = renderHook(() => useTaskStore())
    
    expect(result.current.tasks).toEqual([])
    expect(result.current.currentTask).toBeNull()
    expect(result.current.isLoading).toBe(false)
  })

  it('should add a new task', () => {
    const { result } = renderHook(() => useTaskStore())
    
    const newTask = {
      id: '1',
      title: 'Test Task',
      description: 'A test task',
      status: 'todo' as const,
      priority: 'medium' as const,
      assignee: 'user1',
      projectId: 'project1',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    act(() => {
      result.current.addTask(newTask)
    })

    expect(result.current.tasks).toHaveLength(1)
    expect(result.current.tasks[0]).toEqual(newTask)
  })

  it('should set current task', () => {
    const { result } = renderHook(() => useTaskStore())
    
    const task = {
      id: '1',
      title: 'Test Task',
      description: 'A test task',
      status: 'todo' as const,
      priority: 'medium' as const,
      assignee: 'user1',
      projectId: 'project1',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    act(() => {
      result.current.addTask(task)
      result.current.setCurrentTask(task)
    })

    expect(result.current.currentTask).toEqual(task)
  })

  it('should update task', () => {
    const { result } = renderHook(() => useTaskStore())
    
    const task = {
      id: '1',
      title: 'Test Task',
      description: 'A test task',
      status: 'todo' as const,
      priority: 'medium' as const,
      assignee: 'user1',
      projectId: 'project1',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    act(() => {
      result.current.addTask(task)
    })

    act(() => {
      result.current.updateTask('1', { status: 'in_progress' })
    })

    expect(result.current.tasks[0].status).toBe('in_progress')
  })

  it('should delete task', () => {
    const { result } = renderHook(() => useTaskStore())
    
    const task = {
      id: '1',
      title: 'Test Task',
      description: 'A test task',
      status: 'todo' as const,
      priority: 'medium' as const,
      assignee: 'user1',
      projectId: 'project1',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    act(() => {
      result.current.addTask(task)
      result.current.setCurrentTask(task)
    })

    act(() => {
      result.current.deleteTask('1')
    })

    expect(result.current.tasks).toHaveLength(0)
    expect(result.current.currentTask).toBeNull()
  })

  it('should filter tasks by status', () => {
    const { result } = renderHook(() => useTaskStore())
    
    const tasks = [
      {
        id: '1',
        title: 'Task 1',
        description: 'A test task',
        status: 'todo' as const,
        priority: 'medium' as const,
        assignee: 'user1',
        projectId: 'project1',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      {
        id: '2',
        title: 'Task 2',
        description: 'Another test task',
        status: 'in_progress' as const,
        priority: 'high' as const,
        assignee: 'user2',
        projectId: 'project1',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ]

    act(() => {
      tasks.forEach(task => result.current.addTask(task))
    })

    const todoTasks = result.current.getTasksByStatus('todo')
    expect(todoTasks).toHaveLength(1)
    expect(todoTasks[0].id).toBe('1')

    const inProgressTasks = result.current.getTasksByStatus('in_progress')
    expect(inProgressTasks).toHaveLength(1)
    expect(inProgressTasks[0].id).toBe('2')
  })

  it('should set loading state', () => {
    const { result } = renderHook(() => useTaskStore())

    act(() => {
      result.current.setLoading(true)
    })

    expect(result.current.isLoading).toBe(true)

    act(() => {
      result.current.setLoading(false)
    })

    expect(result.current.isLoading).toBe(false)
  })
})
