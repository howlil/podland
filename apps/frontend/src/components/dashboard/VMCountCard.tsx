interface VMCountCardProps {
  count: number;
}

export function VMCountCard({ count }: VMCountCardProps) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <div className="flex items-center">
        <div className="flex-shrink-0 w-12 h-12 bg-blue-100 dark:bg-blue-900/20 rounded-lg flex items-center justify-center">
          <span className="text-2xl">💻</span>
        </div>
        <div className="ml-4">
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
            VMs Running
          </p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">
            {count}
          </p>
        </div>
      </div>
    </div>
  );
}
