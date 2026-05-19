"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { getAuthToken, logoutUser } from "../lib/api";

const links = [
  { href: "/", label: "Dashboard" },
  { href: "/products", label: "Products" },
  { href: "/alerts", label: "Alerts" },
  { href: "/notifications", label: "Notifications" },
  { href: "/settings", label: "Settings" }
];

export function Navigation() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    setIsAuthenticated(Boolean(getAuthToken()));
  }, []);

  async function handleLogout() {
    await logoutUser();
    setIsAuthenticated(false);
    window.location.href = "/login";
  }

  return (
    <nav className="nav" aria-label="Primary navigation">
      {links.map((link) => (
        <Link className="nav-link" href={link.href} key={link.href}>
          {link.label}
        </Link>
      ))}
      {isAuthenticated ? (
        <button className="nav-button" type="button" onClick={handleLogout}>
          Logout
        </button>
      ) : (
        <>
          <Link className="nav-link" href="/login">
            Login
          </Link>
          <Link className="nav-link" href="/register">
            Register
          </Link>
        </>
      )}
    </nav>
  );
}
