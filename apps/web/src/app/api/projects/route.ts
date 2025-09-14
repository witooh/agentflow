import { NextResponse } from "next/server";
import { supabaseAdmin } from "@/lib/supabase-admin";

export async function POST(req: Request) {
  try {
    const body = await req.json();
    const { name, description, status } = body || {};

    if (!name || typeof name !== "string") {
      return NextResponse.json(
        { error: "name is required" },
        { status: 400 },
      );
    }

    const safeStatus = ["draft", "active", "completed", "cancelled"].includes(
      status,
    )
      ? status
      : "draft";

    // NOTE: Without auth, we tag created_by as a dev placeholder.
    // Replace with the authenticated user id once Auth is wired.
    const createdBy = "dev-local";

    const { data, error } = await (supabaseAdmin as any)
      .from("projects")
      .insert({
        name,
        description: description ?? null,
        status: safeStatus,
        created_by: createdBy,
      })
      .select("*")
      .single();

    if (error) {
      return NextResponse.json(
        { error: error.message },
        { status: 500 },
      );
    }

    return NextResponse.json({ project: data }, { status: 201 });
  } catch (e: any) {
    return NextResponse.json(
      { error: e?.message ?? "unknown error" },
      { status: 500 },
    );
  }
}
