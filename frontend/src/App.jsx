import { useMemo, useState } from "react";
import "./App.css";

const API_BASE_DEFAULT = "http://localhost:7540";

export default function App() {
  const [apiBase, setApiBase] = useState(API_BASE_DEFAULT);
  const [token, setToken] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState("auth");

  const [registerForm, setRegisterForm] = useState({ email: "", password: "" });
  const [loginForm, setLoginForm] = useState({ email: "", password: "" });
  const [targetForm, setTargetForm] = useState({
    url: "",
    timeout: "5",
    interval_time: "10",
  });

  const [targets, setTargets] = useState([]);
  const [manualCheckResult, setManualCheckResult] = useState(null);

  const authHeaders = useMemo(() => {
    return token
        ? {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        }
        : { "Content-Type": "application/json" };
  }, [token]);

  const resetStatus = () => {
    setMessage("");
    setError("");
  };

  const api = async (path, options = {}) => {
    const response = await fetch(`${apiBase}${path}`, {
      ...options,
      headers: {
        ...(options.headers || {}),
      },
    });

    const text = await response.text();
    let data = text;

    try {
      data = text ? JSON.parse(text) : null;
    } catch {
      data = text;
    }

    if (!response.ok) {
      throw new Error(typeof data === "string" ? data : JSON.stringify(data));
    }

    return data;
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    resetStatus();
    setLoading(true);
    try {
      const result = await api("/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(registerForm),
      });
      setMessage(typeof result === "string" ? result : "User registered");
    } catch (err) {
      setError(err.message || "Register failed");
    } finally {
      setLoading(false);
    }
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    resetStatus();
    setLoading(true);
    try {
      const result = await api("/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(loginForm),
      });
      setToken(result?.token || "");
      setMessage("Login successful");
      setActiveTab("targets");
    } catch (err) {
      setError(err.message || "Login failed");
    } finally {
      setLoading(false);
    }
  };

  const loadTargets = async () => {
    resetStatus();
    setLoading(true);
    try {
      const result = await api("/targets", {
        method: "GET",
        headers: authHeaders,
      });
      setTargets(Array.isArray(result) ? result : []);
      setMessage("Targets loaded");
    } catch (err) {
      setError(err.message || "Failed to load targets");
    } finally {
      setLoading(false);
    }
  };

  const createTarget = async (e) => {
    e.preventDefault();
    resetStatus();
    setLoading(true);
    try {
      await api("/targets", {
        method: "POST",
        headers: authHeaders,
        body: JSON.stringify({
          url: targetForm.url,
          timeout: Number(targetForm.timeout),
          interval_time: Number(targetForm.interval_time),
        }),
      });
      setMessage("Target created");
      setTargetForm({ url: "", timeout: "5", interval_time: "10" });
      await loadTargets();
    } catch (err) {
      setError(err.message || "Failed to create target");
    } finally {
      setLoading(false);
    }
  };

  const deleteTarget = async (id) => {
    resetStatus();
    setLoading(true);
    try {
      await api(`/targets/${id}`, {
        method: "DELETE",
        headers: authHeaders,
      });
      setMessage(`Target ${id} deleted`);
      setTargets((prev) => prev.filter((t) => t.id !== id));
    } catch (err) {
      setError(err.message || "Failed to delete target");
    } finally {
      setLoading(false);
    }
  };

  const manualCheck = async (id) => {
    resetStatus();
    setLoading(true);
    try {
      const result = await api(`/targets/${id}/check`, {
        method: "GET",
        headers: authHeaders,
      });
      setManualCheckResult(result);
      setMessage(`Manual check completed for target ${id}`);
    } catch (err) {
      setError(err.message || "Manual check failed");
    } finally {
      setLoading(false);
    }
  };

  return (
      <div className="page">
        <div className="shell">
          <header className="hero">
            <div>
              <p className="eyebrow">Go Backend Playground</p>
              <h1>Uptime Monitor</h1>
              <p className="subtitle">
                Удобная панель, чтобы руками проверить auth, targets и manual check.
              </p>
            </div>

            <div className="api-box">
              <label>API Base URL</label>
              <input
                  value={apiBase}
                  onChange={(e) => setApiBase(e.target.value)}
                  placeholder="http://localhost:7540"
              />
            </div>
          </header>

          {(message || error) && (
              <div className={error ? "status error" : "status success"}>
                {error || message}
              </div>
          )}

          <div className="tabs">
            <button
                className={activeTab === "auth" ? "tab active" : "tab"}
                onClick={() => setActiveTab("auth")}
            >
              Auth
            </button>
            <button
                className={activeTab === "targets" ? "tab active" : "tab"}
                onClick={() => setActiveTab("targets")}
            >
              Targets
            </button>
          </div>

          {activeTab === "auth" && (
              <section className="grid two">
                <div className="card">
                  <h2>Register</h2>
                  <p className="card-subtitle">Создай пользователя для теста.</p>
                  <form onSubmit={handleRegister} className="form">
                    <label>Email</label>
                    <input
                        value={registerForm.email}
                        onChange={(e) =>
                            setRegisterForm((p) => ({ ...p, email: e.target.value }))
                        }
                        placeholder="user@example.com"
                    />

                    <label>Password</label>
                    <input
                        type="password"
                        value={registerForm.password}
                        onChange={(e) =>
                            setRegisterForm((p) => ({ ...p, password: e.target.value }))
                        }
                        placeholder="password"
                    />

                    <button type="submit" disabled={loading}>
                      Register
                    </button>
                  </form>
                </div>

                <div className="card">
                  <h2>Login</h2>
                  <p className="card-subtitle">
                    После логина токен сохранится в поле ниже.
                  </p>
                  <form onSubmit={handleLogin} className="form">
                    <label>Email</label>
                    <input
                        value={loginForm.email}
                        onChange={(e) =>
                            setLoginForm((p) => ({ ...p, email: e.target.value }))
                        }
                        placeholder="user@example.com"
                    />

                    <label>Password</label>
                    <input
                        type="password"
                        value={loginForm.password}
                        onChange={(e) =>
                            setLoginForm((p) => ({ ...p, password: e.target.value }))
                        }
                        placeholder="password"
                    />

                    <button type="submit" disabled={loading}>
                      Login
                    </button>
                  </form>

                  <div className="token-box">
                    <label>JWT Token</label>
                    <textarea
                        value={token}
                        onChange={(e) => setToken(e.target.value)}
                        placeholder="Token appears here after login"
                    />
                  </div>
                </div>
              </section>
          )}

          {activeTab === "targets" && (
              <section className="grid targets-layout">
                <div className="card">
                  <h2>Create Target</h2>
                  <p className="card-subtitle">
                    Добавь сайт с таймаутом и интервалом проверки.
                  </p>

                  <form onSubmit={createTarget} className="form">
                    <label>URL</label>
                    <input
                        value={targetForm.url}
                        onChange={(e) =>
                            setTargetForm((p) => ({ ...p, url: e.target.value }))
                        }
                        placeholder="https://example.com"
                    />

                    <label>Timeout (sec)</label>
                    <input
                        type="number"
                        min="1"
                        value={targetForm.timeout}
                        onChange={(e) =>
                            setTargetForm((p) => ({ ...p, timeout: e.target.value }))
                        }
                    />

                    <label>Interval (sec)</label>
                    <input
                        type="number"
                        min="1"
                        value={targetForm.interval_time}
                        onChange={(e) =>
                            setTargetForm((p) => ({
                              ...p,
                              interval_time: e.target.value,
                            }))
                        }
                    />

                    <button type="submit" disabled={loading}>
                      Create Target
                    </button>
                  </form>

                  <button className="secondary-btn" onClick={loadTargets} disabled={loading}>
                    Refresh Targets
                  </button>
                </div>

                <div className="stack">
                  <div className="card">
                    <div className="card-header-row">
                      <div>
                        <h2>Targets</h2>
                        <p className="card-subtitle">
                          Список targets текущего пользователя.
                        </p>
                      </div>
                      <span className="badge">{targets.length}</span>
                    </div>

                    {targets.length === 0 ? (
                        <p className="empty">No targets loaded yet.</p>
                    ) : (
                        <div className="targets">
                          {targets.map((target) => (
                              <div key={target.id} className="target-item">
                                <div className="target-main">
                                  <div className="target-url">{target.url}</div>
                                  <div className="target-meta">
                                    ID: {target.id} · Timeout: {target.timeout}s · Interval:{" "}
                                    {target.interval_time}s
                                  </div>
                                </div>

                                <div className="actions">
                                  <button
                                      className="secondary-btn"
                                      onClick={() => manualCheck(target.id)}
                                      disabled={loading}
                                  >
                                    Check now
                                  </button>
                                  <button
                                      className="danger-btn"
                                      onClick={() => deleteTarget(target.id)}
                                      disabled={loading}
                                  >
                                    Delete
                                  </button>
                                </div>
                              </div>
                          ))}
                        </div>
                    )}
                  </div>

                  <div className="card">
                    <h2>Manual Check Result</h2>
                    <p className="card-subtitle">
                      Ответ ручной проверки выбранного target.
                    </p>

                    {manualCheckResult ? (
                        <pre>{JSON.stringify(manualCheckResult, null, 2)}</pre>
                    ) : (
                        <p className="empty">Run a manual check to see the result.</p>
                    )}
                  </div>
                </div>
              </section>
          )}
        </div>
      </div>
  );
}