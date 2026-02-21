"use client";

import { useState, useRef, useCallback, useEffect } from "react";
import { X, GripHorizontal } from "lucide-react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useStore } from "@/store";
import { cn } from "@/lib/utils";
import type { OrderRequest, OrderType, OrderProduct } from "@/lib/types";

type ProductTab = "Regular" | "Cover" | "AMO";

export function OrderWindow() {
  const {
    isOrderWindowOpen,
    orderWindowTicker,
    orderWindowSide,
    isPlacingOrder,
    closeOrderWindow,
    submitOrder,
  } = useStore();

  const [exchange, setExchange] = useState<"NSE" | "BSE">("NSE");
  const [productTab, setProductTab] = useState<ProductTab>("Regular");
  const [product, setProduct] = useState<OrderProduct>("MIS");
  const [orderType, setOrderType] = useState<OrderType>("MARKET");
  const [quantity, setQuantity] = useState(1);
  const [price, setPrice] = useState(0);
  const [triggerPrice, setTriggerPrice] = useState(0);

  // Dragging state
  const [position, setPosition] = useState({ x: 0, y: 100 });
  const dragRef = useRef<HTMLDivElement>(null);
  const isDragging = useRef(false);
  const dragStart = useRef({ x: 0, y: 0 });

  // Reset form when opened with new ticker
  useEffect(() => {
    if (isOrderWindowOpen) {
      setExchange("NSE");
      setProductTab("Regular");
      setProduct("MIS");
      setOrderType("MARKET");
      setQuantity(1);
      setPrice(0);
      setTriggerPrice(0);
    }
  }, [isOrderWindowOpen, orderWindowTicker]);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    isDragging.current = true;
    dragStart.current = {
      x: e.clientX - position.x,
      y: e.clientY - position.y,
    };

    const handleMouseMove = (e: MouseEvent) => {
      if (isDragging.current) {
        setPosition({
          x: e.clientX - dragStart.current.x,
          y: e.clientY - dragStart.current.y,
        });
      }
    };

    const handleMouseUp = () => {
      isDragging.current = false;
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    };

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);
  }, [position]);

  const handleSubmit = async () => {
    const req: OrderRequest = {
      ticker: orderWindowTicker,
      exchange,
      side: orderWindowSide,
      order_type: orderType,
      product,
      quantity,
      ...(orderType === "LIMIT" || orderType === "SL" ? { price } : {}),
      ...(orderType === "SL" || orderType === "SL-M" ? { trigger_price: triggerPrice } : {}),
    };

    const result = await submitOrder(req);
    if (result) {
      closeOrderWindow();
    }
  };

  if (!isOrderWindowOpen) return null;

  const isBuy = orderWindowSide === "BUY";
  const accentColor = isBuy ? "blue" : "red";

  return (
    <div
      ref={dragRef}
      className="fixed z-50 shadow-2xl"
      style={{
        left: `${position.x}px`,
        top: `${position.y}px`,
        width: "420px",
      }}
    >
      <Card className="border-2" style={{ borderColor: isBuy ? "rgb(59, 130, 246)" : "rgb(239, 68, 68)" }}>
        {/* Header — Draggable */}
        <CardHeader
          className={cn(
            "flex flex-row items-center justify-between py-3 px-4 cursor-move select-none",
            isBuy ? "bg-blue-500/10" : "bg-red-500/10",
          )}
          onMouseDown={handleMouseDown}
        >
          <div className="flex items-center gap-2">
            <GripHorizontal size={14} className="text-muted-foreground" />
            <Badge variant={isBuy ? "default" : "destructive"} className="text-xs">
              {orderWindowSide}
            </Badge>
            <span className="font-mono font-bold text-sm">{orderWindowTicker}</span>
          </div>
          <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={closeOrderWindow}>
            <X size={14} />
          </Button>
        </CardHeader>

        <CardContent className="p-4 space-y-4">
          {/* Exchange Selector */}
          <div className="flex items-center gap-4">
            <span className="text-xs text-muted-foreground w-16">Exchange</span>
            <div className="flex gap-2">
              {(["NSE", "BSE"] as const).map((ex) => (
                <button
                  key={ex}
                  className={cn(
                    "px-3 py-1 text-xs rounded-md border transition-colors",
                    exchange === ex
                      ? isBuy
                        ? "bg-blue-500 text-white border-blue-500"
                        : "bg-red-500 text-white border-red-500"
                      : "border-border hover:bg-muted",
                  )}
                  onClick={() => setExchange(ex)}
                >
                  {ex}
                </button>
              ))}
            </div>
          </div>

          {/* Product Tabs */}
          <div className="flex border-b">
            {(["Regular", "Cover", "AMO"] as ProductTab[]).map((tab) => (
              <button
                key={tab}
                className={cn(
                  "px-4 py-2 text-xs font-medium transition-colors border-b-2",
                  productTab === tab
                    ? isBuy
                      ? "border-blue-500 text-blue-600"
                      : "border-red-500 text-red-600"
                    : "border-transparent text-muted-foreground hover:text-foreground",
                )}
                onClick={() => setProductTab(tab)}
              >
                {tab}
              </button>
            ))}
          </div>

          {/* Validity: Intraday / Longterm */}
          <div className="flex items-center gap-4">
            <span className="text-xs text-muted-foreground w-16">Validity</span>
            <div className="flex gap-2">
              <button
                className={cn(
                  "px-3 py-1 text-xs rounded-md border transition-colors",
                  product === "MIS"
                    ? isBuy
                      ? "bg-blue-500 text-white border-blue-500"
                      : "bg-red-500 text-white border-red-500"
                    : "border-border hover:bg-muted",
                )}
                onClick={() => setProduct("MIS")}
              >
                Intraday (MIS)
              </button>
              <button
                className={cn(
                  "px-3 py-1 text-xs rounded-md border transition-colors",
                  product === "CNC"
                    ? isBuy
                      ? "bg-blue-500 text-white border-blue-500"
                      : "bg-red-500 text-white border-red-500"
                    : "border-border hover:bg-muted",
                )}
                onClick={() => setProduct("CNC")}
              >
                Longterm (CNC)
              </button>
            </div>
          </div>

          {/* Quantity */}
          <div className="flex items-center gap-4">
            <label className="text-xs text-muted-foreground w-16">Qty</label>
            <Input
              type="number"
              min={1}
              value={quantity}
              onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value) || 1))}
              className="h-8 text-sm flex-1"
            />
          </div>

          {/* Price (for LIMIT / SL) */}
          <div className="flex items-center gap-4">
            <label className="text-xs text-muted-foreground w-16">Price</label>
            <Input
              type="number"
              step={0.05}
              value={price}
              onChange={(e) => setPrice(parseFloat(e.target.value) || 0)}
              disabled={orderType === "MARKET" || orderType === "SL-M"}
              className="h-8 text-sm flex-1"
            />
          </div>

          {/* Trigger Price (for SL / SL-M) */}
          <div className="flex items-center gap-4">
            <label className="text-xs text-muted-foreground w-16">Trigger</label>
            <Input
              type="number"
              step={0.05}
              value={triggerPrice}
              onChange={(e) => setTriggerPrice(parseFloat(e.target.value) || 0)}
              disabled={orderType === "MARKET" || orderType === "LIMIT"}
              className="h-8 text-sm flex-1"
            />
          </div>

          {/* Order Type */}
          <div className="flex items-center gap-4">
            <span className="text-xs text-muted-foreground w-16">Type</span>
            <div className="flex gap-1.5 flex-wrap">
              {(["MARKET", "LIMIT", "SL", "SL-M"] as OrderType[]).map((ot) => (
                <button
                  key={ot}
                  className={cn(
                    "px-2.5 py-1 text-xs rounded-md border transition-colors",
                    orderType === ot
                      ? isBuy
                        ? "bg-blue-500 text-white border-blue-500"
                        : "bg-red-500 text-white border-red-500"
                      : "border-border hover:bg-muted",
                  )}
                  onClick={() => setOrderType(ot)}
                >
                  {ot}
                </button>
              ))}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-3 pt-2">
            <Button
              className={cn(
                "flex-1",
                isBuy
                  ? "bg-blue-500 hover:bg-blue-600 text-white"
                  : "bg-red-500 hover:bg-red-600 text-white",
              )}
              onClick={handleSubmit}
              disabled={isPlacingOrder || quantity <= 0}
            >
              {isPlacingOrder ? "Placing..." : `${orderWindowSide} ${orderWindowTicker}`}
            </Button>
            <Button variant="outline" className="flex-1" onClick={closeOrderWindow}>
              Cancel
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
