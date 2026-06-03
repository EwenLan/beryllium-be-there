import { getToken, clearToken } from "./auth";
import type {
  Student,
  ClassInfo,
  ClassDetail,
  LoginResponse,
  AttendanceEntry,
} from "./types";

// In production (served by Go on :8080), use same origin.
// In development (Next.js on :3000), use the Go backend directly.
function getAPIBase(): string {
  if (typeof window === "undefined") return "http://localhost:8080";
  if (window.location.port === "8080") return "";
  return "http://localhost:8080";
}

const API_BASE = getAPIBase();

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((options.headers as Record<string, string>) || {}),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const data = await res.json();

  if (!res.ok) {
    if (res.status === 401 && token) {
      clearToken();
      window.location.href = "/";
    }
    throw new Error(data.error || "request failed");
  }

  return data as T;
}

export async function login(
  username: string,
  password: string
): Promise<LoginResponse> {
  return request<LoginResponse>("/api/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
}

export async function fetchStudents(): Promise<Student[]> {
  return request<Student[]>("/api/students");
}

export async function createStudent(data: {
  account: string;
  name: string;
  password: string;
}): Promise<void> {
  await request("/api/students", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function deleteStudent(account: string): Promise<void> {
  await request(`/api/students/${account}`, { method: "DELETE" });
}

export async function fetchClasses(): Promise<ClassInfo[]> {
  return request<ClassInfo[]>("/api/classes");
}

export async function createClass(name: string): Promise<ClassInfo> {
  return request<ClassInfo>("/api/classes", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export async function deleteClass(classId: string): Promise<void> {
  await request(`/api/classes/${classId}`, { method: "DELETE" });
}

export async function fetchClassDetail(classId: string): Promise<ClassDetail> {
  return request<ClassDetail>(`/api/classes/${classId}`);
}

export async function fetchAttendance(
  classId: string
): Promise<AttendanceEntry[]> {
  return request<AttendanceEntry[]>(`/api/classes/${classId}/attendance`);
}

export async function initAttendance(classId: string): Promise<AttendanceEntry[]> {
  return request<AttendanceEntry[]>(`/api/classes/${classId}/attendance`, {
    method: "POST",
  });
}

export async function downloadCSV(classId: string): Promise<void> {
  const token = getToken();
  const res = await fetch(`${API_BASE}/api/classes/${classId}/export`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error("export failed");
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "attendance.csv";
  a.click();
  URL.revokeObjectURL(url);
}
