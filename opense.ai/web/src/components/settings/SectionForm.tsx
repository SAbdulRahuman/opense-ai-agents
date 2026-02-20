// ============================================================================
// OpeNSE.ai — Settings Section Form (reusable wrapper)
// ============================================================================

"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

interface SectionFormProps {
  title: string;
  description: string;
  children: React.ReactNode;
}

export function SectionForm({ title, description, children }: SectionFormProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">{children}</div>
      </CardContent>
    </Card>
  );
}
