/**
 * Unit tests for POST /api/projects route
 * Mocks Supabase admin client to test responses deterministically
 */

// Mock Next.js server module to avoid environment dependency on Next's Request
jest.mock("next/server", () => ({
  NextResponse: {
    json: (data: any, init?: any) => ({
      status: init?.status ?? 200,
      json: async () => data,
    }),
  },
}));
import { NextResponse } from "next/server";

// Provide a basic module mock; implementation is configured per-test
jest.mock("@/lib/supabase-admin", () => ({
  supabaseAdmin: { from: jest.fn() },
}));

// Import after mocks are set up
import { POST } from "@/app/api/projects/route";
const { supabaseAdmin } = jest.requireMock("@/lib/supabase-admin");
const fromMock = supabaseAdmin.from as jest.Mock;

function makeRequest(body: any) {
  // Minimal stub that satisfies route's usage (req.json())
  return { json: async () => body } as any as Request;
}

describe("POST /api/projects (unit)", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("returns 400 when name is missing", async () => {
    const req = makeRequest({ description: "x" });
    const res = await POST(req);
    expect(res.status).toBe(400);
    const data = await (res as NextResponse).json();
    expect(data.error).toMatch(/name is required/i);
    expect(fromMock).not.toHaveBeenCalled();
  });

  it("creates project and returns 201 on success", async () => {
    // Configure mock chain success
    const single = jest.fn().mockResolvedValueOnce({
      data: { id: "proj_123", name: "My Project", status: "draft" },
      error: null,
    });
    const select = jest.fn(() => ({ single }));
    const insert = jest.fn(() => ({ select }));
    fromMock.mockImplementationOnce(() => ({ insert }));

    const req = makeRequest({ name: "My Project" });
    const res = await POST(req);
    expect(res.status).toBe(201);
    const data = await (res as NextResponse).json();
    expect(data.project).toBeDefined();
    expect(data.project.id).toBe("proj_123");
    expect(fromMock).toHaveBeenCalledWith("projects");
    expect(insert).toHaveBeenCalled();
    expect(select).toHaveBeenCalled();
    expect(single).toHaveBeenCalled();
  });

  it("returns 500 when Supabase insert fails", async () => {
    const single = jest
      .fn()
      .mockResolvedValueOnce({ data: null, error: { message: "db error" } });
    const select = jest.fn(() => ({ single }));
    const insert = jest.fn(() => ({ select }));
    fromMock.mockImplementationOnce(() => ({ insert }));

    const req = makeRequest({ name: "Bad" });
    const res = await POST(req);
    expect(res.status).toBe(500);
    const data = await (res as NextResponse).json();
    expect(data.error).toMatch(/db error/i);
  });
});
