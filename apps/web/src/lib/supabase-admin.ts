import { createClient } from "@supabase/supabase-js";
import type { Database } from "./supabase";

const supabaseUrl = process.env.SUPABASE_URL!;
const supabaseServiceRoleKey = process.env.SUPABASE_SERVICE_ROLE_KEY!;

if (!supabaseUrl) {
    throw new Error("Missing SUPABASE_URL environment variable");
}

if (!supabaseServiceRoleKey) {
    throw new Error("Missing SUPABASE_SERVICE_ROLE_KEY environment variable");
}

// Server-side Supabase admin client with service role key
// This bypasses RLS and should only be used on the server
export const supabaseAdmin = createClient<Database>(
    supabaseUrl,
    supabaseServiceRoleKey,
    {
        auth: {
            autoRefreshToken: false,
            persistSession: false,
        },
    },
);

// Helper function to get the regular client for server-side use
// This respects RLS and should be used when you want to act as a user
export const getSupabaseClient = (accessToken?: string) => {
    const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!;

    const client = createClient<Database>(supabaseUrl, supabaseAnonKey, {
        auth: {
            autoRefreshToken: false,
            persistSession: false,
        },
    });

    if (accessToken) {
        // Set the session with the access token
        client.auth.setSession({
            access_token: accessToken,
            refresh_token: "",
        });
    }

    return client;
};
