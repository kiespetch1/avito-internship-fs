import { Tag } from "@/components/ui/tag";
import type { RunStatus } from "@/lib/api";
import { runStatusLabel } from "@/lib/format";

const variantByStatus: Record<RunStatus, "secondary" | "success" | "destructive"> = {
  pending: "secondary",
  success: "success",
  failed: "destructive",
};

export function RunStatusBadge({ status }: { status: RunStatus }) {
  return <Tag variant={variantByStatus[status]}>{runStatusLabel[status]}</Tag>;
}
