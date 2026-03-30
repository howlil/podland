import { createFileRoute } from "@tanstack/react-router";
import VMsPage from "@/pages/VMsPage";

export const Route = createFileRoute("/dashboard-vms")({
  component: VMsPage,
});
