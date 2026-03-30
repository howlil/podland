import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { toast } from "sonner";

interface User {
  id: string;
  email: string;
  display_name: string;
  role: string;
  nim?: string;
  created_at: string;
}

export function useAdminUsers(roleFilter?: string) {
  const queryClient = useQueryClient();

  const { data: users = [], isLoading, error } = useQuery<User[]>({
    queryKey: ["admin-users", roleFilter],
    queryFn: async () => {
      const { data } = await api.get(`/admin/users${roleFilter ? `?role=${roleFilter}` : ""}`);
      return data;
    },
  });

  const changeRoleMutation = useMutation({
    mutationFn: async ({ userId, role }: { userId: string; role: string }) => {
      const { data } = await api.patch(`/admin/users/${userId}/role`, { role });
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      toast.success("User role updated successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to update role: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  const banUserMutation = useMutation({
    mutationFn: async (userId: string) => {
      const { data } = await api.post(`/admin/users/${userId}/ban`);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      toast.success("User banned successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to ban user: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  const unbanUserMutation = useMutation({
    mutationFn: async (userId: string) => {
      const { data } = await api.post(`/admin/users/${userId}/unban`);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      toast.success("User unbanned successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to unban user: ${error.response?.data?.message || "Unknown error"}`);
    },
  });

  return {
    users,
    isLoading,
    error,
    changeRole: changeRoleMutation.mutate,
    banUser: banUserMutation.mutate,
    unbanUser: unbanUserMutation.mutate,
    isChangingRole: changeRoleMutation.isPending,
    isBanning: banUserMutation.isPending,
    isUnbanning: unbanUserMutation.isPending,
  };
}
