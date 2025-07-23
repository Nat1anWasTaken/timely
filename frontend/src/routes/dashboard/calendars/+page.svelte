<script lang="ts">
    import CalendarsHeader from "$lib/components/dashboard/calendars/calendars-header.svelte";
    import CalendarList from "$lib/components/dashboard/calendars/calendar-list.svelte";
    import AddCalendarCard from "$lib/components/dashboard/calendars/add-calendar-card.svelte";
    import { createUserDataQuery } from "$lib/globalQueries";
    import { createQuery } from "@tanstack/svelte-query";
    import { api } from "$lib/api";

    const userDataQuery = createUserDataQuery();

    let importedCalendarQuery = createQuery({
        queryKey: ["imported-calendars"],
        queryFn: () => api.getImportedCalendars()
    });
</script>

<div class="w-2xl max-w-[90vw] space-y-6">
    <CalendarsHeader />

    <CalendarList
        calendars={$importedCalendarQuery.data?.calendars || []}
        isLoading={$importedCalendarQuery.isLoading}
        isError={$importedCalendarQuery.isError}
        errorMessage={$importedCalendarQuery.error?.message || ""}
        isUserLoggedIn={!!$userDataQuery.data?.user}
    />

    <AddCalendarCard user={$userDataQuery.data?.user} />
</div>
