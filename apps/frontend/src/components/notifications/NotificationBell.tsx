import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../lib/api";
import { Bell, Check, CheckCheck, AlertTriangle, Info, AlertCircle } from "lucide-react";

interface Notification {
  id: string;
  vm_id: string;
  alert_name: string;
  severity: string;
  title: string;
  message: string;
  is_read: boolean;
  created_at: string;
  resolved_at?: string;
}

interface NotificationsResponse {
  data: Notification[];
}

export function NotificationBell() {
  const [isOpen, setIsOpen] = useState(false);
  const queryClient = useQueryClient();

  // Fetch unread count
  const { data: unreadData } = useQuery<{ data: { count: number } }>({
    queryKey: ["notifications-unread"],
    queryFn: () => api.get("/notifications/unread-count"),
    refetchInterval: 30000, // Poll every 30 seconds
  });

  // Fetch notifications
  const { data: notificationsData } = useQuery<{ data: NotificationsResponse }>({
    queryKey: ["notifications"],
    queryFn: () => api.get("/notifications?limit=10"),
    enabled: isOpen,
  });

  // Mark as read mutation
  const markAsReadMutation = useMutation({
    mutationFn: (notificationId: string) =>
      api.post(`/notifications/${notificationId}/read`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["notifications-unread"] });
    },
  });

  // Mark all as read mutation
  const markAllAsReadMutation = useMutation({
    mutationFn: () => api.post("/notifications/read-all"),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["notifications-unread"] });
    },
  });

  const unreadCount = unreadData?.data.count || 0;
  const notifications = notificationsData?.data.data || [];

  return (
    <div className="relative">
      {/* Bell Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white focus:outline-none"
      >
        <Bell className="h-5 w-5" />
        {unreadCount > 0 && (
          <span className="absolute top-0 right-0 h-4 w-4 bg-red-500 rounded-full text-xs text-white flex items-center justify-center">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </button>

      {/* Dropdown */}
      {isOpen && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />

          {/* Dropdown Content */}
          <div className="absolute right-0 mt-2 w-80 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-20">
            {/* Header */}
            <div className="flex items-center justify-between p-3 border-b border-gray-200 dark:border-gray-700">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white">
                Notifications
              </h3>
              {unreadCount > 0 && (
                <button
                  onClick={() => markAllAsReadMutation.mutate()}
                  className="text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 flex items-center gap-1"
                >
                  <CheckCheck className="h-3 w-3" />
                  Mark all read
                </button>
              )}
            </div>

            {/* Notifications List */}
            <div className="max-h-96 overflow-y-auto">
              {notifications.length === 0 ? (
                <div className="p-4 text-center text-sm text-gray-500 dark:text-gray-400">
                  No notifications
                </div>
              ) : (
                <div className="divide-y divide-gray-200 dark:divide-gray-700">
                  {notifications.map((notification) => (
                    <NotificationItem
                      key={notification.id}
                      notification={notification}
                      onMarkRead={() => markAsReadMutation.mutate(notification.id)}
                    />
                  ))}
                </div>
              )}
            </div>

            {/* Footer */}
            <div className="p-3 border-t border-gray-200 dark:border-gray-700">
              <a
                href="/dashboard/-vms"
                className="text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 text-center block"
              >
                View all notifications
              </a>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

interface NotificationItemProps {
  notification: Notification;
  onMarkRead: () => void;
}

function NotificationItem({ notification, onMarkRead }: NotificationItemProps) {
  const alertIcon = getAlertIcon(notification.alert_name);
  const severityColor = getSeverityColor(notification.severity);

  return (
    <div
      className={`p-3 hover:bg-gray-50 dark:hover:bg-gray-700/50 ${
        !notification.is_read ? "bg-blue-50/50 dark:bg-blue-900/10" : ""
      }`}
    >
      <div className="flex items-start gap-2">
        <div className={`mt-0.5 ${severityColor}`}>{alertIcon}</div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between gap-2">
            <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
              {notification.title}
            </p>
            {!notification.is_read && (
              <button
                onClick={onMarkRead}
                className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 flex-shrink-0"
                title="Mark as read"
              >
                <Check className="h-4 w-4" />
              </button>
            )}
          </div>
          <p className="text-xs text-gray-600 dark:text-gray-400 mt-1 line-clamp-2">
            {notification.message}
          </p>
          <div className="flex items-center gap-2 mt-2">
            <span className="text-xs text-gray-400">
              {formatTimeAgo(notification.created_at)}
            </span>
            {notification.resolved_at && (
              <span className="text-xs text-green-600 dark:text-green-400 flex items-center gap-1">
                <Check className="h-3 w-3" />
                Resolved
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function getAlertIcon(alertName: string) {
  if (alertName.includes("CPU")) {
    return <AlertTriangle className="h-4 w-4" />;
  }
  if (alertName.includes("Memory")) {
    return <AlertCircle className="h-4 w-4" />;
  }
  return <Info className="h-4 w-4" />;
}

function getSeverityColor(severity: string) {
  switch (severity) {
    case "critical":
      return "text-red-600 dark:text-red-400";
    case "warning":
      return "text-yellow-600 dark:text-yellow-400";
    default:
      return "text-blue-600 dark:text-blue-400";
  }
}

function formatTimeAgo(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return "Just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  return `${diffDays}d ago`;
}
