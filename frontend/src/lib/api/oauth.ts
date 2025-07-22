import { apiClient } from "./client.js";
import type { GoogleOAuthLoginParams, GoogleOAuthCallbackParams } from "../types/api.js";

export class OAuthAPI {
    /**
     * Get Google OAuth login URL
     * @param params Optional parameters for OAuth flow
     * @returns Complete Google OAuth URL
     */
    getGoogleLoginUrl(params?: GoogleOAuthLoginParams): string {
        const baseUrl = "http://localhost:8000"; // TODO: Make this configurable
        const url = new URL(`${baseUrl}/api/auth/google/login`);
        if (params?.mode) url.searchParams.set("mode", params.mode);
        if (params?.from) url.searchParams.set("from", params.from);
        return url.toString();
    }

    /**
     * Initiate Google OAuth login by redirecting to Google
     * @param params Optional parameters for OAuth flow
     */
    initiateGoogleLogin(params?: GoogleOAuthLoginParams): void {
        if (typeof window !== "undefined") {
            window.location.href = this.getGoogleLoginUrl(params);
        }
    }

    /**
     * Handle Google OAuth callback (typically called automatically by redirect)
     * This endpoint is usually handled by the server redirect, but can be called manually
     * @param params OAuth callback parameters (code and state)
     */
    async handleGoogleCallback(params: GoogleOAuthCallbackParams): Promise<void> {
        const url = `/api/auth/google/callback?${new URLSearchParams(params).toString()}`;

        // This will typically redirect, so we don't expect a JSON response
        await apiClient.request(url, {
            method: "GET",
            headers: {} // Remove Content-Type header for this request
        });
    }

    /**
     * Extract OAuth parameters from current URL
     * Useful for handling OAuth callbacks in SPA applications
     */
    getOAuthParamsFromUrl(): GoogleOAuthCallbackParams | null {
        if (typeof window === "undefined") return null;

        const urlParams = new URLSearchParams(window.location.search);
        const code = urlParams.get("code");
        const state = urlParams.get("state");

        if (code && state) {
            return { code, state };
        }

        return null;
    }

    /**
     * Check if current URL contains OAuth callback parameters
     */
    isOAuthCallback(): boolean {
        return this.getOAuthParamsFromUrl() !== null;
    }
}

// Export singleton instance
export const oauth = new OAuthAPI();
