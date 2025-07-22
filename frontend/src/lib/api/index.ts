// Core API client
export { ApiClient, ApiError, apiClient } from "./client.js";

// Authentication API
export { AuthenticationAPI, auth } from "./authentication.js";

// OAuth API
export { OAuthAPI, oauth } from "./oauth.js";

// Calendar API
export { CalendarAPI, calendar } from "./calendar.js";

// Main API object with all endpoints
export const api = {
    auth,
    oauth,
    calendar,
    client: apiClient
} as const;

// Utility function for authenticated API calls
export async function withAuth<T>(
    apiCall: () => Promise<T>,
    redirectTo?: string
): Promise<T | null> {
    if (!apiClient.isAuthenticated()) {
        if (typeof window !== "undefined" && redirectTo) {
            window.location.href = redirectTo;
        }
        throw new ApiError("Authentication required", 401);
    }

    return apiClient.handleApiCall(apiCall);
}

// Re-export all types
export type * from "../types/api.js";
