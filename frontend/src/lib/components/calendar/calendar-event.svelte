<script lang="ts">
    import type { CalendarEvent, Calendar } from "$lib/types/api";
    import { HoverCard, HoverCardContent, HoverCardTrigger } from "$lib/components/ui/hover-card";
    import { getTextColor } from "$lib/utils";
    import CalendarEventDetails from "./calendar-event-details.svelte";

    interface Props {
        event: CalendarEvent;
        calendar: Calendar;
    }

    let { event, calendar }: Props = $props();

    let backgroundColor = $derived(event.event_color || calendar.event_color || "#3b82f6");
    let textColor = $derived(getTextColor(backgroundColor));
</script>

<HoverCard openDelay={100} closeDelay={100}>
    <HoverCardTrigger class="w-full">
        <div
            class="mb-1 max-w-full cursor-pointer truncate rounded px-1 py-0.5 text-xs transition-opacity hover:opacity-80"
            style="background-color: {backgroundColor}; color: {textColor}"
        >
            {event.title}
        </div>
    </HoverCardTrigger>
    <HoverCardContent class="w-80" side="right">
        <CalendarEventDetails {event} {calendar} />
    </HoverCardContent>
</HoverCard>
