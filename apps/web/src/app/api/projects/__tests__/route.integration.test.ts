/**
 * Integration tests for POST /api/projects route
 * Requires a running Supabase instance and valid env vars.
 * Run with: npm run test:integration
 */

let POST: (req: Request) => Promise<Response>;
let supabaseAdmin: any;

const shouldRun = process.env.RUN_INTEGRATION_TESTS === "true";
const describeIntegration = shouldRun ? describe : describe.skip;

function makeRequest(body: any) {
  // Minimal stub that satisfies route's usage (req.json())
  return { json: async () => body } as any as Request;
}

describeIntegration("POST /api/projects (integration)", () => {
  beforeAll(async () => {
    const required = [
      "NEXT_PUBLIC_SUPABASE_URL",
      "NEXT_PUBLIC_SUPABASE_ANON_KEY",
      "SUPABASE_URL",
      "SUPABASE_SERVICE_ROLE_KEY",
    ];
    for (const k of required) {
      if (!process.env[k]) throw new Error(`${k} missing for integration tests`);
    }
    const route = await import("@/app/api/projects/route");
    POST = route.POST;
    const admin = await import("@/lib/supabase-admin");
    supabaseAdmin = admin.supabaseAdmin;
  });

  it("creates a project in DB and returns it", async () => {
    const name = `itest-${Date.now()}`;
    const res = await POST(makeRequest({ name, status: "draft" }));
    expect(res.status).toBe(201);
    const data: any = await res.json();
    expect(data?.project?.id).toBeDefined();
    expect(data?.project?.name).toBe(name);

    // Cleanup: delete the created project
    const id = data.project.id;
    await supabaseAdmin.from("projects").delete().eq("id", id);
  });
});
