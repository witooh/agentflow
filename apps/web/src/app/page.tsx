"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useProjectStore, useTaskStore } from "@/stores";

export default function Home() {
  const { projects, addProject } = useProjectStore();
  const { tasks, addTask } = useTaskStore();

  const handleAddSampleProject = () => {
    const newProject = {
      id: `project-${Date.now()}`,
      name: "Sample AI Software House Project",
      description: "A demonstration project for the AI Software House platform",
      status: "active" as const,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    addProject(newProject);
  };

  const handleAddSampleTask = () => {
    const newTask = {
      id: `task-${Date.now()}`,
      title: "Sample Task",
      description: "A demonstration task for the AI Software House platform",
      status: "todo" as const,
      priority: "medium" as const,
      projectId: projects[0]?.id || "default",
      estimateHour: 4,
      dependsOn: [],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    addTask(newTask);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      <div className="container mx-auto px-4 py-8">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-slate-900 dark:text-slate-100 mb-4">
            AI Software House
          </h1>
          <p className="text-xl text-slate-600 dark:text-slate-400 mb-8">
            TypeScript-first, Supabase-first multi-agent development platform
          </p>
          <Badge variant="outline" className="text-sm">
            T-WEB-01 Complete: Next.js + Tailwind + shadcn/ui + Zustand
          </Badge>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl mx-auto">
          {/* Projects Section */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                Projects
                <Badge variant="secondary">{projects.length}</Badge>
              </CardTitle>
              <CardDescription>
                Manage your AI software development projects
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {projects.length === 0 ? (
                <p className="text-slate-500 text-center py-4">
                  No projects yet. Create one to get started!
                </p>
              ) : (
                <div className="space-y-2">
                  {projects.map((project) => (
                    <div
                      key={project.id}
                      className="p-3 border rounded-lg bg-slate-50 dark:bg-slate-800"
                    >
                      <h3 className="font-medium">{project.name}</h3>
                      <p className="text-sm text-slate-600 dark:text-slate-400">
                        {project.description}
                      </p>
                      <Badge variant="outline" className="mt-2">
                        {project.status}
                      </Badge>
                    </div>
                  ))}
                </div>
              )}
              <Button onClick={handleAddSampleProject} className="w-full">
                Add Sample Project
              </Button>
            </CardContent>
          </Card>

          {/* Tasks Section */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                Tasks
                <Badge variant="secondary">{tasks.length}</Badge>
              </CardTitle>
              <CardDescription>
                Track development tasks and progress
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {tasks.length === 0 ? (
                <p className="text-slate-500 text-center py-4">
                  No tasks yet. Add one to get started!
                </p>
              ) : (
                <div className="space-y-2">
                  {tasks.map((task) => (
                    <div
                      key={task.id}
                      className="p-3 border rounded-lg bg-slate-50 dark:bg-slate-800"
                    >
                      <h3 className="font-medium">{task.title}</h3>
                      <p className="text-sm text-slate-600 dark:text-slate-400">
                        {task.description}
                      </p>
                      <div className="flex gap-2 mt-2">
                        <Badge variant="outline">{task.status}</Badge>
                        <Badge variant="outline">{task.priority}</Badge>
                        <Badge variant="outline">{task.estimateHour}h</Badge>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              <Button 
                onClick={handleAddSampleTask} 
                className="w-full"
                disabled={projects.length === 0}
              >
                Add Sample Task
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Tech Stack Section */}
        <Card className="mt-8 max-w-4xl mx-auto">
          <CardHeader>
            <CardTitle>Tech Stack</CardTitle>
            <CardDescription>
              Technologies configured in this Next.js scaffold
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="text-center p-4 border rounded-lg">
                <h3 className="font-medium">Next.js</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">App Router</p>
              </div>
              <div className="text-center p-4 border rounded-lg">
                <h3 className="font-medium">TypeScript</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">Type Safety</p>
              </div>
              <div className="text-center p-4 border rounded-lg">
                <h3 className="font-medium">Tailwind CSS</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">Styling</p>
              </div>
              <div className="text-center p-4 border rounded-lg">
                <h3 className="font-medium">shadcn/ui</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">Components</p>
              </div>
              <div className="text-center p-4 border rounded-lg">
                <h3 className="font-medium">Zustand</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">State Management</p>
              </div>
              <div className="text-center p-4 border rounded-lg">
                <h3 className="font-medium">ESLint</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">Code Quality</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
