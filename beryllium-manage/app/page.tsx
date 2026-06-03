"use client";

import { useState, useEffect } from "react";
import { login } from "./lib/api";
import { setToken, isAuthenticated } from "./lib/auth";

export default function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (isAuthenticated()) {
      window.location.href = "/student-manage";
    }
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await login(username, password);
      setToken(res.token);
      window.location.href = "/student-manage";
    } catch {
      setError("用户名或密码错误");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 to-blue-50 px-4">
      <div className="w-full max-w-md">
        {/* Brand header */}
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-primary shadow-lg shadow-blue-200">
            <svg
              className="h-7 w-7 text-white"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={2}
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">
            Beryllium Manage
          </h1>
          <p className="mt-1.5 text-sm text-muted">课堂签到管理系统</p>
        </div>

        {/* Login card */}
        <div className="rounded-2xl bg-white p-8 shadow-sm ring-1 ring-zinc-200/60">
          <h2 className="mb-6 text-lg font-semibold text-zinc-900">管理员登录</h2>

          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div>
              <label
                htmlFor="username"
                className="mb-1.5 block text-sm font-medium text-zinc-700"
              >
                用户名
              </label>
              <input
                id="username"
                type="text"
                placeholder="请输入用户名"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full rounded-lg border border-zinc-300 px-3.5 py-2.5 text-sm text-zinc-900 placeholder:text-zinc-400
                  transition-colors focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                required
                autoFocus
              />
            </div>

            <div>
              <label
                htmlFor="password"
                className="mb-1.5 block text-sm font-medium text-zinc-700"
              >
                密码
              </label>
              <input
                id="password"
                type="password"
                placeholder="请输入密码"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-lg border border-zinc-300 px-3.5 py-2.5 text-sm text-zinc-900 placeholder:text-zinc-400
                  transition-colors focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                required
              />
            </div>

            {error && (
              <div className="rounded-lg bg-danger-light px-4 py-3 text-sm text-danger">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="mt-1 w-full rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-white
                shadow-sm transition-all hover:bg-primary-hover focus:outline-none focus:ring-2 focus:ring-primary/30
                disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24">
                    <circle
                      className="opacity-25"
                      cx="12" cy="12" r="10"
                      stroke="currentColor" strokeWidth="4" fill="none"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                    />
                  </svg>
                  登录中...
                </span>
              ) : (
                "登录"
              )}
            </button>
          </form>
        </div>

        <p className="mt-6 text-center text-xs text-muted">
          默认账号: admin / admin
        </p>
      </div>
    </div>
  );
}
