import { apiClient } from "./client.js";
import type { LoginRequest, RegisterRequest, AuthResponse } from "../types/api.js";

export class AuthenticationAPI {
    /**
     * Authenticate user with email and password
     */
    async login(credentials: LoginRequest): Promise<AuthResponse> {
        const response = await apiClient.post<AuthResponse>("/api/auth/login", credentials);
        if (response.success && response.token) {
            apiClient.setToken(response.token);
        }
        return response;
    }

    /**
     * Register a new user account
     */
    async register(userData: RegisterRequest): Promise<AuthResponse> {
        const response = await apiClient.post<AuthResponse>("/api/auth/register", userData);
        if (response.success && response.token) {
            apiClient.setToken(response.token);
        }
        return response;
    }

    /**
     * Logout current user (clears token)
     */
    logout(): void {
        apiClient.setToken(null);
    }

    /**
     * Check if user is authenticated
     */
    isAuthenticated(): boolean {
        return apiClient.isAuthenticated();
    }

    /**
     * Get current auth token
     */
    getToken(): string | null {
        return apiClient.getToken();
    }
}

// Export singleton instance
export const auth = new AuthenticationAPI();
