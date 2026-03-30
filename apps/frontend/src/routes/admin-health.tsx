import { createFileRoute } from "@tanstack/react-router";
import AdminHealthPage from "@/pages/AdminHealthPage";

export const Route = createFileRoute("/admin-health")({
  component: AdminHealthPage,
});
