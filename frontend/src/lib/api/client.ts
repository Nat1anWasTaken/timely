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

export class ApiClient {
    private baseUrl: string;
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

    isAuthenticated(): boolean {
        return this.token !== null;
    }

    async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
        const url = `${this.baseUrl}${endpoint}`;

        const defaultHeaders: HeadersInit = {
            "Content-Type": "application/json"
        };

        if (this.token) {
            defaultHeaders.Authorization = `Bearer ${this.token}`;
        }

        const config: RequestInit = {
            ...options,
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

    get<T>(endpoint: string, params?: Record<string, string>): Promise<T> {
        const url = params ? `${endpoint}?${new URLSearchParams(params).toString()}` : endpoint;
        return this.request<T>(url, { method: "GET" });
    }

    post<T>(endpoint: string, body?: unknown): Promise<T> {
        return this.request<T>(endpoint, {
            method: "POST",
            body: body ? JSON.stringify(body) : undefined
        });
    }

    put<T>(endpoint: string, body?: unknown): Promise<T> {
        return this.request<T>(endpoint, {
            method: "PUT",
            body: body ? JSON.stringify(body) : undefined
        });
    }

    delete<T>(endpoint: string): Promise<T> {
        return this.request<T>(endpoint, { method: "DELETE" });
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
                    this.setToken(null);
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

// Create singleton instance
export const apiClient = new ApiClient();
