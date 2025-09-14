import { supabase } from "../supabase";

// Mock the Supabase client
jest.mock("@supabase/supabase-js", () => ({
    createClient: jest.fn(() => ({
        auth: {
            autoRefreshToken: true,
            persistSession: true,
            detectSessionInUrl: true,
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

describe("Supabase Client Configuration", () => {
    beforeEach(() => {
        // Reset environment variables
        delete process.env.NEXT_PUBLIC_SUPABASE_URL;
        delete process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY;
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it("should throw error when NEXT_PUBLIC_SUPABASE_URL is missing", () => {
        // Mock environment variables
        process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = "test-anon-key";

        // Clear the module cache to force re-import
        jest.resetModules();

        expect(() => {
            require("../supabase");
        }).toThrow("Missing NEXT_PUBLIC_SUPABASE_URL environment variable");
    });

    it("should throw error when NEXT_PUBLIC_SUPABASE_ANON_KEY is missing", () => {
        // Mock environment variables
        process.env.NEXT_PUBLIC_SUPABASE_URL = "https://test.supabase.co";

        // Clear the module cache to force re-import
        jest.resetModules();

        expect(() => {
            require("../supabase");
        }).toThrow(
            "Missing NEXT_PUBLIC_SUPABASE_ANON_KEY environment variable",
        );
    });

    it("should create client with correct configuration when environment variables are set", () => {
        // Mock environment variables
        process.env.NEXT_PUBLIC_SUPABASE_URL = "https://test.supabase.co";
        process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = "test-anon-key";

        // Clear the module cache to force re-import
        jest.resetModules();

        const { createClient } = require("@supabase/supabase-js");
        const { supabase: client } = require("../supabase");

        expect(createClient).toHaveBeenCalledWith(
            "https://test.supabase.co",
            "test-anon-key",
            {
                auth: {
                    autoRefreshToken: true,
                    persistSession: true,
                    detectSessionInUrl: true,
                },
            },
        );
        expect(client).toBeDefined();
    });

    it("should export supabase client", () => {
        // Mock environment variables
        process.env.NEXT_PUBLIC_SUPABASE_URL = "https://test.supabase.co";
        process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = "test-anon-key";

        // Clear the module cache to force re-import
        jest.resetModules();

        const { supabase: client } = require("../supabase");
        expect(client).toBeDefined();
        expect(client.auth).toBeDefined();
    });
});
