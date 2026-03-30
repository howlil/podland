import { useState, useRef, useEffect } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import api from "@/lib/api";
import { useAuth } from "@/lib/auth";

interface Tier {
  name: string;
  cpu: number;
  ram: number;
  storage: number;
  min_role: "external" | "internal";
}

interface Quota {
  cpu_limit: number;
  ram_limit: number;
  storage_limit: number;
  vm_count_limit: number;
  cpu_used: number;
  ram_used: number;
  storage_used: number;
  vm_count: number;
}

interface CreateVMWizardProps {
  onClose: () => void;
  onSuccess: () => void;
}

interface CreateVMResponse {
  id: string;
  status: string;
  ssh_key: string;
  message: string;
}

export function CreateVMWizard({ onClose, onSuccess }: CreateVMWizardProps) {
  const { user } = useAuth();
  const [step, setStep] = useState<1 | 2 | 3 | 4>(1);
  const [vmName, setVmName] = useState("");
  const [selectedTier, setSelectedTier] = useState<string>("");
  const [selectedOS] = useState("ubuntu-2204");
  const [createdVM, setCreatedVM] = useState<CreateVMResponse | null>(null);
  const [showSSHKey, setShowSSHKey] = useState(false);
  
  // Accessibility: Focus management
  const stepContentRef = useRef<HTMLDivElement>(null);
  const close_button_ref = useRef<HTMLButtonElement>(null);
  
  // Focus management: Move focus to step heading when step changes
  useEffect(() => {
    if (stepContentRef.current) {
      const heading = stepContentRef.current.querySelector('h3');
      if (heading) {
        heading.setAttribute('tabindex', '-1');
        heading.focus();
      }
    }
  }, [step]);
  
  // Keyboard navigation: Close on Escape
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    document.addEventListener('keydown', handleEscape as EventListener);
    return () => document.removeEventListener('keydown', handleEscape as EventListener);
  }, [onClose]);
  
  // Focus trap: Keep focus within modal
  useEffect(() => {
    const modal = stepContentRef.current?.closest('[role="dialog"]');
    if (modal) {
      const focusableElements = modal.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      const firstElement = focusableElements[0] as HTMLElement;
      const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;
      
      const handleTabKey = (e: KeyboardEvent) => {
        if (e.key !== 'Tab') return;

        if (e.shiftKey) {
          if (document.activeElement === firstElement) {
            e.preventDefault();
            lastElement.focus();
          }
        } else {
          if (document.activeElement === lastElement) {
            e.preventDefault();
            firstElement.focus();
          }
        }
      };

      modal.addEventListener('keydown', handleTabKey as EventListener);
      return () => modal.removeEventListener('keydown', handleTabKey as EventListener);
    }
  }, []);

  const { data: tiers = [] } = useQuery<Tier[]>({
    queryKey: ["tiers"],
    queryFn: async () => {
      // For now, return hardcoded tiers
      // In production, fetch from API
      return [
        { name: "nano", cpu: 0.25, ram: 536870912, storage: 5368709120, min_role: "external" },
        { name: "micro", cpu: 0.5, ram: 1073741824, storage: 10737418240, min_role: "external" },
        { name: "small", cpu: 1.0, ram: 2147483648, storage: 21474836480, min_role: "internal" },
        { name: "medium", cpu: 2.0, ram: 4294967296, storage: 42949672960, min_role: "internal" },
        { name: "large", cpu: 4.0, ram: 8589934592, storage: 85899345920, min_role: "internal" },
        { name: "xlarge", cpu: 4.0, ram: 8589934592, storage: 107374182400, min_role: "internal" },
      ];
    },
  });

  const { data: quota } = useQuery<Quota>({
    queryKey: ["quota"],
    queryFn: async () => {
      await api.get("/users/me");
      // Calculate quota from user data
      // In production, fetch from dedicated quota endpoint
      return {
        cpu_limit: user?.role === "internal" ? 4.0 : 0.5,
        ram_limit: user?.role === "internal" ? 8589934592 : 1073741824,
        storage_limit: user?.role === "internal" ? 107374182400 : 10737418240,
        vm_count_limit: user?.role === "internal" ? 5 : 2,
        cpu_used: 0,
        ram_used: 0,
        storage_used: 0,
        vm_count: 0,
      };
    },
  });

  const createMutation = useMutation({
    mutationFn: async () => {
      const { data } = await api.post("/vms", {
        name: vmName,
        os: selectedOS,
        tier: selectedTier,
      });
      return data as CreateVMResponse;
    },
    onSuccess: (data) => {
      setCreatedVM(data);
      setShowSSHKey(true);
      setStep(4);
    },
  });

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const checkQuota = (tier: Tier): { valid: boolean; message: string } => {
    if (!quota) return { valid: true, message: "" };

    const newCpu = quota.cpu_used + tier.cpu;
    const newRam = quota.ram_used + tier.ram;
    const newStorage = quota.storage_used + tier.storage;
    const newVmCount = quota.vm_count + 1;

    if (newCpu > quota.cpu_limit) {
      return { valid: false, message: `CPU quota exceeded (${newCpu.toFixed(2)} / ${quota.cpu_limit} cores)` };
    }
    if (newRam > quota.ram_limit) {
      return { valid: false, message: `RAM quota exceeded (${formatBytes(newRam)} / ${formatBytes(quota.ram_limit)})` };
    }
    if (newStorage > quota.storage_limit) {
      return { valid: false, message: `Storage quota exceeded (${formatBytes(newStorage)} / ${formatBytes(quota.storage_limit)})` };
    }
    if (newVmCount > quota.vm_count_limit) {
      return { valid: false, message: `VM count quota exceeded (${newVmCount} / ${quota.vm_count_limit} VMs)` };
    }

    return { valid: true, message: "" };
  };

  const handleNext = () => {
    if (step === 1 && vmName.trim()) {
      setStep(2);
    } else if (step === 2 && selectedTier) {
      setStep(3);
    } else if (step === 3) {
      createMutation.mutate();
    }
  };

  const handleBack = () => {
    if (step > 1 && step < 4) {
      setStep((step - 1) as 1 | 2 | 3);
    }
  };

  const handleDownloadSSHKey = () => {
    if (!createdVM?.ssh_key) return;

    const blob = new Blob([createdVM.ssh_key], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${vmName}-ssh-key.pem`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const isTierAvailable = (tier: Tier) => {
    if (!user) return false;
    return tier.min_role === "external" || user.role === "internal";
  };

  const stepTitles = ["Name Your VM", "Choose a Tier", "Review & Create", "VM Created!"];
  
  const renderStepIndicator = () => (
    <div className="flex items-center justify-center mb-6" role="group" aria-label="Progress steps">
      {[1, 2, 3, 4].map((s) => (
        <div key={s} className="flex items-center">
          <div
            className={`w-8 h-8 rounded-full flex items-center justify-center font-medium ${
              s <= step
                ? "bg-primary text-white"
                : "bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400"
            }`}
            aria-current={s === step ? "step" : undefined}
            aria-label={`Step ${s}: ${stepTitles[s - 1]}`}
          >
            {s < 4 ? s : "✓"}
          </div>
          {s < 4 && (
            <div
              className={`w-16 h-1 ${
                s < step ? "bg-primary" : "bg-gray-200 dark:bg-gray-700"
              }`}
              aria-hidden="true"
            />
          )}
        </div>
      ))}
    </div>
  );

  const renderStep1 = () => (
    <div>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Name Your VM
      </h3>
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          VM Name
        </label>
        <input
          type="text"
          value={vmName}
          onChange={(e) => setVmName(e.target.value)}
          placeholder="my-app"
          className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary focus:border-transparent"
          maxLength={50}
        />
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          This will be used for the domain: {vmName || "my-app"}.podland.app
        </p>
      </div>
    </div>
  );

  const renderStep2 = () => (
    <div>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Choose a Tier
      </h3>
      <div className="grid gap-3">
        {tiers.filter(isTierAvailable).map((tier) => {
          const quotaCheck = checkQuota(tier);
          const isSelected = selectedTier === tier.name;

          return (
            <button
              key={tier.name}
              onClick={() => quotaCheck.valid && setSelectedTier(tier.name)}
              disabled={!quotaCheck.valid}
              className={`p-4 border-2 rounded-lg text-left transition-all ${
                isSelected
                  ? "border-primary bg-primary/5 dark:bg-primary/10"
                  : quotaCheck.valid
                  ? "border-gray-200 dark:border-gray-700 hover:border-primary/50"
                  : "border-gray-200 dark:border-gray-700 opacity-50 cursor-not-allowed"
              }`}
            >
              <div className="flex justify-between items-center">
                <div>
                  <p className="font-semibold text-gray-900 dark:text-white capitalize">
                    {tier.name}
                  </p>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    {tier.cpu} CPU · {formatBytes(tier.ram)} RAM · {formatBytes(tier.storage)} Disk
                  </p>
                </div>
                {tier.min_role === "internal" && (
                  <span className="px-2 py-1 text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400 rounded">
                    Internal
                  </span>
                )}
              </div>
              {!quotaCheck.valid && (
                <p className="mt-2 text-sm text-red-600 dark:text-red-400">
                  ⚠️ {quotaCheck.message}
                </p>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );

  const renderStep3 = () => {
    const tier = tiers.find((t) => t.name === selectedTier);
    const quotaCheck = tier ? checkQuota(tier) : { valid: true, message: "" };

    return (
      <div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          Review & Create
        </h3>
        <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 mb-4">
          <h4 className="font-medium text-gray-900 dark:text-white mb-3">Summary</h4>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">Name:</span>
              <span className="text-gray-900 dark:text-white font-medium">{vmName}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">OS:</span>
              <span className="text-gray-900 dark:text-white">
                {selectedOS === "ubuntu-2204" ? "Ubuntu 22.04 LTS" : "Debian 12"}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">Tier:</span>
              <span className="text-gray-900 dark:text-white capitalize">{selectedTier}</span>
            </div>
            {tier && (
              <>
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">CPU:</span>
                  <span className="text-gray-900 dark:text-white">{tier.cpu} cores</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">RAM:</span>
                  <span className="text-gray-900 dark:text-white">{formatBytes(tier.ram)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Storage:</span>
                  <span className="text-gray-900 dark:text-white">{formatBytes(tier.storage)}</span>
                </div>
              </>
            )}
          </div>
        </div>
        {!quotaCheck.valid && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-600 dark:text-red-400">
              ⚠️ {quotaCheck.message}
            </p>
          </div>
        )}
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-3">
          <p className="text-sm text-blue-800 dark:text-blue-400">
            ℹ️ Your VM will be ready within 30 seconds. You'll receive an SSH key to connect.
          </p>
        </div>
      </div>
    );
  };

  const renderStep4 = () => (
    <div>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        VM Created Successfully!
      </h3>
      <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-4 mb-4">
        <p className="text-green-800 dark:text-green-400 font-medium">
          ✓ {vmName} is being provisioned
        </p>
      </div>

      {showSSHKey && createdVM?.ssh_key && (
        <div className="mb-4">
          <div className="flex justify-between items-center mb-2">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              SSH Private Key
            </label>
            <button
              onClick={handleDownloadSSHKey}
              className="text-sm text-primary hover:text-primary/80 font-medium flex items-center gap-1"
            >
              📥 Download
            </button>
          </div>
          <div className="bg-gray-900 rounded-lg p-4 overflow-x-auto">
            <pre className="text-xs text-green-400 font-mono whitespace-pre-wrap break-all">
              {createdVM.ssh_key}
            </pre>
          </div>
          <div className="mt-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-3">
            <p className="text-sm text-yellow-800 dark:text-yellow-400">
              ⚠️ <strong>Important:</strong> This is the only time your SSH private key will be shown.
              Please download it or copy it to a safe location. You cannot retrieve it later!
            </p>
          </div>
        </div>
      )}

      <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
        <p className="text-sm text-blue-800 dark:text-blue-400">
          ℹ️ Your VM will be ready in about 30 seconds. You can monitor its status in the VM list.
        </p>
      </div>
    </div>
  );

  return (
    <div 
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
    >
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          {/* Screen reader announcement */}
          <div 
            role="status" 
            aria-live="polite" 
            aria-atomic="true"
            className="sr-only"
          >
            Step {step} of 4: {stepTitles[step - 1]}
          </div>
          
          {/* Header */}
          <div className="flex justify-between items-center mb-4">
            <h2 id="modal-title" className="text-xl font-bold text-gray-900 dark:text-white">
              Create New VM
            </h2>
            <button
              ref={close_button_ref}
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              aria-label="Close create VM dialog"
            >
              ✕
            </button>
          </div>

          {renderStepIndicator()}

          {/* Step Content */}
          <div ref={stepContentRef} className="py-4">
            {step === 1 && renderStep1()}
            {step === 2 && renderStep2()}
            {step === 3 && renderStep3()}
            {step === 4 && renderStep4()}
          </div>

          {/* Footer Buttons */}
          {step < 4 && (
            <div className="flex justify-between pt-4 border-t border-gray-200 dark:border-gray-700">
              <button
                onClick={step === 1 ? onClose : handleBack}
                className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg font-medium transition-colors"
              >
                {step === 1 ? "Cancel" : "Back"}
              </button>
              <button
                onClick={handleNext}
                disabled={
                  (step === 1 && !vmName.trim()) ||
                  (step === 2 && !selectedTier) ||
                  (step === 3 && createMutation.isPending)
                }
                className="px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 disabled:bg-gray-400 disabled:cursor-not-allowed font-medium transition-colors"
              >
                {step === 3
                  ? createMutation.isPending
                    ? "Creating..."
                    : "Create VM"
                  : "Next"}
              </button>
            </div>
          )}

          {step === 4 && (
            <div className="flex justify-end pt-4 border-t border-gray-200 dark:border-gray-700">
              <button
                onClick={onSuccess}
                className="px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 font-medium transition-colors"
              >
                Done
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
