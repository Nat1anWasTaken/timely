<script lang="ts">
    import Weekday from "$lib/components/calendar/weekday.svelte";
    import Day from "$lib/components/calendar/day.svelte";
    import CalendarEventComponent from "$lib/components/calendar/calendar-event.svelte";
    import calendar from "calendar-js";
    import { Card } from "$lib/components/ui/card";
    import YearMonthSelector from "$lib/components/calendar/year-month-selector.svelte";
    import type { CalendarEvent, CalendarWithEvents } from "$lib/types/api";

    interface Props {
        calendars?: CalendarWithEvents[];
        year: number;
        month: number;
        onMonthChange?: (year: number, month: number) => void;
    }

    let { calendars = [], year = $bindable(), month = $bindable(), onMonthChange }: Props = $props();

    let calendarDays = $derived(calendar().of(year, month).calendar);

    // Flatten events from all calendars with their calendar context
    let events = $derived(
        calendars.flatMap(calendar => 
            calendar.events.map(event => ({ event, calendar }))
        )
    );

    // Call onMonthChange when year or month changes
    $effect(() => {
        onMonthChange?.(year, month);
    });

    // Helper function to get events for a specific day
    function getEventsForDay(day: number) {
        const targetDate = new Date(year, month, day);
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

<div class="w-4xl flex-col">
    <!--Calendar Header-->
    <div class="mb-4 flex flex-row items-center justify-end">
        <YearMonthSelector bind:year bind:month />
    </div>

    <!--Calendar Grid-->
    <div class="grid grid-cols-7 gap-2">
        <!--Top field-->
        <Weekday specialDay>Sunday</Weekday>
        <Weekday>Monday</Weekday>
        <Weekday>Tuesday</Weekday>
        <Weekday>Wednesday</Weekday>
        <Weekday>Thursday</Weekday>
        <Weekday>Friday</Weekday>
        <Weekday specialDay>Saturday</Weekday>

        <!--Bottom field-->
        {#each calendarDays as daysInWeek, index (index)}
            {#each daysInWeek as day, index (index)}
                {#if day !== 0}
                    <Day {day}>
                        {#each getEventsForDay(day) as { event, calendar } (event.id)}
                            <CalendarEventComponent {event} {calendar} />
                        {/each}
                    </Day>
                {:else}
                    <!-- Empty day cell for days not in the current month -->
                    <Card class="h-32 p-4" />
                {/if}
            {/each}
        {/each}
    </div>
</div>
