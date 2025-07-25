<script lang="ts">
    import { api } from "$lib/api";
    import type { Calendar } from "$lib/types/api";
    import { createQuery } from "@tanstack/svelte-query";
    import CalendarManagerCard from "./calendar-manager-card.svelte";

    const importedCalendarsQuery = createQuery({
        queryKey: ["imported-calendars"],
        queryFn: () => api.getImportedCalendars()
    });

    let calendars = $derived($importedCalendarsQuery.data?.calendars || []);

    function handleCalendarUpdate(updatedCalendar: Calendar) {
        // Query cache will be automatically invalidated by the mutation
    }
</script>

<div class="space-y-4">
    {#if $importedCalendarsQuery.isLoading}
        <div class="flex items-center justify-center py-8">
            <div class="text-sm text-muted-foreground">Loading calendars...</div>
        </div>
    {:else if $importedCalendarsQuery.error}
        <div class="rounded bg-destructive/10 p-4 text-sm text-destructive">
            Error loading calendars: {$importedCalendarsQuery.error.message}
        </div>
    {:else if calendars.length === 0}
        <div class="flex flex-col items-center justify-center py-8 text-center">
            <p class="text-sm text-muted-foreground">You haven't imported any calendars yet.</p>
            <p class="mt-1 text-xs text-muted-foreground">
                Click "Import Calendar" above to get started.
            </p>
        </div>
    {:else}
        <div class="flex flex-col gap-3 space-y-3">
            {#each calendars as calendar (calendar.id)}
                <CalendarManagerCard {calendar} onUpdate={handleCalendarUpdate} />
            {/each}
        </div>
    {/if}
</div>
