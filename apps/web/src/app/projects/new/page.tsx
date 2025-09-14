"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useProjectStore } from "@/stores";

export default function NewProjectPage() {
  const router = useRouter();
  const { addProject } = useProjectStore();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [status, setStatus] = useState<"draft" | "active">("draft");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    const now = new Date().toISOString();

    // Try server API (Supabase admin) first
    try {
      const res = await fetch("/api/projects", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, description, status }),
      });

      if (res.ok) {
        const { project } = await res.json();
        const id = project?.id ?? `project-${Date.now()}`;
        // Mirror into local store for UI
        addProject({
          id,
          name,
          description,
          status: status,
          createdAt: now,
          updatedAt: now,
        });
        router.push(`/projects/${id}`);
        return;
      }
      const data = await res.json().catch(() => ({}));
      throw new Error(data?.error || `Failed to create project (${res.status})`);
    } catch (err: any) {
      // Fallback to local store only
      const id = `project-${Date.now()}`;
      addProject({
        id,
        name,
        description,
        status: status,
        createdAt: now,
        updatedAt: now,
      });
      // Surface non-blocking warning for persistence failure
      setError(err?.message || "Unable to persist to server, created locally.");
      router.push(`/projects/${id}`);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
      <div className="container mx-auto px-4 py-8">
        <Card className="max-w-2xl mx-auto">
          <CardHeader>
            <CardTitle>New Project</CardTitle>
            <CardDescription>Kick off a new AI Software House project</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Name</label>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Customer Onboarding Revamp"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Description</label>
                <Textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Short summary of goals and scope"
                  rows={4}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Status</label>
                <Select value={status} onValueChange={(v) => setStatus(v as any)}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="draft">Draft</SelectItem>
                    <SelectItem value="active">Active</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {error && (
                <p className="text-sm text-amber-600">{error}</p>
              )}

              <div className="flex items-center gap-2">
                <Button type="submit" disabled={submitting}>
                  {submitting ? "Creating..." : "Create Project"}
                </Button>
                <Link href="/" className="text-blue-600 hover:underline">
                  Cancel
                </Link>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

