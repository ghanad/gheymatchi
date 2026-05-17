import Link from "next/link";

const links = [
  { href: "/", label: "Dashboard" },
  { href: "/products", label: "Products" },
  { href: "/alerts", label: "Alerts" },
  { href: "/settings", label: "Settings" }
];

export function Navigation() {
  return (
    <nav className="nav" aria-label="Primary navigation">
      {links.map((link) => (
        <Link className="nav-link" href={link.href} key={link.href}>
          {link.label}
        </Link>
      ))}
    </nav>
  );
}
