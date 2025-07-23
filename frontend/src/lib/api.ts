import { env } from "$env/dynamic/public";
import type {
    AuthResponse,
    CalendarEventsResponse,
    CalendarListResponse,
    GetCalendarEventsParams,
    GetGoogleCalendarsParams,
    GoogleOAuthLoginParams,
    ImportCalendarRequest,
    ImportCalendarResponse,
    LoginRequest,
    RegisterRequest,
    UserProfileResponse
} from "./types/api.js";

export class ApiError extends Error {
    constructor(
        message: string,
        public status: number,
        public data?: unknown
    ) {
        super(message);
        this.name = "ApiError";
    }
}

class ApiClient {
    public baseUrl: string;
    private token: string | null = null;

    constructor(baseUrl: string = "http://localhost:8000") {
        this.baseUrl = baseUrl;

        if (typeof window !== "undefined") {
            this.token = localStorage.getItem("auth_token");
        }
    }

    setToken(token: string | null) {
        this.token = token;
        if (typeof window !== "undefined") {
            if (token) {
                localStorage.setItem("auth_token", token);
            } else {
                localStorage.removeItem("auth_token");
            }
        }
    }

    getToken(): string | null {
        return this.token;
    }

    private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
        const url = `${this.baseUrl}${endpoint}`;

        const defaultHeaders: HeadersInit = {
            "Content-Type": "application/json"
        };

        if (this.token) {
            defaultHeaders.Authorization = `Bearer ${this.token}`;
        }

        const config: RequestInit = {
            ...options,
            credentials: "include",
            headers: {
                ...defaultHeaders,
                ...options.headers
            }
        };

        try {
            const response = await fetch(url, config);

            let data: unknown;
            const contentType = response.headers.get("content-type");

            if (contentType && contentType.includes("application/json")) {
                data = await response.json();
            } else {
                data = await response.text();
            }

            if (!response.ok) {
                const errorData =
                    typeof data === "object" && data !== null
                        ? (data as Record<string, unknown>)
                        : { message: String(data) };
                throw new ApiError(
                    (errorData.message as string) || `HTTP ${response.status}`,
                    response.status,
                    errorData
                );
            }

            return data as T;
        } catch (error) {
            if (error instanceof ApiError) {
                throw error;
            }
            throw new ApiError(error instanceof Error ? error.message : "Network error", 0);
        }
    }

    private get<T>(endpoint: string, params?: Record<string, string>): Promise<T> {
        const url = params ? `${endpoint}?${new URLSearchParams(params).toString()}` : endpoint;
        return this.request<T>(url, { method: "GET" });
    }

    private post<T>(endpoint: string, body?: unknown): Promise<T> {
        return this.request<T>(endpoint, {
            method: "POST",
            body: body ? JSON.stringify(body) : undefined
        });
    }

    // Authentication Methods
    async login(credentials: LoginRequest): Promise<AuthResponse> {
        const response = await this.post<AuthResponse>("/api/auth/login", credentials);
        if (response.success && response.token) {
            this.setToken(response.token);
        }
        return response;
    }

    async register(userData: RegisterRequest): Promise<AuthResponse> {
        const response = await this.post<AuthResponse>("/api/auth/register", userData);
        if (response.success && response.token) {
            this.setToken(response.token);
        }
        return response;
    }

    logout(): void {
        this.setToken(null);
    }

    // OAuth Methods
    getGoogleOAuthUrl(params?: GoogleOAuthLoginParams): string {
        const url = new URL(`${this.baseUrl}/api/auth/google/login`);
        if (params?.mode) url.searchParams.set("mode", params.mode);
        if (params?.from) url.searchParams.set("from", params.from);
        return url.toString();
    }

    // Calendar Methods
    async getCalendarEvents(params: GetCalendarEventsParams): Promise<CalendarEventsResponse> {
        const queryParams: Record<string, string> = {
            start_timestamp: params.start_timestamp,
            end_timestamp: params.end_timestamp
        };
        if (params.force_sync !== undefined) {
            queryParams.force_sync = params.force_sync.toString();
        }
        return this.get<CalendarEventsResponse>("/api/calendar/events", queryParams);
    }

    async getGoogleCalendars(params?: GetGoogleCalendarsParams): Promise<CalendarListResponse> {
        const queryParams: Record<string, string> = {};
        if (params?.force_sync !== undefined) {
            queryParams.force_sync = params.force_sync.toString();
        }
        return this.get<CalendarListResponse>(
            "/api/calendar/google",
            Object.keys(queryParams).length > 0 ? queryParams : undefined
        );
    }

    async importGoogleCalendar(request: ImportCalendarRequest): Promise<ImportCalendarResponse> {
        return this.post<ImportCalendarResponse>("/api/calendar/google", request);
    }

    // User Methods
    async getUserProfile(): Promise<UserProfileResponse> {
        return this.get<UserProfileResponse>("/api/user/profile");
    }

    // Utility Methods
    isAuthenticated(): boolean {
        return this.token !== null;
    }

    async handleApiCall<T>(
        apiCall: () => Promise<T>,
        onError?: (error: ApiError) => void
    ): Promise<T | null> {
        try {
            return await apiCall();
        } catch (error) {
            if (error instanceof ApiError) {
                if (error.status === 401) {
                    this.logout();
                }
                if (onError) {
                    onError(error);
                } else {
                    console.error("API Error:", error.message);
                }
            } else {
                console.error("Unexpected error:", error);
            }
            return null;
        }
    }
}

// Export singleton instance
export const api = new ApiClient(env.PUBLIC_API_BASE_URL || "http://localhost:8000");

// Export class for custom instances if needed
export { ApiClient };

// Helper function to handle common API patterns
export async function withAuth<T>(
    apiCall: () => Promise<T>,
    redirectTo?: string
): Promise<T | null> {
    if (!api.isAuthenticated()) {
        if (typeof window !== "undefined" && redirectTo) {
            window.location.href = redirectTo;
        }
        throw new ApiError("Authentication required", 401);
    }

    return api.handleApiCall(apiCall);
}
