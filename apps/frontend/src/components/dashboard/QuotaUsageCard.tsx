interface QuotaUsageCardProps {
  usedCpu: number;
  maxCpu: number;
  usedRam: number;
  maxRam: number;
}

export function QuotaUsageCard({
  usedCpu,
  maxCpu,
  usedRam,
  maxRam,
}: QuotaUsageCardProps) {
  const cpuPercent = maxCpu > 0 ? (usedCpu / maxCpu) * 100 : 0;
  const ramPercent = maxRam > 0 ? (usedRam / maxRam) * 100 : 0;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Quota Usage
      </h2>

      <div className="space-y-4">
        {/* CPU */}
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              CPU
            </span>
            <span className="text-sm text-gray-600 dark:text-gray-400">
              {usedCpu.toFixed(2)} / {maxCpu} cores
            </span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5">
            <div
              className={`h-2.5 rounded-full transition-all ${
                cpuPercent > 90
                  ? "bg-red-600"
                  : cpuPercent > 70
                  ? "bg-yellow-500"
                  : "bg-green-500"
              }`}
              style={{ width: `${cpuPercent}%` }}
            />
          </div>
        </div>

        {/* RAM */}
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              RAM
            </span>
            <span className="text-sm text-gray-600 dark:text-gray-400">
              {(usedRam / 1024).toFixed(1)} / {(maxRam / 1024).toFixed(1)} GB
            </span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5">
            <div
              className={`h-2.5 rounded-full transition-all ${
                ramPercent > 90
                  ? "bg-red-600"
                  : ramPercent > 70
                  ? "bg-yellow-500"
                  : "bg-green-500"
              }`}
              style={{ width: `${ramPercent}%` }}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
