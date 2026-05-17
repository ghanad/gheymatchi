import type { Metadata } from "next";
import type { ReactNode } from "react";
import "./globals.css";
import { Navigation } from "../components/navigation";

export const metadata: Metadata = {
  title: "GheymatChi",
  description: "Local price tracking dashboard"
};

export default function RootLayout({
  children
}: Readonly<{
  children: ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <div className="app-shell">
          <header className="topbar">
            <div className="topbar-inner">
              <a className="brand" href="/">
                GheymatChi
              </a>
              <Navigation />
            </div>
          </header>
          <main className="main">{children}</main>
        </div>
      </body>
    </html>
  );
}
