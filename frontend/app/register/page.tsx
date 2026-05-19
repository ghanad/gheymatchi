"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { registerUser, setAuthToken } from "../../lib/api";

export default function RegisterPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setIsSaving(true);
      const session = await registerUser(email, password);
      setAuthToken(session.token);
      window.location.href = "/products";
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not register");
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <>
      <section className="page-heading">
        <span className="eyebrow">Account</span>
        <h1>Create account</h1>
      </section>

      <section className="panel auth-panel">
        {error ? <div className="form-error">{error}</div> : null}
        <form className="auth-form" onSubmit={handleSubmit}>
          <label className="field">
            <span>Email</span>
            <input type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
          </label>
          <label className="field">
            <span>Password</span>
            <input
              type="password"
              minLength={8}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
            />
          </label>
          <div className="button-row">
            <button type="submit" disabled={isSaving}>
              Register
            </button>
            <Link className="secondary-link-button" href="/login">
              Log in
            </Link>
          </div>
        </form>
      </section>
    </>
  );
}
