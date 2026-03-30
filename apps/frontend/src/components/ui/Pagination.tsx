interface PaginationProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  totalItems?: number;
  itemsPerPage?: number;
}

export function Pagination({
  currentPage,
  totalPages,
  onPageChange,
  totalItems,
  itemsPerPage,
}: PaginationProps) {
  if (totalPages <= 1) {
    return null;
  }

  const currentPageSafe = Math.min(currentPage, totalPages);
  const startItem = (currentPageSafe - 1) * (itemsPerPage || 10) + 1;
  const endItem = Math.min(currentPageSafe * (itemsPerPage || 10), totalItems || 0);

  return (
    <div className="flex justify-center gap-2 mt-6">
      <button
        onClick={() => onPageChange(Math.max(1, currentPageSafe - 1))}
        disabled={currentPageSafe === 1}
        className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
      >
        Previous
      </button>
      <span className="px-4 py-2 text-gray-700 dark:text-gray-300">
        Page {currentPageSafe} of {totalPages}
        {totalItems !== undefined && (
          <span className="ml-2 text-sm text-gray-500 dark:text-gray-400">
            ({startItem}-{endItem} of {totalItems})
          </span>
        )}
      </span>
      <button
        onClick={() => onPageChange(Math.min(totalPages, currentPageSafe + 1))}
        disabled={currentPageSafe === totalPages}
        className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
      >
        Next
      </button>
    </div>
  );
}
