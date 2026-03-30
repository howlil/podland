import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";
import { Server, Zap, Globe, HardDrive, ArrowRight, Plus } from "lucide-react";
import { CreateVMWizard } from "@/components/vm/CreateVMWizard";
import { useState } from "react";

export const Route = createFileRoute("/dashboard")({
  component: DashboardIndex,
});

interface VM {
  id: string;
  name: string;
  status: string;
  cpu: number;
  ram: number;
  storage: number;
  domain?: string;
  os?: string;
  created_at: string;
}

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

function DashboardIndex() {
  const [isWizardOpen, setIsWizardOpen] = useState(false);

  const { data: vms = [] } = useQuery<VM[]>({
    queryKey: ["vms"],
    queryFn: async () => {
      const { data } = await api.get("/vms");
      return data;
    },
    refetchInterval: 10000,
  });

  const { data: quota } = useQuery<Quota>({
    queryKey: ["quota"],
    queryFn: async () => {
      const { data } = await api.get("/users/me");
      // Calculate from user data
      return {
        cpu_limit: data.role === "internal" ? 4.0 : 0.5,
        cpu_used: 0,
        ram_limit: data.role === "internal" ? 8589934592 : 1073741824,
        ram_used: 0,
        storage_limit: data.role === "internal" ? 107374182400 : 10737418240,
        storage_used: 0,
        vm_count: vms.length,
        vm_count_limit: data.role === "internal" ? 5 : 2,
      };
    },
  });

  const runningVMs = vms.filter((vm) => vm.status === "running").length;
  const totalVMs = vms.length;

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  return (
    <div className="max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Dashboard
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Welcome back! Here's what's happening with your VMs.
          </p>
        </div>
        <button
          onClick={() => setIsWizardOpen(true)}
          className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white rounded-xl font-semibold shadow-lg hover:shadow-xl transition-all transform hover:scale-105"
        >
          <Plus className="h-5 w-5" />
          Create VM
        </button>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <StatCard
          icon={Server}
          label="Total VMs"
          value={totalVMs.toString()}
          color="from-blue-500 to-cyan-500"
        />
        <StatCard
          icon={Zap}
          label="Running"
          value={runningVMs.toString()}
          color="from-green-500 to-emerald-500"
        />
        <StatCard
          icon={HardDrive}
          label="Storage Used"
          value={quota ? formatBytes(quota.storage_used) : "-"}
          color="from-purple-500 to-pink-500"
        />
        <StatCard
          icon={Globe}
          label="Domains"
          value={vms.filter((vm) => vm.domain).length.toString()}
          color="from-yellow-500 to-orange-500"
        />
      </div>

      {/* Quota Usage */}
      {quota && (
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
      )}

      {/* Recent VMs */}
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Recent VMs
          </h2>
          <a
            href="/dashboard/-vms"
            className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 font-medium"
          >
            View All
            <ArrowRight className="h-4 w-4" />
          </a>
        </div>
        {vms.length === 0 ? (
          <div className="text-center py-12">
            <Server className="h-16 w-16 mx-auto mb-4 text-gray-400 dark:text-gray-600" />
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              No VMs yet. Create your first VM to get started!
            </p>
            <button
              onClick={() => setIsWizardOpen(true)}
              className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white rounded-xl font-semibold shadow-lg hover:shadow-xl transition-all"
            >
              <Plus className="h-5 w-5" />
              Create Your First VM
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {vms.slice(0, 5).map((vm) => (
              <VMRow key={vm.id} vm={vm} />
            ))}
          </div>
        )}
      </div>

      {/* Create VM Wizard */}
      {isWizardOpen && (
        <CreateVMWizard
          onClose={() => setIsWizardOpen(false)}
          onSuccess={() => {
            setIsWizardOpen(false);
          }}
        />
      )}
    </div>
  );
}

function StatCard({
  icon: Icon,
  label,
  value,
  color,
}: {
  icon: React.ElementType;
  label: string;
  value: string;
  color: string;
}) {
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

function QuotaBar({
  label,
  used,
  limit,
  unit,
  formatBytes = false,
}: {
  label: string;
  used: number;
  limit: number;
  unit?: string;
  formatBytes?: boolean;
}) {
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

function VMRow({ vm }: { vm: VM }) {
  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
      case "stopped":
        return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300";
      case "pending":
        return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
      default:
        return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300";
    }
  };

  return (
    <a
      href={`/dashboard/-vms/${vm.id}`}
      className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors group"
    >
      <div className="flex items-center gap-4">
        <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
          <Server className="h-5 w-5 text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <p className="font-medium text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-400">
            {vm.name}
          </p>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            {vm.os === "ubuntu-2204" ? "Ubuntu 22.04" : "Debian 12"}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <span
          className={`px-3 py-1 rounded-full text-xs font-semibold ${getStatusColor(
            vm.status
          )}`}
        >
          {vm.status}
        </span>
        <ArrowRight className="h-4 w-4 text-gray-400 group-hover:text-blue-600 dark:group-hover:text-blue-400" />
      </div>
    </a>
  );
}
