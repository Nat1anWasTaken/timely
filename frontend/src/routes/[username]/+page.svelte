<script lang="ts">
    import type { PageProps } from "./$types";
    import Calendar from "$lib/components/calendar/calendar.svelte";
    import { Avatar, AvatarFallback, AvatarImage } from "$lib/components/ui/avatar";
    import { Button } from "$lib/components/ui/button";
    import CalendarManagerSheet from "$lib/components/calendar/calendar-manager-sheet.svelte";
    import { createQuery } from "@tanstack/svelte-query";
    import { api } from "$lib/api";
    import { getMonthBoundaries, createQueryKey } from "$lib/utils/date";

    let { data }: PageProps = $props();

    let currentDate = new Date();
    let year = $state(currentDate.getFullYear());
    let month = $state(currentDate.getMonth());

    // Create reactive query for calendar events
    let calendarEventsQuery = $derived(
        createQuery({
            queryKey: createQueryKey(year, month),
            queryFn: async () => {
                const { start_timestamp, end_timestamp } = getMonthBoundaries(year, month);
                return await api.getCalendarEvents({ start_timestamp, end_timestamp });
            },
            staleTime: 5 * 60 * 1000, // 5 minutes
            gcTime: 30 * 60 * 1000 // 30 minutes
        })
    );

    let calendars = $derived(
        $calendarEventsQuery.data?.success && $calendarEventsQuery.data.calendars
            ? $calendarEventsQuery.data.calendars
            : []
    );

    function handleMonthChange(newYear: number, newMonth: number) {
        year = newYear;
        month = newMonth;
    }
</script>

{#if data.isViewingSelf && data.user?.user}
    <div class="container flex flex-row items-start justify-center gap-4 p-4">
        <!-- User Profile Header -->
        <div class="mb-8 flex flex-row items-start gap-4">
            <Avatar class="h-16 w-16">
                <AvatarImage src={data.user.user.picture} alt={data.user.user.display_name} />
                <AvatarFallback>
                    {data.user.user.display_name.charAt(0).toUpperCase()}
                </AvatarFallback>
            </Avatar>
            <div class="flex-1">
                <h1 class="text-3xl font-bold">{data.user.user.display_name}'s Calendar</h1>
                <div class="mt-3">
                    <CalendarManagerSheet user={data.user.user}>
                        <Button variant="outline" size="sm">Manage my calendars</Button>
                    </CalendarManagerSheet>
                </div>
            </div>
        </div>

        <!-- Calendar Component -->
        {#if $calendarEventsQuery.error}
            <div class="mb-4 rounded-lg bg-red-50 p-4 text-red-800">
                Error loading calendar events: {$calendarEventsQuery.error.message}
            </div>
        {/if}

        <div class:opacity-50={$calendarEventsQuery.isLoading}>
            <Calendar {calendars} bind:year bind:month onMonthChange={handleMonthChange} />
        </div>
    </div>
{:else}
    <div class="container p-4">
        <p>{data.message}</p>
    </div>
{/if}
