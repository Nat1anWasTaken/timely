import { createQuery } from "@tanstack/svelte-query";
import { api } from "./api";

export function createUserDataQuery() {
    return createQuery({
        queryKey: ["userData"],
        queryFn: () => api.getUserProfile(),
        staleTime: 1000 * 60 * 5,
        refetchOnMount: true,
        refetchOnReconnect: true,
        refetchOnWindowFocus: true,
        retry: (failureCount, error) => {
            // Don't retry on 401 (unauthorized) - user is not logged in
            if (error && typeof error === "object" && "status" in error && error.status === 401) {
                return false;
            }
            return failureCount < 3;
        }
    });
}
