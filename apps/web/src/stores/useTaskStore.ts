import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

export interface Task {
  id: string
  title: string
  description?: string
  status: 'todo' | 'in_progress' | 'review' | 'done' | 'blocked'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  assignee?: string
  projectId: string
  estimateHour: number
  dependsOn: string[]
  createdAt: string
  updatedAt: string
}

interface TaskState {
  tasks: Task[]
  currentTask: Task | null
  isLoading: boolean
  error: string | null
  filters: {
    status?: Task['status']
    priority?: Task['priority']
    assignee?: string
    projectId?: string
  }
}

interface TaskActions {
  setTasks: (tasks: Task[]) => void
  setCurrentTask: (task: Task | null) => void
  addTask: (task: Task) => void
  updateTask: (id: string, updates: Partial<Task>) => void
  deleteTask: (id: string) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
  setFilters: (filters: Partial<TaskState['filters']>) => void
  clearFilters: () => void
  getFilteredTasks: () => Task[]
  getTasksByStatus: (status: Task['status']) => Task[]
  reset: () => void
}

export const useTaskStore = create<TaskState & TaskActions>()(
  devtools(
    (set, get) => ({
      // Initial state
      tasks: [],
      currentTask: null,
      isLoading: false,
      error: null,
      filters: {},

      // Actions
      setTasks: (tasks) => set({ tasks }),
      
      setCurrentTask: (task) => set({ currentTask: task }),
      
      addTask: (task) => set((state) => ({
        tasks: [...state.tasks, task]
      })),
      
      updateTask: (id, updates) => set((state) => ({
        tasks: state.tasks.map(task =>
          task.id === id ? { ...task, ...updates } : task
        ),
        currentTask: state.currentTask?.id === id 
          ? { ...state.currentTask, ...updates }
          : state.currentTask
      })),
      
      deleteTask: (id) => set((state) => ({
        tasks: state.tasks.filter(task => task.id !== id),
        currentTask: state.currentTask?.id === id ? null : state.currentTask
      })),
      
      setLoading: (loading) => set({ isLoading: loading }),
      
      setError: (error) => set({ error }),
      
      clearError: () => set({ error: null }),
      
      setFilters: (filters) => set((state) => ({
        filters: { ...state.filters, ...filters }
      })),
      
      clearFilters: () => set({ filters: {} }),
      
      getFilteredTasks: () => {
        const { tasks, filters } = get()
        return tasks.filter(task => {
          if (filters.status && task.status !== filters.status) return false
          if (filters.priority && task.priority !== filters.priority) return false
          if (filters.assignee && task.assignee !== filters.assignee) return false
          if (filters.projectId && task.projectId !== filters.projectId) return false
          return true
        })
      },

      getTasksByStatus: (status) => {
        const { tasks } = get()
        return tasks.filter(task => task.status === status)
      },

      reset: () => set({
        tasks: [],
        currentTask: null,
        isLoading: false,
        error: null,
        filters: {}
      }),
    }),
    {
      name: 'task-store',
    }
  )
)
