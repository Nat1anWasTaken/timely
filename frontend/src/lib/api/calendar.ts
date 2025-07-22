import { apiClient } from "./client.js";
import type {
    CalendarEventsResponse,
    CalendarListResponse,
    ImportCalendarRequest,
    ImportCalendarResponse,
    GetCalendarEventsParams
} from "../types/api.js";

export class CalendarAPI {
    /**
     * Get calendar events within a specified time range
     * @param params Time range parameters (start and end timestamps)
     */
    async getEvents(params: GetCalendarEventsParams): Promise<CalendarEventsResponse> {
        return apiClient.get<CalendarEventsResponse>("/api/calendar/events", params);
    }

    /**
     * Get all Google calendars for the authenticated user
     */
    async getGoogleCalendars(): Promise<CalendarListResponse> {
        return apiClient.get<CalendarListResponse>("/api/calendar/google");
    }

    /**
     * Import a specific Google calendar to the user's database
     * @param request Calendar import request with calendar_id
     */
    async importGoogleCalendar(request: ImportCalendarRequest): Promise<ImportCalendarResponse> {
        return apiClient.post<ImportCalendarResponse>("/api/calendar/google", request);
    }

    /**
     * Helper method to get events for a specific date range using Date objects
     * @param startDate Start date
     * @param endDate End date
     */
    async getEventsByDateRange(startDate: Date, endDate: Date): Promise<CalendarEventsResponse> {
        const params: GetCalendarEventsParams = {
            start_timestamp: Math.floor(startDate.getTime() / 1000).toString(),
            end_timestamp: Math.floor(endDate.getTime() / 1000).toString()
        };
        return this.getEvents(params);
    }

    /**
     * Helper method to get events for the current month
     */
    async getCurrentMonthEvents(): Promise<CalendarEventsResponse> {
        const now = new Date();
        const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
        const endOfMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0, 23, 59, 59);

        return this.getEventsByDateRange(startOfMonth, endOfMonth);
    }

    /**
     * Helper method to get events for a specific month and year
     * @param year Year (e.g., 2024)
     * @param month Month (0-11, where 0 = January)
     */
    async getMonthEvents(year: number, month: number): Promise<CalendarEventsResponse> {
        const startOfMonth = new Date(year, month, 1);
        const endOfMonth = new Date(year, month + 1, 0, 23, 59, 59);

        return this.getEventsByDateRange(startOfMonth, endOfMonth);
    }

    /**
     * Helper method to get events for today
     */
    async getTodayEvents(): Promise<CalendarEventsResponse> {
        const today = new Date();
        const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate());
        const endOfDay = new Date(
            today.getFullYear(),
            today.getMonth(),
            today.getDate(),
            23,
            59,
            59
        );

        return this.getEventsByDateRange(startOfDay, endOfDay);
    }

    /**
     * Helper method to get events for the next 7 days
     */
    async getWeekEvents(): Promise<CalendarEventsResponse> {
        const today = new Date();
        const startOfToday = new Date(today.getFullYear(), today.getMonth(), today.getDate());
        const endOfWeek = new Date(startOfToday.getTime() + 7 * 24 * 60 * 60 * 1000);
        endOfWeek.setHours(23, 59, 59);

        return this.getEventsByDateRange(startOfToday, endOfWeek);
    }
}

// Export singleton instance
export const calendar = new CalendarAPI();
