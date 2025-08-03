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

    // Helper function to get corner rounding classes based on position
    function getCornerRoundingClass(rowIndex: number, colIndex: number) {
        const isFirstRow = rowIndex === 0;
        const isLastRow = rowIndex === calendarDays.length - 1;
        const isFirstCol = colIndex === 0;
        const isLastCol = colIndex === 6;

        if (isFirstRow && isFirstCol) return "rounded-tl-lg";
        if (isFirstRow && isLastCol) return "rounded-tr-lg";
        if (isLastRow && isFirstCol) return "rounded-bl-lg";
        if (isLastRow && isLastCol) return "rounded-br-lg";
        return "";
    }

    function getDayOfWeek(rowIndex: number, colIndex: number) {
        if (rowIndex !== 0) {
            return "";
        }

        if (colIndex === 0) {
            return "Sun";
        } else if (colIndex === 1) {
            return "Mon";
        } else if (colIndex === 2) {
            return "Tue";
        } else if (colIndex === 3) {
            return "Wed";
        } else if (colIndex === 4) {
            return "Thu";
        } else if (colIndex === 5) {
            return "Fri";
        } else if (colIndex === 6) {
            return "Sat";
        }
    }
</script>

<div class={cn("flex h-full w-full flex-col", className)}>
    <!--Calendar Header-->
    <div class="mb-4 flex flex-row items-center justify-end">
        <YearMonthSelector bind:year bind:month />
    </div>

    <!--Calendar Grid-->
    <div class="grid grid-cols-7" style="grid-template-rows: auto repeat(6, 1fr);">
        <!--Calendar Days-->
        {#each calendarDays as daysInWeek, rowIndex (rowIndex)}
            {#each daysInWeek as dayInfo, colIndex (colIndex)}
                <Day
                    day={dayInfo.day}
                    isCurrentMonth={dayInfo.isInPrimaryMonth}
                    specialBorderClass={getCornerRoundingClass(rowIndex, colIndex)}
                    dayOfWeek={getDayOfWeek(rowIndex, colIndex)}
                    date={dayInfo.date}
                    events={getEventsForDate(dayInfo.date)}
                />
            {/each}
        {/each}
    </div>
</div>
