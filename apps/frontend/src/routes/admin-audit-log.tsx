import { createFileRoute } from "@tanstack/react-router";
import { AdminAuditLogPage } from "@/pages/AdminAuditLogPage";

export const Route = createFileRoute("/admin-audit-log")({
  component: AdminAuditLogPage,
});
