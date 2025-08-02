<script lang="ts">
    import type { CalendarEvent, Calendar } from "$lib/types/api";
    import { Dialog, DialogContent, DialogTrigger } from "$lib/components/ui/dialog";
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

<Dialog>
    <DialogTrigger class="w-full">
        {#if event.all_day}
            <div
                class="mb-1 max-w-full cursor-pointer truncate rounded px-1 py-0.5 text-xs text-left transition-opacity hover:opacity-80"
                style="background-color: {backgroundColor}; color: {textColor}"
            >
                {event.title}
            </div>
        {:else}
            <div
                class="mb-1 flex max-w-full cursor-pointer items-center gap-1 text-xs transition-opacity hover:opacity-80"
            >
                <div
                    class="h-2 w-2 shrink-0 rounded-full"
                    style="background-color: {backgroundColor}"
                ></div>
                <span
                    class="truncate text-left"
                >
                    {event.title}
                </span>
            </div>
        {/if}
    </DialogTrigger>
    <DialogContent class="max-w-md">
        <CalendarEventDetails {event} {calendar} />
    </DialogContent>
</Dialog>
