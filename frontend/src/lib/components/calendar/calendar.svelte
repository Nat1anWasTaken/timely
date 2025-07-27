<script lang="ts">
    import Weekday from "$lib/components/calendar/weekday.svelte";
    import Day from "$lib/components/calendar/day.svelte";
    import CalendarEventComponent from "$lib/components/calendar/calendar-event.svelte";
    import calendar from "calendar-js";
    import { Card } from "$lib/components/ui/card";
    import YearMonthSelector from "$lib/components/calendar/year-month-selector.svelte";
    import type { CalendarEvent, CalendarWithEvents } from "$lib/types/api";
    import { cn } from "$lib/utils";

    interface Props {
        calendars?: CalendarWithEvents[];
        year: number;
        month: number;
        onMonthChange?: (year: number, month: number) => void;
        class?: string;
    }

    let {
        calendars = [],
        year = $bindable(),
        month = $bindable(),
        onMonthChange,
        class: className
    }: Props = $props();

    let calendarDays = $derived((calendar() as any).detailed(year, month).calendar);

    // Flatten events from all calendars with their calendar context
    let events = $derived(
        calendars.flatMap((calendar) => calendar.events.map((event) => ({ event, calendar })))
    );

    // Call onMonthChange when year or month changes
    $effect(() => {
        onMonthChange?.(year, month);
    });

    // Helper function to get events for a specific date
    function getEventsForDate(targetDate: Date) {
        return events.filter(({ event }) => {
            const eventStart = new Date(event.start);
            const eventEnd = new Date(event.end);

            // Check if the event occurs on this day
            return (
                eventStart.toDateString() === targetDate.toDateString() ||
                (eventStart <= targetDate && eventEnd >= targetDate)
            );
        });
    }
</script>

<div class={cn("flex h-full w-full flex-col", className)}>
    <!--Calendar Header-->
    <div class="mb-4 flex flex-row items-center justify-end">
        <YearMonthSelector bind:year bind:month />
    </div>

    <!--Calendar Grid-->
    <div class="grid flex-1 grid-cols-7 gap-2" style="grid-template-rows: auto repeat(6, 1fr);">
        <!--Weekday Headers-->
        <Weekday specialDay>Sunday</Weekday>
        <Weekday>Monday</Weekday>
        <Weekday>Tuesday</Weekday>
        <Weekday>Wednesday</Weekday>
        <Weekday>Thursday</Weekday>
        <Weekday>Friday</Weekday>
        <Weekday specialDay>Saturday</Weekday>

        <!--Calendar Days-->
        {#each calendarDays as daysInWeek, index (index)}
            {#each daysInWeek as dayInfo, index (index)}
                <Day day={dayInfo.day} isCurrentMonth={dayInfo.isInPrimaryMonth}>
                    {#each getEventsForDate(dayInfo.date) as { event, calendar } (event.id)}
                        <CalendarEventComponent {event} {calendar} />
                    {/each}
                </Day>
            {/each}
        {/each}
    </div>
</div>
