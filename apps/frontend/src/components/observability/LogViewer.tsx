import { useState, useEffect, useRef } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "../../lib/api";
import { Play, Square, RefreshCw } from "lucide-react";

interface LogEntry {
  timestamp: string;
  line: string;
  labels?: Record<string, string>;
}

interface LogsResponse {
  entries: LogEntry[];
  total: number;
}

interface LogViewerProps {
  vmId: string;
  limit?: number;
  showLiveTail?: boolean;
  height?: string;
}

export function LogViewer({ vmId, limit = 1000, showLiveTail = false, height = "400px" }: LogViewerProps) {
  const [liveTail, setLiveTail] = useState(showLiveTail);
  const [levelFilter, setLevelFilter] = useState<string>("");
  const logsEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const { data, isLoading, error, refetch } = useQuery<{ data: LogsResponse }>({
    queryKey: ["logs", vmId, limit, levelFilter],
    queryFn: () => {
      const params = new URLSearchParams({ limit: limit.toString() });
      if (levelFilter) params.append("level", levelFilter);
      return api.get(`/vms/${vmId}/logs?${params}`);
    },
    refetchInterval: liveTail ? 5000 : false,
    enabled: !liveTail,
  });

  // WebSocket for live tail
  useEffect(() => {
    if (liveTail) {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${protocol}//${window.location.host}/api/vms/${vmId}/logs/stream`;
      
      wsRef.current = new WebSocket(wsUrl);

      wsRef.current.onmessage = (event) => {
        const message = JSON.parse(event.data);
        if (message.type === "update" && message.entries) {
          // New logs arrived - trigger a refetch
          refetch();
        }
      };

      wsRef.current.onerror = (error) => {
        console.error("WebSocket error:", error);
        setLiveTail(false);
      };

      wsRef.current.onclose = () => {
        wsRef.current = null;
      };

      return () => {
        if (wsRef.current) {
          wsRef.current.close();
        }
      };
    }
  }, [liveTail, vmId, refetch]);

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (data?.data.entries) {
      logsEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [data?.data.entries]);

  const entries = data?.data.entries || [];

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      {/* Toolbar */}
      <div className="flex items-center justify-between p-3 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white">VM Logs</h3>
          <span className="text-xs text-gray-500 dark:text-gray-400">
            {entries.length} entries
          </span>
        </div>
        <div className="flex items-center gap-2">
          {/* Level Filter */}
          <div className="relative">
            <select
              value={levelFilter}
              onChange={(e) => setLevelFilter(e.target.value)}
              className="text-xs bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All Levels</option>
              <option value="ERROR">ERROR</option>
              <option value="WARN">WARN</option>
              <option value="INFO">INFO</option>
              <option value="DEBUG">DEBUG</option>
            </select>
          </div>

          {/* Live Tail Toggle */}
          {showLiveTail && (
            <button
              onClick={() => setLiveTail(!liveTail)}
              className={`flex items-center gap-1 px-2 py-1 text-xs rounded border ${
                liveTail
                  ? "bg-red-100 dark:bg-red-900/20 text-red-600 dark:text-red-400 border-red-300 dark:border-red-700"
                  : "bg-green-100 dark:bg-green-900/20 text-green-600 dark:text-green-400 border-green-300 dark:border-green-700"
              }`}
            >
              {liveTail ? (
                <>
                  <Square className="h-3 w-3" />
                  Stop Live
                </>
              ) : (
                <>
                  <Play className="h-3 w-3" />
                  Live Tail
                </>
              )}
            </button>
          )}

          {/* Refresh Button */}
          <button
            onClick={() => refetch()}
            className="p-1 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            title="Refresh logs"
          >
            <RefreshCw className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Logs Content */}
      <div
        className="font-mono text-xs overflow-y-auto p-3"
        style={{ height }}
      >
        {isLoading && !liveTail ? (
          <div className="text-gray-500 dark:text-gray-400">Loading logs...</div>
        ) : error ? (
          <div className="text-red-500 dark:text-red-400">Failed to load logs</div>
        ) : entries.length === 0 ? (
          <div className="text-gray-500 dark:text-gray-400">No logs found</div>
        ) : (
          <div className="space-y-1">
            {entries.map((entry, index) => (
              <LogLine key={index} entry={entry} />
            ))}
            <div ref={logsEndRef} />
          </div>
        )}
      </div>
    </div>
  );
}

interface LogLineProps {
  entry: LogEntry;
}

function LogLine({ entry }: LogLineProps) {
  const timestamp = new Date(entry.timestamp).toLocaleTimeString();
  
  // Try to detect log level from the line
  const level = detectLogLevel(entry.line);
  
  const levelColors: Record<string, string> = {
    ERROR: "text-red-600 dark:text-red-400",
    WARN: "text-yellow-600 dark:text-yellow-400",
    INFO: "text-green-600 dark:text-green-400",
    DEBUG: "text-gray-500 dark:text-gray-400",
  };

  const colorClass = level ? levelColors[level] : "text-gray-700 dark:text-gray-300";

  return (
    <div className="flex gap-2 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded px-1 py-0.5">
      <span className="text-gray-400 dark:text-gray-500 whitespace-nowrap">{timestamp}</span>
      {level && (
        <span className={`font-semibold whitespace-nowrap ${colorClass}`}>
          {level}
        </span>
      )}
      <span className={`${colorClass} break-all`}>{entry.line}</span>
    </div>
  );
}

function detectLogLevel(line: string): string | null {
  const upperLine = line.toUpperCase();
  if (upperLine.includes("ERROR") || upperLine.includes("FATAL") || upperLine.includes("CRIT")) {
    return "ERROR";
  }
  if (upperLine.includes("WARN")) {
    return "WARN";
  }
  if (upperLine.includes("INFO")) {
    return "INFO";
  }
  if (upperLine.includes("DEBUG") || upperLine.includes("TRACE")) {
    return "DEBUG";
  }
  return null;
}
