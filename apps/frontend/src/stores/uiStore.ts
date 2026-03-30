import { create } from "zustand";

interface UIState {
  // VM List filters
  vmStatusFilter: string;
  vmSortField: string;
  vmSortOrder: "asc" | "desc";
  vmCurrentPage: number;

  // Actions
  setVMStatusFilter: (filter: string) => void;
  setVMSortField: (field: string) => void;
  setVMSortOrder: (order: "asc" | "desc") => void;
  setVMCurrentPage: (page: number) => void;
  resetVMFilters: () => void;
}

export const useUIStore = create<UIState>((set) => ({
  // Initial state
  vmStatusFilter: "all",
  vmSortField: "created_at",
  vmSortOrder: "desc",
  vmCurrentPage: 1,

  // Actions
  setVMStatusFilter: (filter) => set({ vmStatusFilter: filter, vmCurrentPage: 1 }),
  setVMSortField: (field) => set({ vmSortField: field }),
  setVMSortOrder: (order) => set({ vmSortOrder: order }),
  setVMCurrentPage: (page) => set({ vmCurrentPage: page }),
  resetVMFilters: () => set({
    vmStatusFilter: "all",
    vmSortField: "created_at",
    vmSortOrder: "desc",
    vmCurrentPage: 1,
  }),
}));
