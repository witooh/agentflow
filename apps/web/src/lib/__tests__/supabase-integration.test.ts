/**
 * Integration tests for Supabase connection
 * These tests require a running Supabase instance (local or remote)
 * Run with: npm run test:integration
 */

// NOTE: Avoid importing Supabase clients at module load time.
// Import lazily inside the integration describe so that the suite can
// be skipped cleanly when RUN_INTEGRATION_TESTS !== "true".
let supabase: any;
let supabaseAdmin: any;

// Skip integration tests if not explicitly enabled
const shouldRunIntegrationTests = process.env.RUN_INTEGRATION_TESTS === "true";

const describeIntegration = shouldRunIntegrationTests
    ? describe
    : describe.skip;

describeIntegration("Supabase Integration Tests", () => {
    beforeAll(async () => {
        // Verify environment variables are set
        if (!process.env.NEXT_PUBLIC_SUPABASE_URL) {
            throw new Error(
                "NEXT_PUBLIC_SUPABASE_URL is required for integration tests",
            );
        }
        if (!process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY) {
            throw new Error(
                "NEXT_PUBLIC_SUPABASE_ANON_KEY is required for integration tests",
            );
        }
        if (!process.env.SUPABASE_URL) {
            throw new Error("SUPABASE_URL is required for integration tests");
        }
        if (!process.env.SUPABASE_SERVICE_ROLE_KEY) {
            throw new Error(
                "SUPABASE_SERVICE_ROLE_KEY is required for integration tests",
            );
        }

        // Lazy import only after envs are validated
        const clientMod = await import("../supabase");
        supabase = clientMod.supabase;
        const adminMod = await import("../supabase-admin");
        supabaseAdmin = adminMod.supabaseAdmin;
    });

    describe("Client Connection", () => {
        it("should connect to Supabase successfully", async () => {
            const { data, error } = await supabase
                .from("projects")
                .select("*")
                .limit(1);

            expect(error).toBeNull();
            expect(data).toBeDefined();
            expect(Array.isArray(data)).toBe(true);
        });

        it("should handle authentication state", async () => {
            const { data: { user } } = await supabase.auth.getUser();

            // Should not have a user by default (anon key)
            expect(user).toBeNull();
        });

        it("should be able to query projects table", async () => {
            const { data, error } = await supabase
                .from("projects")
                .select("id, name, status")
                .limit(5);

            expect(error).toBeNull();
            expect(data).toBeDefined();
            expect(Array.isArray(data)).toBe(true);
        });

        it("should be able to query tasks table", async () => {
            const { data, error } = await supabase
                .from("tasks")
                .select("id, title, status")
                .limit(5);

            expect(error).toBeNull();
            expect(data).toBeDefined();
            expect(Array.isArray(data)).toBe(true);
        });

        it("should be able to query agent_runs table", async () => {
            const { data, error } = await supabase
                .from("agent_runs")
                .select("id, agent, status")
                .limit(5);

            expect(error).toBeNull();
            expect(data).toBeDefined();
            expect(Array.isArray(data)).toBe(true);
        });
    });

    describe("Admin Connection", () => {
        it("should connect with admin privileges", async () => {
            const { data, error } = await supabaseAdmin
                .from("projects")
                .select("*")
                .limit(1);

            expect(error).toBeNull();
            expect(data).toBeDefined();
            expect(Array.isArray(data)).toBe(true);
        });

        it("should be able to insert data with admin client", async () => {
            const testProject = {
                name: "Test Project",
                description: "Integration test project",
                status: "draft" as const,
                created_by: "test-user-id",
            };

            const { data, error } = await supabaseAdmin
                .from("projects")
                .insert(testProject)
                .select()
                .single();

            expect(error).toBeNull();
            expect(data).toBeDefined();
            expect(data.name).toBe(testProject.name);
            expect(data.status).toBe(testProject.status);

            // Clean up - delete the test project
            if (data?.id) {
                await supabaseAdmin
                    .from("projects")
                    .delete()
                    .eq("id", data.id);
            }
        });

        it("should be able to update data with admin client", async () => {
            // First create a test project
            const testProject = {
                name: "Test Update Project",
                description: "Project for update test",
                status: "draft" as const,
                created_by: "test-user-id",
            };

            const { data: insertData, error: insertError } = await supabaseAdmin
                .from("projects")
                .insert(testProject)
                .select()
                .single();

            expect(insertError).toBeNull();
            expect(insertData).toBeDefined();

            // Update the project
            const updateData = {
                name: "Updated Test Project",
                status: "active" as const,
            };

            const { data: updateResult, error: updateError } =
                await supabaseAdmin
                    .from("projects")
                    .update(updateData)
                    .eq("id", insertData.id)
                    .select()
                    .single();

            expect(updateError).toBeNull();
            expect(updateResult).toBeDefined();
            expect(updateResult.name).toBe(updateData.name);
            expect(updateResult.status).toBe(updateData.status);

            // Clean up - delete the test project
            if (insertData?.id) {
                await supabaseAdmin
                    .from("projects")
                    .delete()
                    .eq("id", insertData.id);
            }
        });
    });

    describe("Database Schema Validation", () => {
        it("should have correct projects table structure", async () => {
            const { data, error } = await supabase
                .from("projects")
                .select("*")
                .limit(1);

            expect(error).toBeNull();

            if (data && data.length > 0) {
                const project = data[0];
                expect(project).toHaveProperty("id");
                expect(project).toHaveProperty("name");
                expect(project).toHaveProperty("description");
                expect(project).toHaveProperty("status");
                expect(project).toHaveProperty("created_at");
                expect(project).toHaveProperty("updated_at");
                expect(project).toHaveProperty("created_by");

                // Check status enum values
                expect(["draft", "active", "completed", "cancelled"]).toContain(
                    project.status,
                );
            }
        });

        it("should have correct tasks table structure", async () => {
            const { data, error } = await supabase
                .from("tasks")
                .select("*")
                .limit(1);

            expect(error).toBeNull();

            if (data && data.length > 0) {
                const task = data[0];
                expect(task).toHaveProperty("id");
                expect(task).toHaveProperty("project_id");
                expect(task).toHaveProperty("title");
                expect(task).toHaveProperty("description");
                expect(task).toHaveProperty("status");
                expect(task).toHaveProperty("role");
                expect(task).toHaveProperty("estimate_hour");
                expect(task).toHaveProperty("created_at");
                expect(task).toHaveProperty("updated_at");
                expect(task).toHaveProperty("created_by");

                // Check status enum values
                expect(["todo", "in_progress", "review", "done", "blocked"])
                    .toContain(task.status);

                // Check role enum values
                expect([
                    "pm",
                    "sa",
                    "architect",
                    "techlead",
                    "fe",
                    "be",
                    "qa",
                    "devops",
                    "writer",
                ]).toContain(task.role);
            }
        });

        it("should have correct agent_runs table structure", async () => {
            const { data, error } = await supabase
                .from("agent_runs")
                .select("*")
                .limit(1);

            expect(error).toBeNull();

            if (data && data.length > 0) {
                const agentRun = data[0];
                expect(agentRun).toHaveProperty("id");
                expect(agentRun).toHaveProperty("project_id");
                expect(agentRun).toHaveProperty("agent");
                expect(agentRun).toHaveProperty("input");
                expect(agentRun).toHaveProperty("output");
                expect(agentRun).toHaveProperty("status");
                expect(agentRun).toHaveProperty("created_at");
                expect(agentRun).toHaveProperty("updated_at");
                expect(agentRun).toHaveProperty("completed_at");

                // Check status enum values
                expect(["running", "completed", "failed", "blocked"]).toContain(
                    agentRun.status,
                );
            }
        });
    });

    describe("Error Handling", () => {
        it("should handle invalid table queries gracefully", async () => {
            const { data, error } = await supabase
                .from("nonexistent_table")
                .select("*");

            expect(error).toBeDefined();
            expect(error?.message).toContain(
                'relation "nonexistent_table" does not exist',
            );
            expect(data).toBeNull();
        });

        it("should handle invalid column queries gracefully", async () => {
            const { data, error } = await supabase
                .from("projects")
                .select("nonexistent_column");

            expect(error).toBeDefined();
            expect(error?.message).toContain(
                'column "nonexistent_column" does not exist',
            );
            expect(data).toBeNull();
        });
    });
});
