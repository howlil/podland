interface VMFiltersProps {
  statusFilter: string;
  onStatusFilterChange: (status: string) => void;
  onCreateVM: () => void;
}

export function VMFilters({
  statusFilter,
  onStatusFilterChange,
  onCreateVM,
}: VMFiltersProps) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
      <div className="flex flex-wrap gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Status
          </label>
          <select
            value={statusFilter}
            onChange={(e) => onStatusFilterChange(e.target.value)}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
          >
            <option value="all">All</option>
            <option value="running">Running</option>
            <option value="stopped">Stopped</option>
            <option value="pending">Pending</option>
            <option value="error">Error</option>
          </select>
        </div>
        <div className="flex-1" />
        <div className="flex items-end">
          <button
            onClick={onCreateVM}
            className="px-6 py-2 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white rounded-xl font-semibold shadow-lg transition-all"
          >
            <span className="inline-flex items-center gap-2">
              <span>+</span> Create VM
            </span>
          </button>
        </div>
      </div>
    </div>
  );
}
