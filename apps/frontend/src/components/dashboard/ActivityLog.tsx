interface Activity {
  id: string;
  action: string;
  createdAt: string;
  metadata?: Record<string, unknown>;
}

interface ActivityLogProps {
  activities?: Activity[];
  className?: string;
}

export function ActivityLog({ activities, className }: ActivityLogProps) {
  const formatActivityText = (action: string): string => {
    const map: Record<string, string> = {
      account_created: "Account created",
      signed_in: "Signed in",
      signed_out: "Signed out",
      nim_confirmed: "NIM confirmed",
    };
    return map[action] || action;
  };

  const formatDistance = (date: Date, now: Date): string => {
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
    
    if (seconds < 60) return "Just now";
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h`;
    const days = Math.floor(hours / 24);
    return `${days}d`;
  };

  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg shadow p-6 ${className}`}>
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Recent Activity
      </h2>

      <div className="space-y-3">
        {activities?.map((activity) => (
          <div key={activity.id} className="flex items-start gap-3">
            <div className="w-2 h-2 mt-2 bg-gray-400 rounded-full" />
            <div className="flex-1">
              <p className="text-sm text-gray-900 dark:text-white">
                {formatActivityText(activity.action)}
              </p>
              <p className="text-xs text-gray-500 dark:text-gray-400">
                {formatDistance(
                  new Date(activity.createdAt),
                  new Date()
                )}{" "}
                ago
              </p>
            </div>
          </div>
        ))}

        {activities?.length === 0 && (
          <p className="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
            No recent activity
          </p>
        )}

        {!activities && (
          <div className="flex items-center justify-center py-4">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
          </div>
        )}
      </div>
    </div>
  );
}
