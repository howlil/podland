interface Quota {
  cpu_limit: number;
  cpu_used: number;
  ram_limit: number;
  ram_used: number;
  storage_limit: number;
  storage_used: number;
  vm_count: number;
  vm_count_limit: number;
}

interface QuotaUsageProps {
  quota?: Quota;
  isLoading?: boolean;
}

export function QuotaUsage({ quota, isLoading }: QuotaUsageProps) {
  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-8 animate-pulse">
        <div className="h-6 w-32 bg-gray-200 dark:bg-gray-700 rounded mb-6" />
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="space-y-2">
              <div className="h-4 w-24 bg-gray-200 dark:bg-gray-700 rounded" />
              <div className="h-2.5 w-full bg-gray-200 dark:bg-gray-700 rounded-full" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!quota) return null;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-8">
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Quota Usage
      </h2>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
        <QuotaBar
          label="CPU"
          used={quota.cpu_used}
          limit={quota.cpu_limit}
          unit="cores"
        />
        <QuotaBar
          label="RAM"
          used={quota.ram_used}
          limit={quota.ram_limit}
          formatBytes
        />
        <QuotaBar
          label="VMs"
          used={quota.vm_count}
          limit={quota.vm_count_limit}
          unit="VMs"
        />
      </div>
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

function QuotaBar({ label, used, limit, unit, formatBytes = false }: QuotaBarProps) {
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
