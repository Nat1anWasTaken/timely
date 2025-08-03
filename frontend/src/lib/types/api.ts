// API Response Types
export interface ApiResponse<T = unknown> {
    success: boolean;
    message: string;
    data?: T;
}

export interface ErrorResponse {
    success: false;
    message: string;
    error: string;
}

// Authentication Types
export interface LoginRequest {
    email: string;
    password: string;
}

export interface RegisterRequest {
    email: string;
    username: string;
    display_name: string;
    password: string;
}

export interface AuthResponse {
    success: true;
    message: string;
    token: string;
    user: User;
}

export interface UserProfileResponse {
    success: boolean;
    message: string;
    user: User;
}

export interface PublicUserProfile {
    id: string;
    username: string;
    display_name: string;
    picture?: string;
    created_at: string;
}

export interface PublicUserProfileResponse {
    success: boolean;
    message: string;
    user: PublicUserProfile;
}

export interface UpdateUserProfileRequest {
    username?: string;
    display_name?: string;
}

export interface UpdateUserProfileResponse {
    success: boolean;
    message: string;
    user: User;
}

// User Types
export interface User {
    id: string;
    username: string;
    display_name: string;
    email?: string;
    picture?: string;
    created_at: string;
    updated_at: string;
    accounts?: Account[];
}

export interface Account {
    id: string;
    user_id: string;
    provider: string;
    provider_id: string;
    email?: string;
    expiry?: string;
    created_at: string;
    updated_at: string;
}

// Calendar Types
export type CalendarSource = "google" | "isc";
export type CalendarVisibility = "public" | "private";
export type CalendarEventVisibility = "public" | "private" | "inherited";

export interface Calendar {
    id: string;
    user_id: string;
    source: CalendarSource;
    source_id: string;
    summary: string;
    description?: string;
    time_zone?: string;
    visibility: CalendarVisibility;
    event_color?: string;
    event_redaction?: string;
    synced_at?: string;
    created_at: string;
    updated_at: string;
}

export interface CalendarEvent {
    id: string;
    calendar_id: string;
    source_id: string;
    title: string;
    description?: string;
    start: string;
    end: string;
    all_day: boolean;
    location?: string;
    visibility: CalendarEventVisibility;
    event_color?: string;
    created_at: string;
    updated_at: string;
}

export interface CalendarWithEvents extends Calendar {
    events: CalendarEvent[];
}

export interface GoogleCalendar {
    id: string;
    kind: string;
    etag: string;
    summary: string;
    description?: string;
    location?: string;
    timeZone: string;
    summaryOverride?: string;
    colorId?: string;
    backgroundColor?: string;
    foregroundColor?: string;
    hidden: boolean;
    selected: boolean;
    accessRole: string;
    primary?: boolean;
    deleted: boolean;
    conferenceProperties?: GoogleCalendarConferenceProperties;
}

export interface GoogleCalendarConferenceProperties {
    allowedConferenceSolutionTypes: string[];
}

// API Response Types
export interface CalendarEventsResponse {
    success: boolean;
    message: string;
    calendars: CalendarWithEvents[];
}

export interface ImportedCalendarsResponse {
    success: boolean;
    message: string;
    calendars: Calendar[];
}

export interface CalendarListResponse {
    success: boolean;
    message: string;
    calendars: GoogleCalendar[];
}

export interface ImportCalendarRequest {
    calendar_id: string;
}

export interface ImportCalendarResponse {
    success: boolean;
    message: string;
    calendar: Calendar;
}

export interface ImportICSRequest {
    ics_data: string;
    calendar_name?: string;
}

export interface ImportICSResponse {
    success: boolean;
    message: string;
    calendar: Calendar;
    events_count?: number;
}

export interface CalendarUpdateRequest {
    summary?: string;
    description?: string;
    time_zone?: string;
    visibility?: CalendarVisibility;
    event_color?: string;
    event_redaction?: string;
}

export interface CalendarUpdateResponse {
    success: boolean;
    message: string;
    calendar: Calendar;
}

export interface CalendarDeleteResponse {
    success: boolean;
    message: string;
}

// API Query Parameters
export interface GetCalendarEventsParams {
    start_timestamp: string;
    end_timestamp: string;
    force_sync?: boolean;
}

export interface GoogleOAuthCallbackParams {
    code: string;
    state: string;
}

export interface GetGoogleCalendarsParams {
    force_sync?: boolean;
}

export interface GoogleOAuthLoginParams {
    mode?: "login" | "link";
    from?: string;
}

export interface GetPublicUserEventsParams {
    username: string;
    start_timestamp: string;
    end_timestamp: string;
}
