<script lang="ts">
    import { Card } from "$lib/components/ui/card/index.js";
    import { cn } from "$lib/utils";
    import type { Snippet } from "svelte";
    import type { CalendarEvent, Calendar } from "$lib/types/api";
    import CalendarEventComponent from "./calendar-event.svelte";
    import EventsOverflowDialog from "./events-overflow-dialog.svelte";

    interface Props {
        day: number;
        isCurrentMonth?: boolean;
        specialBorderClass?: string;
        dayOfWeek?: string;
        children?: Snippet;
        date?: Date;
        events?: Array<{ event: CalendarEvent; calendar: Calendar }>;
    }

    let { day, isCurrentMonth = true, specialBorderClass, children, dayOfWeek, date, events = [] }: Props = $props();

    // First row (with weekday headers) shows max 2 events, other rows show max 3
    let isFirstRow = $derived(!!dayOfWeek);
    let maxVisibleEvents = $derived(isFirstRow ? 2 : 3);
    
    // Smart overflow logic: if only 1 additional event, show it instead of "+1 more"
    let shouldShowOverflow = $derived(events.length > maxVisibleEvents + 1);
    let visibleEventCount = $derived(shouldShowOverflow ? maxVisibleEvents : events.length);
    let visibleEvents = $derived(events.slice(0, visibleEventCount));
    let remainingCount = $derived(Math.max(0, events.length - maxVisibleEvents));
</script>

<div class={cn("flex flex-col gap-0 rounded-none radis py-2 border max-h-24 min-h-30", specialBorderClass, isCurrentMonth ? "bg-muted/50" : "bg-muted/20")}>
    {#if dayOfWeek}
        <p class={cn("flex items-center justify-center text-[0.7rem] text-muted-foreground")}>{dayOfWeek}</p>
    {/if}
    <p class={cn("flex items-center justify-center text-sm", isCurrentMonth ? "text-muted-foreground" : "text-muted-foreground/50")}>{day}</p>
    <div class="flex flex-col items-start justify-start">
        {#if events.length > 0}
            {#each visibleEvents as { event, calendar } (event.id)}
                <CalendarEventComponent {event} {calendar} />
            {/each}
            {#if shouldShowOverflow && remainingCount > 0 && date}
                <EventsOverflowDialog {date} events={events} {remainingCount} />
            {/if}
        {:else if children}
            {@render children()}
        {/if}
    </div>
</div>
