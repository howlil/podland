import { createFileRoute } from "@tanstack/react-router";
import VMDetailPage from "@/pages/VMDetailPage";

export const Route = createFileRoute("/dashboard-vms-$id")({
  component: VMDetailPage,
});
