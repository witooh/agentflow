// Core cross-agent contracts (TS-only, no runtime deps)

export type JobEnvelope<T> = {
  type: string;
  version: '1.0';
  projectId: string;
  taskId: string;
  payload: T;
  replyTo?: string;
  requestedBy: 'agent:pm' | 'agent:sa' | 'agent:fe' | 'agent:be' | 'agent:qa' | 'agent:devops' | 'human';
};

export type Question = { key: string; ask: string };
export type Questions = { questions: Question[] };

export type Story = { id: string; title: string; ac: string[] };
export type Stories = { stories: Story[] };

export type TaskItem = {
  id: string;
  title: string;
  role: 'pm' | 'sa' | 'architect' | 'techlead' | 'fe' | 'be' | 'qa' | 'devops' | 'writer';
  estimateHour: number; // 1..80
  dependsOn?: string[];
};
export type Tasks = { tasks: TaskItem[] };

