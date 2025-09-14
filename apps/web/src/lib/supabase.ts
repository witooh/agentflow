import { createClient } from "@supabase/supabase-js";

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!;
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!;

if (!supabaseUrl) {
    throw new Error("Missing NEXT_PUBLIC_SUPABASE_URL environment variable");
}

if (!supabaseAnonKey) {
    throw new Error(
        "Missing NEXT_PUBLIC_SUPABASE_ANON_KEY environment variable",
    );
}

// Client-side Supabase client
export const supabase = createClient(supabaseUrl, supabaseAnonKey, {
    auth: {
        autoRefreshToken: true,
        persistSession: true,
        detectSessionInUrl: true,
    },
});

// Database types (will be generated from Supabase later)
export type Database = {
    public: {
        Tables: {
            projects: {
                Row: {
                    id: string;
                    name: string;
                    description: string | null;
                    status: "draft" | "active" | "completed" | "cancelled";
                    created_at: string;
                    updated_at: string;
                    created_by: string;
                };
                Insert: {
                    id?: string;
                    name: string;
                    description?: string | null;
                    status?: "draft" | "active" | "completed" | "cancelled";
                    created_at?: string;
                    updated_at?: string;
                    created_by: string;
                };
                Update: {
                    id?: string;
                    name?: string;
                    description?: string | null;
                    status?: "draft" | "active" | "completed" | "cancelled";
                    created_at?: string;
                    updated_at?: string;
                    created_by?: string;
                };
            };
            tasks: {
                Row: {
                    id: string;
                    project_id: string;
                    title: string;
                    description: string | null;
                    status:
                        | "todo"
                        | "in_progress"
                        | "review"
                        | "done"
                        | "blocked";
                    role:
                        | "pm"
                        | "sa"
                        | "architect"
                        | "techlead"
                        | "fe"
                        | "be"
                        | "qa"
                        | "devops"
                        | "writer";
                    estimate_hour: number;
                    created_at: string;
                    updated_at: string;
                    created_by: string;
                };
                Insert: {
                    id?: string;
                    project_id: string;
                    title: string;
                    description?: string | null;
                    status?:
                        | "todo"
                        | "in_progress"
                        | "review"
                        | "done"
                        | "blocked";
                    role:
                        | "pm"
                        | "sa"
                        | "architect"
                        | "techlead"
                        | "fe"
                        | "be"
                        | "qa"
                        | "devops"
                        | "writer";
                    estimate_hour: number;
                    created_at?: string;
                    updated_at?: string;
                    created_by: string;
                };
                Update: {
                    id?: string;
                    project_id?: string;
                    title?: string;
                    description?: string | null;
                    status?:
                        | "todo"
                        | "in_progress"
                        | "review"
                        | "done"
                        | "blocked";
                    role?:
                        | "pm"
                        | "sa"
                        | "architect"
                        | "techlead"
                        | "fe"
                        | "be"
                        | "qa"
                        | "devops"
                        | "writer";
                    estimate_hour?: number;
                    created_at?: string;
                    updated_at?: string;
                    created_by?: string;
                };
            };
            agent_runs: {
                Row: {
                    id: string;
                    project_id: string;
                    agent: string;
                    input: Record<string, unknown>;
                    output: Record<string, unknown>;
                    status: "running" | "completed" | "failed" | "blocked";
                    created_at: string;
                    updated_at: string;
                    completed_at: string | null;
                };
                Insert: {
                    id?: string;
                    project_id: string;
                    agent: string;
                    input: Record<string, unknown>;
                    output?: Record<string, unknown>;
                    status?: "running" | "completed" | "failed" | "blocked";
                    created_at?: string;
                    updated_at?: string;
                    completed_at?: string | null;
                };
                Update: {
                    id?: string;
                    project_id?: string;
                    agent?: string;
                    input?: Record<string, unknown>;
                    output?: Record<string, unknown>;
                    status?: "running" | "completed" | "failed" | "blocked";
                    created_at?: string;
                    updated_at?: string;
                    completed_at?: string | null;
                };
            };
        };
        Views: {
            [_ in never]: never;
        };
        Functions: {
            [_ in never]: never;
        };
        Enums: {
            [_ in never]: never;
        };
    };
};
