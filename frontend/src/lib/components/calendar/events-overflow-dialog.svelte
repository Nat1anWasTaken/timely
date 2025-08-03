<script lang="ts">
    import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "$lib/components/ui/dialog";
    import CalendarEventComponent from "./calendar-event.svelte";
    import type { CalendarEvent, Calendar } from "$lib/types/api";
    import { formatDate } from "$lib/utils";

    interface Props {
        date: Date;
        events: Array<{ event: CalendarEvent; calendar: Calendar }>;
        remainingCount: number;
    }

    let { date, events, remainingCount }: Props = $props();
</script>

<Dialog>
    <DialogTrigger class="w-full">
        <div class="mb-1 cursor-pointer rounded bg-muted/50 px-1 py-0.5 text-xs text-muted-foreground transition-colors hover:bg-muted/70">
            +{remainingCount} more
        </div>
    </DialogTrigger>
    <DialogContent class="max-w-md max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader>
            <DialogTitle>Events for {formatDate(date.toISOString())}</DialogTitle>
        </DialogHeader>
        <div class="flex-1 overflow-y-auto space-y-2 p-4">
            {#each events as { event, calendar } (event.id)}
                <CalendarEventComponent {event} {calendar} />
            {/each}
        </div>
    </DialogContent>
</Dialog>