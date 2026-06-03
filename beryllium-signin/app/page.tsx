"use client";

import { useState, useEffect } from "react";
import { encryptPassword } from "./lib/crypto";

export default function SignInPage() {
  const [classId, setClassId] = useState("");
  const [publicKey, setPublicKey] = useState("");
  const [account, setAccount] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");
  const [messageType, setMessageType] = useState<"success" | "error" | "">(
    ""
  );
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (window.__CLASS_ID__ && window.__CLASS_PUBLIC_KEY__) {
      setClassId(window.__CLASS_ID__);
      setPublicKey(window.__CLASS_PUBLIC_KEY__);
      return;
    }

    const path = window.location.pathname;
    const match = path.match(/\/signin\/([a-f0-9-]+)$/);
    if (match) {
      const id = match[1];
      setClassId(id);
      fetch(`http://localhost:8080/api/classes/${id}/public-key`)
        .then((res) => res.json())
        .then((data) => setPublicKey(data.public_key))
        .catch(() => {
          setMessage("无法加载课堂信息");
          setMessageType("error");
        });
    }
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!publicKey) {
      setMessage("课堂公钥不可用");
      setMessageType("error");
      return;
    }

    setLoading(true);
    setMessage("");
    setMessageType("");

    try {
      const encryptedPassword = await encryptPassword(publicKey, password);
      const res = await fetch(
        `http://localhost:8080/signin/${classId}`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            account,
            password: encryptedPassword,
          }),
        }
      );

      const data = await res.json();

      if (res.ok) {
        setMessage(`签到成功！欢迎你，${data.name}`);
        setMessageType("success");
      } else {
        setMessage(data.error || "签到失败，请检查账号和密码");
        setMessageType("error");
      }
    } catch {
      setMessage("网络错误，请重试");
      setMessageType("error");
    } finally {
      setLoading(false);
    }
  }

  if (!classId) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 to-blue-50">
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
            课堂签到
          </h1>
          <p className="mt-1.5 font-mono text-sm text-muted">
            {classId.slice(0, 8)}...
          </p>
        </div>

        {/* Sign-in card */}
        <div className="rounded-2xl bg-white p-8 shadow-sm ring-1 ring-zinc-200/60">
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div>
              <label
                htmlFor="account"
                className="mb-1.5 block text-sm font-medium text-zinc-700"
              >
                学号
              </label>
              <input
                id="account"
                type="text"
                placeholder="请输入学号"
                value={account}
                onChange={(e) => setAccount(e.target.value)}
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

            {message && (
              <div
                className={`rounded-lg px-4 py-3 text-sm ${
                  messageType === "success"
                    ? "bg-success-light text-success"
                    : "bg-danger-light text-danger"
                }`}
              >
                {message}
              </div>
            )}

            <button
              type="submit"
              disabled={loading || !publicKey}
              className="mt-1 w-full rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-white
                shadow-sm transition-all hover:bg-primary-hover focus:outline-none focus:ring-2 focus:ring-primary/30
                disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  签到中...
                </span>
              ) : (
                "立即签到"
              )}
            </button>
          </form>
        </div>

        <p className="mt-6 text-center text-xs text-zinc-300">
          密码传输已使用 RSA 加密保护
        </p>
      </div>
    </div>
  );
}
