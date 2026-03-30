import { useState } from "react";
import { useDashboard } from "@/hooks/useDashboard";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { StatsGrid } from "@/components/dashboard/StatsGrid";
import { QuotaUsage } from "@/components/dashboard/QuotaUsage";
import { RecentVMs } from "@/components/dashboard/RecentVMs";
import { CreateVMWizard } from "@/components/vm/CreateVMWizard";
import { Plus } from "lucide-react";

export default function DashboardPage() {
  const [isWizardOpen, setIsWizardOpen] = useState(false);
  const { stats, quota, vms, isLoading } = useDashboard();

  return (
    <DashboardLayout>
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

        {/* Stats */}
        <StatsGrid stats={stats} isLoading={isLoading} />

        {/* Quota Usage */}
        <QuotaUsage quota={quota} isLoading={isLoading} />

        {/* Recent VMs */}
        <RecentVMs
          vms={vms}
          isLoading={isLoading}
          onCreateVM={() => setIsWizardOpen(true)}
        />

        {/* Create VM Wizard */}
        {isWizardOpen && (
          <CreateVMWizard
            onClose={() => setIsWizardOpen(false)}
            onSuccess={() => setIsWizardOpen(false)}
          />
        )}
      </div>
    </DashboardLayout>
  );
}
