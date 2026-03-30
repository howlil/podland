import { useParams } from "@tanstack/react-router";
import { useVM } from "@/hooks/useVMs";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { VMHeader } from "@/components/vm/VMHeader";
import { ResourceMetrics } from "@/components/vm/ResourceMetrics";
import { ConnectionInfo } from "@/components/vm/ConnectionInfo";
import { VMActions } from "@/components/vm/VMActions";
import { toast } from "sonner";

export default function VMDetailPage() {
  const { id } = useParams({ from: "/dashboard/vms/$id" });
  const {
    vm,
    isLoading,
    startVM,
    stopVM,
    restartVM,
    deleteVM,
    pinVM,
    unpinVM,
    isStarting,
    isStopping,
    isRestarting,
    isDeleting,
    isPinning,
    isUnpinning,
  } = useVM(id);

  const handleDownloadSSHKey = () => {
    toast.info("SSH key was shown only during VM creation", {
      description: "You'll need to recreate the VM if you didn't save it.",
    });
  };

  return (
    <DashboardLayout>
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <VMHeader
          vm={vm}
          isLoading={isLoading}
          onPin={pinVM}
          onUnpin={unpinVM}
          isPinning={isPinning}
          isUnpinning={isUnpinning}
        />

        {/* Resource Metrics */}
        <ResourceMetrics vm={vm} isLoading={isLoading} />

        {/* Connection Info */}
        <ConnectionInfo
          vm={vm}
          isLoading={isLoading}
          onDownloadSSHKey={handleDownloadSSHKey}
        />

        {/* VM Actions */}
        <VMActions
          vm={vm}
          isLoading={isLoading}
          onStart={startVM}
          onStop={stopVM}
          onRestart={restartVM}
          onDelete={deleteVM}
          isStarting={isStarting}
          isStopping={isStopping}
          isRestarting={isRestarting}
          isDeleting={isDeleting}
        />
      </div>
    </DashboardLayout>
  );
}
