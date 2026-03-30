type ActiveTab = "metrics" | "logs" | "alerts";

interface TabNavProps {
  activeTab: ActiveTab;
  onTabChange: (tab: ActiveTab) => void;
  alertsCount?: number;
}

export function TabNav({ activeTab, onTabChange, alertsCount }: TabNavProps) {
  const tabs: { id: ActiveTab; label: string; icon?: React.ReactNode }[] = [
    { id: "metrics", label: "Metrics" },
    { id: "logs", label: "Logs" },
    { id: "alerts", label: "Alerts", },
  ];

  return (
    <div className="mb-6">
      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="-mb-px flex gap-4" role="tablist">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => onTabChange(tab.id)}
              role="tab"
              aria-selected={activeTab === tab.id}
              className={`flex items-center gap-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors ${
                activeTab === tab.id
                  ? "border-blue-500 text-blue-600 dark:text-blue-400"
                  : "border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300 hover:border-gray-300"
              }`}
            >
              {tab.label}
              {tab.id === "alerts" && alertsCount && alertsCount > 0 && (
                <span className="ml-1 px-2 py-0.5 text-xs bg-red-500 text-white rounded-full">
                  {alertsCount > 9 ? "9+" : alertsCount}
                </span>
              )}
            </button>
          ))}
        </nav>
      </div>
    </div>
  );
}
