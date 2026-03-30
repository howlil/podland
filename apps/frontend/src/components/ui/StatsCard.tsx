interface StatsCardProps {
  icon: React.ElementType;
  label: string;
  value: string | number;
  color: string;
}

export function StatsCard({ icon: Icon, label, value, color }: StatsCardProps) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
      <div className="flex items-center justify-between mb-4">
        <div className={`p-3 rounded-xl bg-gradient-to-br ${color} text-white`}>
          <Icon className="h-6 w-6" />
        </div>
      </div>
      <div className="text-3xl font-bold text-gray-900 dark:text-white mb-1">
        {value}
      </div>
      <div className="text-sm text-gray-600 dark:text-gray-400">{label}</div>
    </div>
  );
}

interface QuotaBarProps {
  label: string;
  used: number;
  limit: number;
  unit?: string;
  formatBytes?: boolean;
}

export function QuotaBar({ label, used, limit, unit, formatBytes = false }: QuotaBarProps) {
  const percentage = Math.min((used / limit) * 100, 100);

  const formatValue = (val: number) => {
    if (formatBytes) {
      const k = 1024;
      const sizes = ["B", "KB", "MB", "GB", "TB"];
      const i = Math.floor(Math.log(val) / Math.log(k));
      return parseFloat((val / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
    }
    return val.toString();
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-2">
        <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
          {label}
        </span>
        <span className="text-sm text-gray-600 dark:text-gray-400">
          {formatValue(used)} / {formatValue(limit)}
          {unit}
        </span>
      </div>
      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5">
        <div
          className="bg-gradient-to-r from-blue-600 to-purple-600 h-2.5 rounded-full transition-all duration-500"
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
}

interface EmptyStateProps {
  icon: React.ElementType;
  title: string;
  description: string;
  action?: React.ReactNode;
}

export function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="text-center py-12">
      <Icon className="h-16 w-16 mx-auto mb-4 text-gray-400 dark:text-gray-600" />
      <p className="text-lg font-medium text-gray-900 dark:text-white mb-2">{title}</p>
      <p className="text-gray-600 dark:text-gray-400 mb-6">{description}</p>
      {action && <div className="flex justify-center">{action}</div>}
    </div>
  );
}
