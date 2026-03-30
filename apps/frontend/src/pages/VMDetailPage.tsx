import { useParams } from "@tanstack/react-router";
import { useVM } from "@/hooks/useVMs";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { VMHeader } from "@/components/vm/VMHeader";
import { ResourceMetrics } from "@/components/vm/ResourceMetrics";
import { ConnectionInfo } from "@/components/vm/ConnectionInfo";
import { VMActions, getDefaultVMActions } from "@/components/vm/VMActions";
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

  const actions = getDefaultVMActions({
    onStart: () => startVM(),
    onStop: () => stopVM(),
    onRestart: () => restartVM(),
    onDelete: () => deleteVM(),
    isStarting,
    isStopping,
    isRestarting,
    isDeleting,
  });

  return (
    <DashboardLayout>
      <div className="max-w-4xl mx-auto">
        <VMHeader
          vm={vm}
          isLoading={isLoading}
          onPin={pinVM}
          onUnpin={unpinVM}
          isPinning={isPinning}
          isUnpinning={isUnpinning}
        />

        <ResourceMetrics vm={vm} isLoading={isLoading} />

        <ConnectionInfo
          vm={vm}
          isLoading={isLoading}
          onDownloadSSHKey={handleDownloadSSHKey}
        />

        <VMActions
          vm={vm}
          isLoading={isLoading}
          actions={actions}
        />
      </div>
    </DashboardLayout>
  );
}
