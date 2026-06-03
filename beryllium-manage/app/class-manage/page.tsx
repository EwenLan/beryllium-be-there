"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import {
  fetchClasses,
  createClass,
  deleteClass,
  fetchClassDetail,
  initAttendance,
  downloadCSV,
} from "../lib/api";
import { isAuthenticated, clearToken } from "../lib/auth";
import type { ClassInfo, ClassDetail } from "../lib/types";

export default function ClassManagePage() {
  const [classId, setClassId] = useState<string | null>(null);

  useEffect(() => {
    if (!isAuthenticated()) {
      window.location.href = "/";
      return;
    }
    const path = window.location.pathname;
    const match = path.match(/^\/class-manage\/([a-f0-9-]+)$/);
    if (match) {
      setClassId(match[1]);
    }
  }, []);

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
                className="rounded-lg px-3 py-1.5 text-sm font-medium text-muted hover:bg-zinc-100 transition-colors"
              >
                学生管理
              </a>
              <a
                href="/class-manage"
                className="rounded-lg bg-primary-light px-3 py-1.5 text-sm font-medium text-primary"
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
        {classId ? (
          <ClassDetailView classId={classId} onBack={() => {
            window.location.href = "/class-manage";
          }} />
        ) : (
          <ClassListView />
        )}
      </main>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Class list view                                                   */
/* ------------------------------------------------------------------ */

function ClassListView() {
  const [classes, setClasses] = useState<ClassInfo[]>([]);
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const loadClasses = useCallback(async () => {
    try {
      const data = await fetchClasses();
      setClasses(data);
    } catch {
      // api layer handles auth errors
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadClasses();
  }, [loadClasses]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      await createClass(name);
      setName("");
      await loadClasses();
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建失败");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id: string, className: string) {
    if (!window.confirm(`确认删除课堂 "${className}"？\n这将同时删除该课堂的签到记录。`)) return;
    try {
      await deleteClass(id);
      await loadClasses();
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除失败");
    }
  }

  return (
    <>
      <div className="mb-8">
        <h2 className="text-xl font-bold text-zinc-900">课堂管理</h2>
        <p className="mt-1 text-sm text-muted">创建课堂并管理签到</p>
      </div>

      {/* Create form */}
      <div className="mb-8 rounded-2xl bg-white p-6 shadow-sm ring-1 ring-zinc-200/60">
        <h3 className="mb-4 text-sm font-semibold text-zinc-700">创建课堂</h3>
        <form onSubmit={handleCreate} className="flex gap-3">
          <input
            type="text"
            placeholder="输入课堂名称，如：高等数学"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="flex-1 rounded-lg border border-zinc-300 px-3.5 py-2.5 text-sm text-zinc-900 placeholder:text-zinc-400
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
            {submitting ? "创建中..." : "创建课堂"}
          </button>
        </form>
      </div>

      {error && (
        <div className="mb-6 rounded-lg bg-danger-light px-4 py-3 text-sm text-danger">
          {error}
          <button onClick={() => setError("")} className="ml-2 font-medium hover:underline">关闭</button>
        </div>
      )}

      {/* Class grid */}
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
      ) : classes.length === 0 ? (
        <div className="flex flex-col items-center gap-3 py-16">
          <svg className="h-10 w-10 text-zinc-300" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 016-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0018 18a8.967 8.967 0 00-6 2.292m0-14.25v14.25" />
          </svg>
          <p className="text-sm text-zinc-400">暂无课堂</p>
          <p className="text-xs text-zinc-300">请在上方创建第一个课堂</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {classes.map((c) => (
            <div
              key={c.class_id}
              className="group rounded-2xl bg-white p-6 shadow-sm ring-1 ring-zinc-200/60 transition-all hover:shadow-md hover:ring-zinc-300"
            >
              <div className="mb-4 flex items-start justify-between">
                <div className="flex-1">
                  <h3 className="font-semibold text-zinc-900">{c.name}</h3>
                  <p className="mt-1 text-xs text-muted">
                    创建于 {c.created_at}
                  </p>
                </div>
                <span className="rounded-full bg-primary-light px-2 py-0.5 text-xs font-medium text-primary">
                  课堂
                </span>
              </div>

              <p className="mb-4 truncate rounded-lg bg-zinc-50 px-3 py-2 text-xs text-muted font-mono">
                {c.class_id}
              </p>

              <div className="flex gap-2">
                <button
                  onClick={() => {
                    window.location.href = `/class-manage/${c.class_id}`;
                  }}
                  className="flex-1 rounded-lg bg-primary px-3 py-2 text-xs font-semibold text-white
                    transition-colors hover:bg-primary-hover"
                >
                  查看签到
                </button>
                <button
                  onClick={() => handleDelete(c.class_id, c.name)}
                  className="rounded-lg px-3 py-2 text-xs font-medium text-danger
                    transition-colors hover:bg-danger-light"
                >
                  删除
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </>
  );
}

/* ------------------------------------------------------------------ */
/*  Class detail view (attendance)                                    */
/* ------------------------------------------------------------------ */

function ClassDetailView({
  classId,
  onBack,
}: {
  classId: string;
  onBack: () => void;
}) {
  const [detail, setDetail] = useState<ClassDetail | null>(null);
  const [error, setError] = useState("");
  const [exporting, setExporting] = useState(false);
  const [initializing, setInitializing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  const loadDetail = useCallback(async () => {
    try {
      const data = await fetchClassDetail(classId);
      setDetail(data);
      setLastUpdated(new Date());
      setError("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载失败");
    }
  }, [classId]);

  useEffect(() => {
    loadDetail();
  }, [loadDetail]);

  // WebSocket connection for real-time attendance updates
  useEffect(() => {
    if (!autoRefresh) return;

    const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
    const wsHost = window.location.port === "8080"
      ? window.location.host
      : "localhost:8080";
    const wsURL = `${wsProtocol}://${wsHost}/ws/classes/${classId}/attendance`;

    let ws: WebSocket | null = null;
    let reconnectTimer: ReturnType<typeof setTimeout>;

    function connect() {
      ws = new WebSocket(wsURL);

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          if (msg.type === "attendance_update" && msg.attendance) {
            setDetail((prev) =>
              prev ? { ...prev, attendance: msg.attendance } : prev
            );
            setLastUpdated(new Date());
          }
        } catch {
          // ignore malformed messages
        }
      };

      ws.onclose = () => {
        // Reconnect after 3 seconds if still active
        if (autoRefresh) {
          reconnectTimer = setTimeout(connect, 3000);
        }
      };

      ws.onerror = () => {
        ws?.close();
      };
    }

    connect();

    return () => {
      clearTimeout(reconnectTimer);
      ws?.close();
    };
  }, [classId, autoRefresh]);

  useEffect(() => {
    if (detail?.signin_url && canvasRef.current) {
      import("qrcode").then((QRCode) => {
        QRCode.toCanvas(canvasRef.current, detail.signin_url, {
          width: 180,
          margin: 1,
        });
      });
    }
  }, [detail?.signin_url]);

  async function handleInitAttendance() {
    setInitializing(true);
    try {
      await initAttendance(classId);
      await loadDetail();
    } catch (err) {
      setError(err instanceof Error ? err.message : "初始化失败");
    } finally {
      setInitializing(false);
    }
  }

  async function handleExport() {
    setExporting(true);
    try {
      await downloadCSV(classId);
    } catch (err) {
      setError(err instanceof Error ? err.message : "导出失败");
    } finally {
      setExporting(false);
    }
  }

  async function copySigninUrl() {
    if (detail?.signin_url) {
      await navigator.clipboard.writeText(detail.signin_url);
    }
  }

  if (error && !detail) {
    return (
      <div>
        <button onClick={onBack} className="mb-6 inline-flex items-center gap-1 text-sm text-muted hover:text-zinc-700 transition-colors">
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18" />
          </svg>
          返回课堂列表
        </button>
        <div className="rounded-lg bg-danger-light px-4 py-3 text-sm text-danger">{error}</div>
      </div>
    );
  }

  if (!detail) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="flex items-center gap-2 text-sm text-muted">
          <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          加载中...
        </div>
      </div>
    );
  }

  return (
    <div>
      <button onClick={onBack} className="mb-6 inline-flex items-center gap-1 text-sm text-muted hover:text-zinc-700 transition-colors">
        <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18" />
        </svg>
        返回课堂列表
      </button>

      {error && (
        <div className="mb-6 rounded-lg bg-danger-light px-4 py-3 text-sm text-danger">
          {error}
          <button onClick={() => setError("")} className="ml-2 font-medium hover:underline">关闭</button>
        </div>
      )}

      <h2 className="mb-1 text-xl font-bold text-zinc-900">{detail.name}</h2>
      <p className="mb-8 text-sm text-muted">课堂签到详情</p>

      {/* Info card with QR code */}
      <div className="mb-8 overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-zinc-200/60">
        <div className="flex flex-col gap-6 p-6 sm:flex-row">
          {/* Left: class info */}
          <div className="flex-1">
            <h3 className="text-sm font-semibold text-zinc-700">签到信息</h3>

            <div className="mt-3 space-y-3">
              <div>
                <label className="text-xs text-muted">课堂 ID</label>
                <p className="mt-0.5 rounded-lg bg-zinc-50 px-3 py-2 text-sm font-mono text-zinc-600">
                  {detail.class_id}
                </p>
              </div>

              <div>
                <label className="text-xs text-muted">签到地址</label>
                <div className="mt-0.5 flex items-center gap-2">
                  <a
                    href={detail.signin_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex-1 truncate rounded-lg bg-primary-light px-3 py-2 text-sm text-primary hover:underline"
                  >
                    {detail.signin_url}
                  </a>
                  <button
                    onClick={copySigninUrl}
                    className="flex-shrink-0 rounded-lg p-2 text-muted hover:bg-zinc-100 transition-colors"
                    title="复制链接"
                  >
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* Right: QR code */}
          <div className="flex flex-col items-center gap-2">
            <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white p-2">
              <canvas ref={canvasRef} />
            </div>
            <p className="text-xs text-muted">扫码签到</p>
          </div>
        </div>

        {/* Action bar */}
        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-zinc-100 bg-zinc-50/50 px-6 py-3">
          <div className="flex gap-3">
            {detail.attendance.length === 0 ? (
              <button
                onClick={handleInitAttendance}
                disabled={initializing}
                className="rounded-lg bg-success px-4 py-2 text-sm font-semibold text-white
                  transition-colors hover:bg-green-600 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {initializing ? "初始化中..." : "初始化签到表"}
              </button>
            ) : (
              <button
                onClick={handleExport}
                disabled={exporting}
                className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white
                  transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-60"
              >
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" />
                </svg>
                {exporting ? "导出中..." : "导出 CSV"}
              </button>
            )}
          </div>

          {/* Refresh controls */}
          <div className="flex items-center gap-3 text-xs text-muted">
            {lastUpdated && (
              <span title="上次刷新时间">
                更新于 {lastUpdated.toLocaleTimeString("zh-CN")}
              </span>
            )}
            <button
              onClick={() => loadDetail()}
              className="inline-flex items-center gap-1 rounded-md px-2 py-1 hover:bg-zinc-200 transition-colors"
              title="手动刷新"
            >
              <svg className={`h-3.5 w-3.5 ${autoRefresh ? "" : "animate-spin"}`} fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182" />
              </svg>
            </button>
            <label className="flex items-center gap-1.5 cursor-pointer">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="h-3.5 w-3.5 rounded border-zinc-300 text-primary focus:ring-primary"
              />
              自动刷新
            </label>
          </div>
        </div>
      </div>

      {/* Attendance table */}
      {detail.attendance.length > 0 && (
        <div className="overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-zinc-200/60">
          <div className="border-b border-zinc-200 bg-zinc-50/50 px-6 py-3.5">
            <h3 className="text-sm font-semibold text-zinc-700">
              签到记录
              <span className="ml-2 font-normal text-muted">
                {detail.attendance.filter((e) => e.is_present).length}
                /
                {detail.attendance.length}
                已签到
              </span>
            </h3>
          </div>
          <table className="w-full">
            <thead>
              <tr className="border-b border-zinc-200 bg-zinc-50/30">
                <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">学号</th>
                <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">姓名</th>
                <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">状态</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-100">
              {detail.attendance.map((entry) => (
                <tr key={entry.account} className="transition-colors hover:bg-zinc-50/50">
                  <td className="px-6 py-4 text-sm font-medium text-zinc-900">{entry.account}</td>
                  <td className="px-6 py-4 text-sm text-zinc-700">{entry.name}</td>
                  <td className="px-6 py-4">
                    {entry.is_present ? (
                      <span className="inline-flex items-center gap-1 rounded-full bg-success-light px-2.5 py-1 text-xs font-medium text-success">
                        <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" strokeWidth={3} stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                        </svg>
                        已签到
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 rounded-full bg-zinc-100 px-2.5 py-1 text-xs font-medium text-muted">
                        <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        未签到
                      </span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
