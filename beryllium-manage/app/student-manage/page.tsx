"use client";

import { useState, useEffect, useCallback } from "react";
import { fetchStudents, createStudent, deleteStudent } from "../lib/api";
import { isAuthenticated, clearToken } from "../lib/auth";
import type { Student } from "../lib/types";

export default function StudentManagePage() {
  const [students, setStudents] = useState<Student[]>([]);
  const [account, setAccount] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const loadStudents = useCallback(async () => {
    try {
      const data = await fetchStudents();
      setStudents(data);
    } catch {
      // api layer handles auth errors
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!isAuthenticated()) {
      window.location.href = "/";
      return;
    }
    loadStudents();
  }, [loadStudents]);

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSuccess("");
    setSubmitting(true);
    try {
      await createStudent({ account, name, password });
      setAccount("");
      setName("");
      setPassword("");
      setSuccess(`已添加学生: ${name} (${account})`);
      await loadStudents();
      setTimeout(() => setSuccess(""), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "添加失败");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(acc: string, studentName: string) {
    if (!window.confirm(`确认删除学生 "${studentName}" (${acc})？`)) return;
    setError("");
    try {
      await deleteStudent(acc);
      await loadStudents();
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除失败");
    }
  }

  function handleLogout() {
    clearToken();
    window.location.href = "/";
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
      {/* Navigation bar */}
      <nav className="sticky top-0 z-10 border-b border-zinc-200/60 bg-white/80 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
          <div className="flex items-center gap-6">
            <h1 className="text-lg font-bold text-zinc-900">
              <a href="/student-manage" className="hover:text-primary transition-colors">
                Beryllium
              </a>
            </h1>
            <div className="flex gap-1">
              <a
                href="/student-manage"
                className="rounded-lg bg-primary-light px-3 py-1.5 text-sm font-medium text-primary"
              >
                学生管理
              </a>
              <a
                href="/class-manage"
                className="rounded-lg px-3 py-1.5 text-sm font-medium text-muted hover:bg-zinc-100 transition-colors"
              >
                课堂管理
              </a>
            </div>
          </div>
          <button
            onClick={handleLogout}
            className="rounded-lg px-3 py-1.5 text-sm text-muted hover:bg-zinc-100 hover:text-zinc-700 transition-colors"
          >
            退出登录
          </button>
        </div>
      </nav>

      <main className="mx-auto max-w-5xl px-4 py-8">
        {/* Page header */}
        <div className="mb-8">
          <h2 className="text-xl font-bold text-zinc-900">学生管理</h2>
          <p className="mt-1 text-sm text-muted">管理学生账户，添加或删除学生信息</p>
        </div>

        {/* Add student form */}
        <div className="mb-8 rounded-2xl bg-white p-6 shadow-sm ring-1 ring-zinc-200/60">
          <h3 className="mb-4 text-sm font-semibold text-zinc-700">添加学生</h3>
          <form onSubmit={handleAdd} className="flex flex-wrap gap-3">
            <input
              type="text"
              placeholder="学号"
              value={account}
              onChange={(e) => setAccount(e.target.value)}
              className="min-w-0 flex-1 rounded-lg border border-zinc-300 px-3.5 py-2.5 text-sm text-zinc-900 placeholder:text-zinc-400
                transition-colors focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
              required
            />
            <input
              type="text"
              placeholder="姓名"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="min-w-0 flex-1 rounded-lg border border-zinc-300 px-3.5 py-2.5 text-sm text-zinc-900 placeholder:text-zinc-400
                transition-colors focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
              required
            />
            <input
              type="password"
              placeholder="密码"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="min-w-0 flex-1 rounded-lg border border-zinc-300 px-3.5 py-2.5 text-sm text-zinc-900 placeholder:text-zinc-400
                transition-colors focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
              required
            />
            <button
              type="submit"
              disabled={submitting}
              className="rounded-lg bg-primary px-5 py-2.5 text-sm font-semibold text-white shadow-sm
                transition-all hover:bg-primary-hover focus:outline-none focus:ring-2 focus:ring-primary/30
                disabled:cursor-not-allowed disabled:opacity-60"
            >
              {submitting ? "添加中..." : "添加学生"}
            </button>
          </form>
        </div>

        {/* Messages */}
        {error && (
          <div className="mb-6 rounded-lg bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
            <button onClick={() => setError("")} className="ml-2 font-medium hover:underline">关闭</button>
          </div>
        )}
        {success && (
          <div className="mb-6 rounded-lg bg-success-light px-4 py-3 text-sm text-success">
            {success}
          </div>
        )}

        {/* Student table */}
        <div className="overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-zinc-200/60">
          {loading ? (
            <div className="flex items-center justify-center py-16">
              <div className="flex items-center gap-2 text-sm text-muted">
                <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
                加载中...
              </div>
            </div>
          ) : (
            <table className="w-full">
              <thead>
                <tr className="border-b border-zinc-200 bg-zinc-50/50">
                  <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider text-muted">
                    学号
                  </th>
                  <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider text-muted">
                    姓名
                  </th>
                  <th className="px-6 py-3.5 text-right text-xs font-semibold uppercase tracking-wider text-muted">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-100">
                {students.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="px-6 py-16 text-center">
                      <div className="flex flex-col items-center gap-2">
                        <svg className="h-8 w-8 text-zinc-300" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
                        </svg>
                        <p className="text-sm text-zinc-400">暂无学生</p>
                        <p className="text-xs text-zinc-300">请在上方添加学生信息</p>
                      </div>
                    </td>
                  </tr>
                ) : (
                  students.map((s) => (
                    <tr key={s.account} className="transition-colors hover:bg-zinc-50/50">
                      <td className="px-6 py-4">
                        <span className="text-sm font-medium text-zinc-900">{s.account}</span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-sm text-zinc-700">{s.name}</span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <button
                          onClick={() => handleDelete(s.account, s.name)}
                          className="inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-medium text-danger
                            transition-colors hover:bg-danger-light"
                        >
                          <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                          </svg>
                          删除
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          )}
        </div>

        {/* Footer */}
        <p className="mt-6 text-center text-xs text-zinc-300">
          共 {students.length} 名学生
        </p>
      </main>
    </div>
  );
}
