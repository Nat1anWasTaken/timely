<script lang="ts">
    import type { PageProps } from "./$types";
    import Calendar from "$lib/components/calendar/calendar.svelte";
    import { Avatar, AvatarFallback, AvatarImage } from "$lib/components/ui/avatar";
    import { Button } from "$lib/components/ui/button";
    import CalendarManagerSheet from "$lib/components/calendar/calendar-manager-sheet.svelte";
    import { createQuery } from "@tanstack/svelte-query";
    import { api } from "$lib/api";
    import { getExtendedMonthBoundaries, createExtendedQueryKey } from "$lib/utils/date";

    let { data }: PageProps = $props();

    let currentDate = new Date();
    let year = $state(currentDate.getFullYear());
    let month = $state(currentDate.getMonth());

    // Create reactive query for calendar events (private view)
    let calendarEventsQuery = $derived(
        data.isViewingSelf
            ? createQuery({
                  queryKey: createExtendedQueryKey(year, month),
                  queryFn: async () => {
                      const { start_timestamp, end_timestamp } = getExtendedMonthBoundaries(
                          year,
                          month
                      );
                      return await api.getCalendarEvents({ start_timestamp, end_timestamp });
                  },
                  staleTime: 5 * 60 * 1000, // 5 minutes
                  gcTime: 30 * 60 * 1000 // 30 minutes
              })
            : null
    );

    // Create reactive query for public calendar events
    let publicCalendarEventsQuery = $derived(
        !data.isViewingSelf && data.publicUser
            ? createQuery({
                  queryKey: ["public-events-extended", data.publicUser.username, year, month],
                  queryFn: async () => {
                      const { start_timestamp, end_timestamp } = getExtendedMonthBoundaries(
                          year,
                          month
                      );
                      return await api.getPublicUserEvents({
                          username: data.publicUser!.username,
                          start_timestamp,
                          end_timestamp
                      });
                  },
                  staleTime: 5 * 60 * 1000, // 5 minutes
                  gcTime: 30 * 60 * 1000 // 30 minutes
              })
            : null
    );

    let calendars = $derived.by(() => {
        if (data.isViewingSelf && calendarEventsQuery) {
            const query = $calendarEventsQuery;
            if (query?.data?.success && query.data.calendars) {
                return query.data.calendars;
            } else {
                return [];
            }
        } else if (!data.isViewingSelf && publicCalendarEventsQuery) {
            const query = $publicCalendarEventsQuery;
            if (query?.data?.success && query.data.calendars) {
                return query.data.calendars;
            } else {
                return [];
            }
        }
        return [];
    });

    function handleMonthChange(newYear: number, newMonth: number) {
        year = newYear;
        month = newMonth;
    }
</script>

{#if data.isViewingSelf && data.user}
    <div class="container flex h-full flex-col items-start gap-4 p-4 md:flex-row">
        <!-- User Profile Header -->
        <div class="mb-8 flex flex-row items-start gap-4">
            <Avatar class="h-16 w-16">
                <AvatarImage src={data.user.picture} alt={data.user.display_name} />
                <AvatarFallback>
                    {data.user.display_name.charAt(0).toUpperCase()}
                </AvatarFallback>
            </Avatar>
            <div class="flex-1">
                <h1 class="text-3xl font-bold">{data.user.display_name}'s Calendar</h1>
                <div class="mt-3">
                    <CalendarManagerSheet user={data.user}>
                        <Button variant="outline" size="sm">Manage my calendars</Button>
                    </CalendarManagerSheet>
                </div>
            </div>
        </div>

        <!-- Calendar Component -->
        {#if calendarEventsQuery}
            {@const query = $calendarEventsQuery}
            {#if query?.error}
                <div class="mb-4 rounded-lg bg-red-50 p-4 text-red-800">
                    Error loading calendar events: {query.error.message}
                </div>
            {/if}
        {/if}
        <Calendar
            {calendars}
            bind:year
            bind:month
            onMonthChange={handleMonthChange}
            class="h-full w-full"
        />
    </div>
{:else if !data.isViewingSelf && data.publicUser}
    <div class="container flex h-full flex-col items-start gap-4 p-4 md:flex-row">
        <!-- Public User Profile Header -->
        <div class="mb-8 flex flex-row items-start gap-4">
            <Avatar class="h-16 w-16">
                <AvatarImage src={data.publicUser.picture} alt={data.publicUser.display_name} />
                <AvatarFallback>
                    {data.publicUser.display_name.charAt(0).toUpperCase()}
                </AvatarFallback>
            </Avatar>
            <div class="flex-1">
                <h1 class="text-3xl font-bold">{data.publicUser.display_name}'s Calendar</h1>
                <p class="text-muted-foreground text-sm">
                    @{data.publicUser.username} • Joined {new Date(
                        data.publicUser.created_at
                    ).toLocaleDateString()}
                </p>
            </div>
        </div>

        <!-- Public Calendar Component -->
        {#if publicCalendarEventsQuery}
            {@const query = $publicCalendarEventsQuery}
            {#if query?.error}
                <div class="mb-4 rounded-lg bg-red-50 p-4 text-red-800">
                    Error loading calendar events: {query.error.message}
                </div>
            {/if}
        {/if}

        <Calendar
            {calendars}
            bind:year
            bind:month
            onMonthChange={handleMonthChange}
            class="h-full w-full"
        />
    </div>
{:else}
    <div class="container p-4">
        <p>{data.message}</p>
    </div>
{/if}
