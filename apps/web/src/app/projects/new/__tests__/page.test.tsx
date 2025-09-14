import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import React from "react";

// Mock router to capture navigation
const pushMock = jest.fn();
jest.mock("next/navigation", () => ({
  useRouter() {
    return { push: pushMock, replace: jest.fn(), prefetch: jest.fn(), back: jest.fn(), forward: jest.fn(), refresh: jest.fn() };
  },
  useSearchParams() {
    return new URLSearchParams();
  },
  usePathname() {
    return "/projects/new";
  },
}));

// Mock store to intercept addProject
const addProjectMock = jest.fn();
jest.mock("@/stores", () => ({
  useProjectStore: () => ({ addProject: addProjectMock }),
}));

// Mock fetch
const fetchMock = jest.fn();
global.fetch = fetchMock as any;

import NewProjectPage from "@/app/projects/new/page";

describe("NewProjectPage", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("submits form, calls API, updates store and navigates", async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: async () => ({ project: { id: "proj123", name: "My Project" } }),
    });

    render(<NewProjectPage />);

    const nameInput = screen.getByPlaceholderText(/customer onboarding/i);
    fireEvent.change(nameInput, { target: { value: "My Project" } });

    const desc = screen.getByPlaceholderText(/short summary/i);
    fireEvent.change(desc, { target: { value: "Demo" } });

    const submit = screen.getByRole("button", { name: /create project/i });
    fireEvent.click(submit);

    await waitFor(() => expect(fetchMock).toHaveBeenCalled());
    expect(addProjectMock).toHaveBeenCalled();
    expect(pushMock).toHaveBeenCalledWith("/projects/proj123");
  });
});

