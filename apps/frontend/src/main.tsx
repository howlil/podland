import { createRoot } from "react-dom/client";
import { RouterProvider } from "@tanstack/react-router";
import { getRouter } from "./router";
import { initAnalytics } from "./lib/analytics";
import "./styles.css";

// Initialize analytics tracking
initAnalytics();

const router = getRouter();

createRoot(document.getElementById("root")!).render(
  <RouterProvider router={router} />
);
