"use client";

import { useState } from "react";
import {
  Bell,
  BellOff,
  Plus,
  Trash2,
  ChevronUp,
  ChevronDown,
  AlertTriangle,
} from "lucide-react";
import { useStore } from "@/store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Select, SelectTrigger, SelectContent, SelectItem, SelectValue } from "@/components/ui/select";
import { createAlert, deleteAlert as deleteAlertApi } from "@/lib/api";
import { cn } from "@/lib/utils";
import type { Alert } from "@/lib/types";

export function AlertManager() {
  const { alerts, setAlerts } = useStore();
  const [showCreate, setShowCreate] = useState(false);
  const [newAlert, setNewAlert] = useState({
    query: "",
    condition: "gt" as "gt" | "lt" | "eq" | "gte" | "lte",
    threshold: "",
    name: "",
  });
  const [creating, setCreating] = useState(false);

  const conditionLabels: Record<string, string> = {
    gt: ">",
    lt: "<",
    eq: "=",
    gte: "≥",
    lte: "≤",
  };

  async function handleCreate() {
    if (!newAlert.query || !newAlert.threshold) return;
    setCreating(true);
    try {
      const expression = `${newAlert.query} ${newAlert.condition} ${newAlert.threshold}`;
      const alert = await createAlert(expression);
      // Attach local metadata for display
      alert.name = newAlert.name || newAlert.query;
      alert.query = newAlert.query;
      alert.condition = newAlert.condition;
      alert.threshold = parseFloat(newAlert.threshold);
      setAlerts([alert, ...alerts]);
      setShowCreate(false);
      setNewAlert({ query: "", condition: "gt", threshold: "", name: "" });
    } catch {
      // TODO: show toast
    } finally {
      setCreating(false);
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteAlertApi(id);
      setAlerts(alerts.filter((a) => a.id !== id));
    } catch {
      // TODO: show toast
    }
  }

  function alertStatusColor(status: Alert["status"]) {
    switch (status) {
      case "triggered":
        return "destructive";
      case "pending":
        return "default";
      case "expired":
        return "secondary";
      default:
        return "outline";
    }
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Bell size={18} />
            Alerts
          </CardTitle>
          <Button size="sm" onClick={() => setShowCreate(true)}>
            <Plus size={14} className="mr-1" />
            New Alert
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {alerts.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <BellOff size={32} className="mx-auto mb-2 opacity-50" />
            <p className="text-sm">No alerts configured</p>
            <p className="text-xs mt-1">
              Create alerts on FinanceQL queries to get notified
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {alerts.map((alert) => (
              <div
                key={alert.id}
                className={cn(
                  "flex items-center justify-between p-3 rounded-md border",
                  alert.status === "triggered" && "border-destructive/50 bg-destructive/5"
                )}
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-sm font-medium truncate">
                      {alert.name}
                    </span>
                    <Badge variant={alertStatusColor(alert.status)}>{alert.status}</Badge>
                  </div>
                  <code className="text-xs text-muted-foreground font-mono block truncate">
                    {alert.query} {alert.condition ? (conditionLabels[alert.condition] ?? alert.condition) : ""}{" "}
                    {alert.threshold}
                  </code>
                  {alert.status === "triggered" && alert.triggeredAt && (
                    <div className="flex items-center gap-1 mt-1 text-xs text-destructive">
                      <AlertTriangle size={12} />
                      Triggered at {new Date(alert.triggeredAt).toLocaleTimeString("en-IN")}
                      {alert.value !== undefined && (
                        <> — Current: {alert.value.toLocaleString("en-IN")}</>
                      )}
                    </div>
                  )}
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-muted-foreground hover:text-destructive shrink-0 ml-2"
                  onClick={() => handleDelete(alert.id)}
                >
                  <Trash2 size={14} />
                </Button>
              </div>
            ))}
          </div>
        )}
      </CardContent>

      {/* Create Alert Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Alert</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div>
              <label className="text-sm font-medium mb-1 block">Name (optional)</label>
              <Input
                placeholder="My alert"
                value={newAlert.name}
                onChange={(e) => setNewAlert({ ...newAlert, name: e.target.value })}
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-1 block">FinanceQL Query</label>
              <Input
                placeholder="close(RELIANCE)"
                value={newAlert.query}
                onChange={(e) => setNewAlert({ ...newAlert, query: e.target.value })}
                className="font-mono"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium mb-1 block">Condition</label>
                <Select
                  value={newAlert.condition}
                  onValueChange={(v) =>
                    setNewAlert({ ...newAlert, condition: v as typeof newAlert.condition })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="gt">Greater than (&gt;)</SelectItem>
                    <SelectItem value="gte">Greater or equal (≥)</SelectItem>
                    <SelectItem value="lt">Less than (&lt;)</SelectItem>
                    <SelectItem value="lte">Less or equal (≤)</SelectItem>
                    <SelectItem value="eq">Equal (=)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <label className="text-sm font-medium mb-1 block">Threshold</label>
                <Input
                  type="number"
                  placeholder="2500"
                  value={newAlert.threshold}
                  onChange={(e) => setNewAlert({ ...newAlert, threshold: e.target.value })}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={creating || !newAlert.query || !newAlert.threshold}>
              {creating ? "Creating…" : "Create Alert"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
