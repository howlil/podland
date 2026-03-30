import { Cpu, MemoryStick, HardDrive } from "lucide-react";
import { VM } from "@/hooks/useVMs";

interface ResourceMetricsProps {
  vm?: VM;
  isLoading?: boolean;
}

export function ResourceMetrics({ vm, isLoading }: ResourceMetricsProps) {
  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-6 animate-pulse">
        <div className="h-6 w-40 bg-gray-200 dark:bg-gray-700 rounded mb-4" />
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="bg-gray-50 dark:bg-gray-700 rounded-xl p-4">
              <div className="h-4 w-20 bg-gray-200 dark:bg-gray-600 rounded mb-2" />
              <div className="h-8 w-32 bg-gray-200 dark:bg-gray-600 rounded" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!vm) return null;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-6">
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        <Cpu className="h-5 w-5 text-blue-600 dark:text-blue-400" />
        Resource Allocation
      </h2>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <ResourceCard
          icon={Cpu}
          label="CPU"
          value={`${vm.cpu} ${vm.cpu === 1 ? "Core" : "Cores"}`}
          color="blue"
        />
        <ResourceCard
          icon={MemoryStick}
          label="RAM"
          value={formatBytes(vm.ram)}
          color="purple"
        />
        <ResourceCard
          icon={HardDrive}
          label="Storage"
          value={formatBytes(vm.storage)}
          color="green"
        />
      </div>
    </div>
  );
}

interface ResourceCardProps {
  icon: React.ElementType;
  label: string;
  value: string;
  color: "blue" | "purple" | "green";
}

function ResourceCard({ icon: Icon, label, value, color }: ResourceCardProps) {
  const colorClasses = {
    blue: "from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-900/10 border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300",
    purple: "from-purple-50 to-purple-100 dark:from-purple-900/20 dark:to-purple-900/10 border-purple-200 dark:border-purple-800 text-purple-700 dark:text-purple-300",
    green: "from-green-50 to-green-100 dark:from-green-900/20 dark:to-green-900/10 border-green-200 dark:border-green-800 text-green-700 dark:text-green-300",
  };

  const iconColorClasses = {
    blue: "text-blue-600 dark:text-blue-400",
    purple: "text-purple-600 dark:text-purple-400",
    green: "text-green-600 dark:text-green-400",
  };

  return (
    <div className={`bg-gradient-to-br rounded-xl p-4 border ${colorClasses[color]}`}>
      <div className="flex items-center gap-2 mb-2">
        <Icon className={`h-5 w-5 ${iconColorClasses[color]}`} />
        <p className="text-sm font-medium">{label}</p>
      </div>
      <p className="text-2xl font-bold">{value}</p>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
}
