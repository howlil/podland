import { createFileRoute } from "@tanstack/react-router";
import { ObservabilityPage } from "@/pages/ObservabilityPage";

export const Route = createFileRoute("/dashboard-observability")({
  component: ObservabilityPage,
});
