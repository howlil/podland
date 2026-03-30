import { VM } from "@/hooks/useVMs";
import { Globe, Terminal, Download, Shield, ExternalLink } from "lucide-react";

interface ConnectionInfoProps {
  vm?: VM;
  isLoading?: boolean;
  onDownloadSSHKey?: () => void;
}

export function ConnectionInfo({ vm, isLoading, onDownloadSSHKey }: ConnectionInfoProps) {
  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-6 animate-pulse">
        <div className="h-6 w-48 bg-gray-200 dark:bg-gray-700 rounded mb-4" />
        <div className="space-y-4">
          <div className="space-y-2">
            <div className="h-4 w-20 bg-gray-200 dark:bg-gray-700 rounded" />
            <div className="h-6 w-64 bg-gray-200 dark:bg-gray-700 rounded" />
          </div>
          <div className="space-y-2">
            <div className="h-4 w-24 bg-gray-200 dark:bg-gray-700 rounded" />
            <div className="h-10 w-full bg-gray-200 dark:bg-gray-700 rounded" />
          </div>
        </div>
      </div>
    );
  }

  if (!vm) return null;

  const domain = vm.domain || `${vm.name}.podland.app`;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-6">
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        <Globe className="h-5 w-5 text-cyan-600 dark:text-cyan-400" />
        Connection Information
      </h2>
      <div className="space-y-6">
        {/* Domain */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Domain</p>
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 rounded-full text-xs font-medium">
              <span className="w-1.5 h-1.5 rounded-full bg-green-500 animate-pulse" />
              Active
            </span>
          </div>
          <a
            href={`https://${domain}`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 text-lg font-mono text-blue-600 dark:text-blue-400 hover:underline transition-colors"
          >
            {domain}
            <ExternalLink className="h-4 w-4" />
          </a>
        </div>

        {/* SSH Access */}
        <div>
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2 flex items-center gap-2">
            <Terminal className="h-4 w-4" />
            SSH Access
          </p>
          <div className="flex flex-col sm:flex-row gap-3">
            <code className="flex-1 bg-gray-100 dark:bg-gray-700 px-4 py-3 rounded-xl text-sm font-mono text-gray-900 dark:text-white break-all">
              ssh -i ~/.ssh/id_ed25519 user@{domain}
            </code>
            <button
              onClick={onDownloadSSHKey}
              className="inline-flex items-center justify-center gap-2 px-4 py-3 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-xl text-sm font-medium text-gray-700 dark:text-gray-300 transition-all"
            >
              <Download className="h-4 w-4" />
              Download Key
            </button>
          </div>
          <div className="flex items-start gap-2 mt-3 p-3 bg-yellow-50 dark:bg-yellow-900/10 border border-yellow-200 dark:border-yellow-800 rounded-xl">
            <Shield className="h-4 w-4 text-yellow-600 dark:text-yellow-400 mt-0.5 flex-shrink-0" />
            <p className="text-xs text-yellow-700 dark:text-yellow-400">
              <strong>Important:</strong> The SSH private key was shown only once during VM creation. If you didn't save it, you'll need to recreate the VM.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
