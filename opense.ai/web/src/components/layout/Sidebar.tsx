"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  CandlestickChart,
  MessageSquare,
  Terminal,
  Briefcase,
  Filter,
  FlaskConical,
  Settings,
  PanelLeftClose,
  PanelLeft,
  ClipboardList,
  ArrowLeftRight,
  Wallet,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Tooltip } from "@/components/ui/tooltip";

const navItems = [
  { href: "/", icon: LayoutDashboard, label: "Dashboard" },
  { href: "/charts", icon: CandlestickChart, label: "Charts" },
  { href: "/chat", icon: MessageSquare, label: "Chat" },
  { href: "/financeql", icon: Terminal, label: "FinanceQL" },
  { href: "/portfolio", icon: Briefcase, label: "Portfolio" },
  { href: "/orders", icon: ClipboardList, label: "Orders" },
  { href: "/positions", icon: ArrowLeftRight, label: "Positions" },
  { href: "/funds", icon: Wallet, label: "Funds" },
  { href: "/screener", icon: Filter, label: "Screener" },
  { href: "/backtest", icon: FlaskConical, label: "Backtest" },
  { href: "/settings", icon: Settings, label: "Settings" },
];

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const pathname = usePathname();

  return (
    <aside
      className={cn(
        "flex h-screen flex-col border-r bg-card transition-all duration-300",
        collapsed ? "w-16" : "w-56",
      )}
    >
      {/* Logo / Brand */}
      <div className="flex h-14 items-center border-b px-4">
        {!collapsed && (
          <Link href="/" className="flex items-center gap-2">
            <CandlestickChart className="h-6 w-6 text-primary" />
            <span className="font-bold text-lg">OpeNSE.ai</span>
          </Link>
        )}
        {collapsed && (
          <Link href="/" className="mx-auto">
            <CandlestickChart className="h-6 w-6 text-primary" />
          </Link>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 p-2">
        {navItems.map((item) => {
          const isActive = pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));
          const linkContent = (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground",
                isActive && "bg-accent text-accent-foreground",
                collapsed && "justify-center px-2",
              )}
            >
              <item.icon className="h-4 w-4 shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          );

          if (collapsed) {
            return (
              <Tooltip key={item.href} content={item.label} side="right">
                {linkContent}
              </Tooltip>
            );
          }

          return linkContent;
        })}
      </nav>

      {/* Collapse Toggle */}
      <div className="border-t p-2">
        <Button
          variant="ghost"
          size="sm"
          className={cn("w-full", collapsed && "justify-center")}
          onClick={() => setCollapsed(!collapsed)}
        >
          {collapsed ? (
            <PanelLeft className="h-4 w-4" />
          ) : (
            <>
              <PanelLeftClose className="h-4 w-4" />
              <span className="ml-2">Collapse</span>
            </>
          )}
        </Button>
      </div>
    </aside>
  );
}
