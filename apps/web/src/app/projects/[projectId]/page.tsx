"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useProjectStore, useTaskStore, type Task } from "@/stores";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

type PageProps = { params: { projectId: string } };

const STATUSES: Task["status"][] = [
  "todo",
  "in_progress",
  "review",
  "done",
  "blocked",
];

export default function ProjectDetailPage({ params }: PageProps) {
  const { projectId } = params;
  const { projects, setCurrentProject } = useProjectStore();
  const project = useMemo(() => projects.find((p) => p.id === projectId) || null, [projects, projectId]);

  // Keep current project in store for other components if needed
  if (project) setCurrentProject(project);

  if (!project) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <div className="container mx-auto px-4 py-8">
          <Card className="max-w-3xl mx-auto">
            <CardHeader>
              <CardTitle>Project not found</CardTitle>
              <CardDescription>
                The requested project does not exist in the local store.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Link href="/" className="text-blue-600 hover:underline">
                ← Back to Home
              </Link>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      <div className="container mx-auto px-4 py-8">
        <header className="max-w-6xl mx-auto mb-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h1 className="text-3xl font-bold text-slate-900 dark:text-slate-100">{project.name}</h1>
              {project.description && (
                <p className="text-slate-600 dark:text-slate-400 mt-1">{project.description}</p>
              )}
              <div className="mt-2 flex items-center gap-2">
                <Badge variant="outline">{project.status}</Badge>
                <Badge variant="secondary">Project ID: {project.id}</Badge>
              </div>
            </div>
            <div>
              <Link href="/" className="text-blue-600 hover:underline">← Back</Link>
            </div>
          </div>
        </header>

        <Card className="max-w-6xl mx-auto">
          <CardHeader>
            <CardTitle>Project Workspace</CardTitle>
            <CardDescription>Chat, documentation, tasks, and activity</CardDescription>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="tasks" className="w-full">
              <TabsList>
                <TabsTrigger value="chat">Chat</TabsTrigger>
                <TabsTrigger value="srs">SRS</TabsTrigger>
                <TabsTrigger value="tasks">Tasks Kanban</TabsTrigger>
                <TabsTrigger value="agents">Agents</TabsTrigger>
                <TabsTrigger value="artifacts">Artifacts</TabsTrigger>
                <TabsTrigger value="activity">Activity</TabsTrigger>
              </TabsList>

              <TabsContent value="chat">
                <ProjectChat projectId={project.id} />
              </TabsContent>

              <TabsContent value="srs">
                <SRSPlaceholder />
              </TabsContent>

              <TabsContent value="tasks">
                <KanbanBoard projectId={project.id} />
              </TabsContent>

              <TabsContent value="agents">
                <AgentsPlaceholder />
              </TabsContent>

              <TabsContent value="artifacts">
                <ArtifactsPlaceholder />
              </TabsContent>

              <TabsContent value="activity">
                <ActivityPlaceholder />
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function ProjectChat({ projectId }: { projectId: string }) {
  const [messages, setMessages] = useState<{ id: string; role: "user" | "agent"; text: string; at: string }[]>([]);
  const [text, setText] = useState("");

  const send = () => {
    if (!text.trim()) return;
    const now = new Date().toISOString();
    setMessages((prev) => [...prev, { id: `m-${Date.now()}`, role: "user", text, at: now }]);
    setText("");
  };

  return (
    <div className="space-y-3">
      <div className="h-60 overflow-auto rounded-md border p-3 bg-slate-50 dark:bg-slate-900/40">
        {messages.length === 0 ? (
          <p className="text-sm text-slate-500">No messages yet for {projectId}. Start the conversation below.</p>) : (
          <ul className="space-y-2">
            {messages.map((m) => (
              <li key={m.id} className="text-sm">
                <span className="font-medium">{m.role === "user" ? "You" : "Agent"}:</span> {m.text}
                <span className="text-xs text-slate-500 ml-2">{new Date(m.at).toLocaleTimeString()}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
      <div className="flex gap-2">
        <Input
          placeholder="Type a message..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => { if (e.key === "Enter") send(); }}
        />
        <Button onClick={send}>Send</Button>
      </div>
    </div>
  );
}

function SRSPlaceholder() {
  return (
    <div className="space-y-2">
      <p className="text-sm text-slate-600 dark:text-slate-400">
        SRS preview goes here. Link artifacts to Supabase storage later.
      </p>
      <Badge variant="outline">T-WEB-04 Skeleton</Badge>
    </div>
  );
}

function AgentsPlaceholder() {
  return (
    <div className="space-y-2">
      <p className="text-sm text-slate-600 dark:text-slate-400">
        Agent runs and statuses will appear here.
      </p>
      <Badge variant="outline">agent_runs table (read-only)</Badge>
    </div>
  );
}

function ArtifactsPlaceholder() {
  return (
    <div className="space-y-2">
      <p className="text-sm text-slate-600 dark:text-slate-400">
        Artifacts list (SRS.md, ARCHITECTURE.md, test reports) will show here.
      </p>
      <Badge variant="outline">artifacts/&lt;projectId&gt;/...</Badge>
    </div>
  );
}

function ActivityPlaceholder() {
  return (
    <div className="space-y-2">
      <p className="text-sm text-slate-600 dark:text-slate-400">
        Activity feed of tasks, agents, and messages.
      </p>
      <Badge variant="outline">events: project.updated, task.moved</Badge>
    </div>
  );
}

function KanbanBoard({ projectId }: { projectId: string }) {
  const { tasks, addTask, updateTask } = useTaskStore();

  const byStatus = useMemo(() => {
    const groups: Record<Task["status"], Task[]> = {
      todo: [],
      in_progress: [],
      review: [],
      done: [],
      blocked: [],
    };
    for (const t of tasks) {
      if (t.projectId === projectId) groups[t.status].push(t);
    }
    return groups;
  }, [tasks, projectId]);

  const createSampleTask = () => {
    const now = new Date().toISOString();
    const newTask: Task = {
      id: `task-${Date.now()}`,
      title: "New Task",
      description: "Sample task for Kanban",
      status: "todo",
      priority: "medium",
      projectId,
      estimateHour: 2,
      dependsOn: [],
      createdAt: now,
      updatedAt: now,
    };
    addTask(newTask);
  };

  const cycleStatus = (task: Task) => {
    const order = STATUSES;
    const idx = order.indexOf(task.status);
    const next = order[(idx + 1) % order.length];
    updateTask(task.id, { status: next, updatedAt: new Date().toISOString() });
  };

  return (
    <div className="space-y-3">
      <div className="flex justify-between items-center">
        <p className="text-sm text-slate-600 dark:text-slate-400">Click a card to advance status</p>
        <Button onClick={createSampleTask}>Add Task</Button>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 xl:grid-cols-5 gap-3">
        {STATUSES.map((status) => (
          <div key={status} className="rounded-lg border bg-slate-50/50 dark:bg-slate-900/30">
            <div className="flex items-center justify-between p-2 border-b">
              <h3 className="text-sm font-medium capitalize">{status.replace("_", " ")}</h3>
              <Badge variant="secondary">{byStatus[status].length}</Badge>
            </div>
            <div className="p-2 space-y-2 min-h-24">
              {byStatus[status].length === 0 ? (
                <p className="text-xs text-slate-500">No tasks</p>
              ) : (
                byStatus[status].map((t) => (
                  <button
                    key={t.id}
                    onClick={() => cycleStatus(t)}
                    className="w-full text-left p-2 rounded-md border bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 transition"
                    title="Click to move to next status"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-sm">{t.title}</span>
                      <Badge variant="outline" className="text-[10px]">{t.priority}</Badge>
                    </div>
                    {t.description && (
                      <p className="text-xs text-slate-500 mt-1 line-clamp-2">{t.description}</p>
                    )}
                  </button>
                ))
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

