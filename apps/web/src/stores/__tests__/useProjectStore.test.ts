import { renderHook, act } from '@testing-library/react'
import { useProjectStore } from '../useProjectStore'

describe('useProjectStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useProjectStore.getState().reset()
  })

  it('should initialize with empty state', () => {
    const { result } = renderHook(() => useProjectStore())
    
    expect(result.current.projects).toEqual([])
    expect(result.current.currentProject).toBeNull()
    expect(result.current.isLoading).toBe(false)
  })

  it('should add a new project', () => {
    const { result } = renderHook(() => useProjectStore())
    
    const newProject = {
      id: '1',
      name: 'Test Project',
      description: 'A test project',
      status: 'active' as const,
      createdAt: new Date().toISOString(),
    }

    act(() => {
      result.current.addProject(newProject)
    })

    expect(result.current.projects).toHaveLength(1)
    expect(result.current.projects[0]).toEqual(newProject)
  })

  it('should set current project', () => {
    const { result } = renderHook(() => useProjectStore())
    
    const project = {
      id: '1',
      name: 'Test Project',
      description: 'A test project',
      status: 'active' as const,
      createdAt: new Date().toISOString(),
    }

    act(() => {
      result.current.addProject(project)
      result.current.setCurrentProject(project)
    })

    expect(result.current.currentProject).toEqual(project)
  })

  it('should update project', () => {
    const { result } = renderHook(() => useProjectStore())
    
    const project = {
      id: '1',
      name: 'Test Project',
      description: 'A test project',
      status: 'active' as const,
      createdAt: new Date().toISOString(),
    }

    act(() => {
      result.current.addProject(project)
    })

    act(() => {
      result.current.updateProject('1', { name: 'Updated Project' })
    })

    expect(result.current.projects[0].name).toBe('Updated Project')
  })

  it('should delete project', () => {
    const { result } = renderHook(() => useProjectStore())
    
    const project = {
      id: '1',
      name: 'Test Project',
      description: 'A test project',
      status: 'active' as const,
      createdAt: new Date().toISOString(),
    }

    act(() => {
      result.current.addProject(project)
      result.current.setCurrentProject(project)
    })

    act(() => {
      result.current.deleteProject('1')
    })

    expect(result.current.projects).toHaveLength(0)
    expect(result.current.currentProject).toBeNull()
  })

  it('should set loading state', () => {
    const { result } = renderHook(() => useProjectStore())

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
