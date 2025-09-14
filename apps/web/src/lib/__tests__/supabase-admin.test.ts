import { getSupabaseClient, supabaseAdmin } from "../supabase-admin";

// Mock the Supabase client
jest.mock("@supabase/supabase-js", () => ({
    createClient: jest.fn(() => ({
        auth: {
            autoRefreshToken: false,
            persistSession: false,
            setSession: jest.fn(),
        },
        from: jest.fn(() => ({
            select: jest.fn(() => ({
                limit: jest.fn(() =>
                    Promise.resolve({ data: [], error: null })
                ),
            })),
        })),
    })),
}));

describe("Supabase Admin Configuration", () => {
    beforeEach(() => {
        // Reset environment variables
        delete process.env.SUPABASE_URL;
        delete process.env.SUPABASE_SERVICE_ROLE_KEY;
        delete process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY;
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it("should throw error when SUPABASE_URL is missing", () => {
        // Mock environment variables
        process.env.SUPABASE_SERVICE_ROLE_KEY = "test-service-role-key";

        // Clear the module cache to force re-import
        jest.resetModules();

        expect(() => {
            require("../supabase-admin");
        }).toThrow("Missing SUPABASE_URL environment variable");
    });

    it("should throw error when SUPABASE_SERVICE_ROLE_KEY is missing", () => {
        // Mock environment variables
        process.env.SUPABASE_URL = "https://test.supabase.co";

        // Clear the module cache to force re-import
        jest.resetModules();

        expect(() => {
            require("../supabase-admin");
        }).toThrow("Missing SUPABASE_SERVICE_ROLE_KEY environment variable");
    });

    it("should create admin client with correct configuration when environment variables are set", () => {
        // Mock environment variables
        process.env.SUPABASE_URL = "https://test.supabase.co";
        process.env.SUPABASE_SERVICE_ROLE_KEY = "test-service-role-key";

        // Clear the module cache to force re-import
        jest.resetModules();

        const { createClient } = require("@supabase/supabase-js");
        const { supabaseAdmin: adminClient } = require("../supabase-admin");

        expect(createClient).toHaveBeenCalledWith(
            "https://test.supabase.co",
            "test-service-role-key",
            {
                auth: {
                    autoRefreshToken: false,
                    persistSession: false,
                },
            },
        );
        expect(adminClient).toBeDefined();
    });

    it("should export supabaseAdmin client", () => {
        // Mock environment variables
        process.env.SUPABASE_URL = "https://test.supabase.co";
        process.env.SUPABASE_SERVICE_ROLE_KEY = "test-service-role-key";

        // Clear the module cache to force re-import
        jest.resetModules();

        const { supabaseAdmin: adminClient } = require("../supabase-admin");
        expect(adminClient).toBeDefined();
        expect(adminClient.auth).toBeDefined();
    });

    describe("getSupabaseClient", () => {
        it("should create client with anon key when no access token provided", () => {
            // Mock environment variables
            process.env.SUPABASE_URL = "https://test.supabase.co";
            process.env.SUPABASE_SERVICE_ROLE_KEY = "test-service-role-key";
            process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = "test-anon-key";

            // Clear the module cache to force re-import
            jest.resetModules();

            const { createClient } = require("@supabase/supabase-js");
            const { getSupabaseClient } = require("../supabase-admin");

            const client = getSupabaseClient();

            expect(createClient).toHaveBeenCalledWith(
                "https://test.supabase.co",
                "test-anon-key",
                {
                    auth: {
                        autoRefreshToken: false,
                        persistSession: false,
                    },
                },
            );
            expect(client).toBeDefined();
        });

        it("should set session when access token is provided", () => {
            // Mock environment variables
            process.env.SUPABASE_URL = "https://test.supabase.co";
            process.env.SUPABASE_SERVICE_ROLE_KEY = "test-service-role-key";
            process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = "test-anon-key";

            // Clear the module cache to force re-import
            jest.resetModules();

            const { createClient } = require("@supabase/supabase-js");
            const { getSupabaseClient } = require("../supabase-admin");

            const mockSetSession = jest.fn();
            const mockClient = {
                auth: {
                    setSession: mockSetSession,
                },
            };

            // Mock the createClient to return our mock client
            createClient.mockReturnValue(mockClient);

            const accessToken = "test-access-token";
            const client = getSupabaseClient(accessToken);

            expect(mockSetSession).toHaveBeenCalledWith({
                access_token: accessToken,
                refresh_token: "",
            });
            expect(client).toBe(mockClient);
        });
    });
});
