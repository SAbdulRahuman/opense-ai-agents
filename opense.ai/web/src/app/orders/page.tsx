"use client";

import { useEffect, useState } from "react";
import {
  ClipboardList,
  Clock,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  RefreshCw,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { useStore } from "@/store";
import { formatPrice, cn } from "@/lib/utils";
import type { Order, OrderStatus } from "@/lib/types";

const statusConfig: Record<OrderStatus, { label: string; variant: "default" | "secondary" | "destructive" | "outline"; icon: typeof Clock }> = {
  PENDING: { label: "Pending", variant: "outline", icon: Clock },
  OPEN: { label: "Open", variant: "default", icon: AlertTriangle },
  COMPLETE: { label: "Executed", variant: "secondary", icon: CheckCircle2 },
  CANCELLED: { label: "Cancelled", variant: "destructive", icon: XCircle },
  REJECTED: { label: "Rejected", variant: "destructive", icon: XCircle },
};

function OrdersTable({ orders, onCancel }: { orders: Order[]; onCancel: (id: string) => void }) {
  if (orders.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <ClipboardList size={48} className="mb-4 opacity-50" />
        <p className="text-lg font-medium">No orders placed</p>
        <p className="text-sm mt-1">Your orders will appear here once placed</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b text-left">
            <th className="pb-3 font-medium">Time</th>
            <th className="pb-3 font-medium">Type</th>
            <th className="pb-3 font-medium">Instrument</th>
            <th className="pb-3 font-medium">Product</th>
            <th className="pb-3 font-medium text-right">Qty</th>
            <th className="pb-3 font-medium text-right">Price</th>
            <th className="pb-3 font-medium">Status</th>
            <th className="pb-3 font-medium text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {orders.map((order) => {
            const sc = statusConfig[order.status];
            const StatusIcon = sc.icon;
            const canCancel = order.status === "PENDING" || order.status === "OPEN";

            return (
              <tr key={order.order_id} className="border-b last:border-0">
                <td className="py-3 text-xs text-muted-foreground">
                  {new Date(order.placed_at).toLocaleTimeString("en-IN", {
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </td>
                <td className="py-3">
                  <Badge
                    variant={order.side === "BUY" ? "default" : "destructive"}
                    className="text-xs"
                  >
                    {order.side}
                  </Badge>
                </td>
                <td className="py-3">
                  <div className="font-mono font-medium">{order.ticker}</div>
                  <div className="text-xs text-muted-foreground">
                    {order.exchange} · {order.order_type}
                  </div>
                </td>
                <td className="py-3 text-xs">{order.product}</td>
                <td className="py-3 text-right tabular-nums">
                  {order.filled_qty}/{order.quantity}
                </td>
                <td className="py-3 text-right tabular-nums">
                  {formatPrice(order.avg_price || order.price)}
                </td>
                <td className="py-3">
                  <Badge variant={sc.variant} className="text-xs gap-1">
                    <StatusIcon size={10} />
                    {sc.label}
                  </Badge>
                </td>
                <td className="py-3 text-right">
                  {canCancel && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs text-red-500 hover:text-red-600"
                      onClick={() => onCancel(order.order_id)}
                    >
                      Cancel
                    </Button>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

export default function OrdersPage() {
  const { orders, ordersLoading, fetchOrders, cancelOrderById } = useStore();
  const [filter, setFilter] = useState<"all" | "open" | "executed">("all");

  useEffect(() => {
    fetchOrders();
    const interval = setInterval(fetchOrders, 5000);
    return () => clearInterval(interval);
  }, [fetchOrders]);

  const filteredOrders = orders.filter((o) => {
    if (filter === "open") return o.status === "PENDING" || o.status === "OPEN";
    if (filter === "executed") return o.status === "COMPLETE";
    return true;
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Orders</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Track and manage your placed orders
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={fetchOrders} disabled={ordersLoading}>
          <RefreshCw size={14} className={cn(ordersLoading && "animate-spin")} />
          <span className="ml-1.5">Refresh</span>
        </Button>
      </div>

      <Card>
        <CardHeader className="pb-3">
          <Tabs value={filter} onValueChange={(v) => setFilter(v as typeof filter)}>
            <TabsList>
              <TabsTrigger value="all">All ({orders.length})</TabsTrigger>
              <TabsTrigger value="open">
                Open ({orders.filter((o) => o.status === "PENDING" || o.status === "OPEN").length})
              </TabsTrigger>
              <TabsTrigger value="executed">
                Executed ({orders.filter((o) => o.status === "COMPLETE").length})
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </CardHeader>
        <CardContent>
          {ordersLoading && orders.length === 0 ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : (
            <OrdersTable orders={filteredOrders} onCancel={cancelOrderById} />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
