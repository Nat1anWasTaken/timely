<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import { Label } from "$lib/components/ui/label/index.js";
    import { api } from "$lib/api.js";
    import type { GoogleCalendar } from "$lib/types/api.js";
    import { createQuery, useQueryClient } from "@tanstack/svelte-query";
    import GoogleCalendarItem from "./google-calendar-item.svelte";
    import { toast } from "svelte-sonner";

    interface Props {
        onBack: () => void;
        onSuccess: () => void;
    }

    let { onBack, onSuccess }: Props = $props();

    let selectedCalendar: GoogleCalendar | null = $state(null);
    let selectedColor = $state("#3B82F6");
    let loading = $state(false);

    const queryClient = useQueryClient();

    const googleCalendarsQuery = createQuery({
        queryKey: ["google-calendars"],
        queryFn: () => api.getGoogleCalendars()
    });

    const importedCalendarsQuery = createQuery({
        queryKey: ["imported-calendars"],
        queryFn: () => api.getImportedCalendars()
    });

    let importedGoogleCalendarIds = $derived(
        $importedCalendarsQuery.data?.calendars
            ?.filter((cal) => cal.source === "google")
            ?.map((cal) => cal.source_id) || []
    );

    function selectCalendar(calendar: GoogleCalendar) {
        if (importedGoogleCalendarIds.includes(calendar.id)) {
            return;
        }

        selectedCalendar = calendar;
        // Use calendar's existing color if available
        if (calendar.backgroundColor) {
            selectedColor = calendar.backgroundColor;
        }
    }

    function handleColorChange(color: string) {
        selectedColor = color;
    }

    async function handleImport() {
        if (!selectedCalendar) return;

        loading = true;

        try {
            const response = await api.importGoogleCalendar({
                calendar_id: selectedCalendar.id
            });

            if (response.success) {
                toast.success("Google calendar imported successfully!");
                // Invalidate queries to refresh the data
                queryClient.invalidateQueries({ queryKey: ["imported-calendars"] });
                onSuccess();
            } else {
                toast.error(response.message || "Failed to import Google calendar");
            }
        } catch (err) {
            toast.error(err instanceof Error ? err.message : "An error occurred");
        } finally {
            loading = false;
        }
    }
</script>

<div class="space-y-4">
    <div class="mx-4 flex items-center space-x-4">
        <Button variant="ghost" size="sm" onclick={onBack}>← Back</Button>
    </div>

    {#if $googleCalendarsQuery.isLoading}
        <div class="flex items-center justify-center py-4">
            <div class="text-sm text-muted-foreground">Loading calendars...</div>
        </div>
    {:else if $googleCalendarsQuery.error}
        <div class="text-sm text-destructive p-4">
            Failed to load Google calendars: {$googleCalendarsQuery.error.message}
        </div>
    {:else if $googleCalendarsQuery.data?.calendars}
        <div class="space-y-3 p-4">
            <Label class="text-sm font-medium">Select a Google calendar</Label>
            
            <!-- Scrollable calendar list -->
            <div class="max-h-60 overflow-y-auto space-y-3 p-0.5">
                {#each $googleCalendarsQuery.data.calendars as calendar}
                    <GoogleCalendarItem
                        {calendar}
                        isSelected={selectedCalendar?.id === calendar.id}
                        isImported={importedGoogleCalendarIds.includes(calendar.id)}
                        onSelect={selectCalendar}
                    />
                {/each}
            </div>

            {#if selectedCalendar}
                <div class="space-y-4 pt-4 border-t">
                    <div class="space-y-2">
                        <Label for="color-picker" class="text-sm font-medium">Calendar color</Label>
                        <div class="flex items-center space-x-2">
                            <input
                                id="color-picker"
                                type="color"
                                value={selectedColor}
                                onchange={(e) => handleColorChange(e.currentTarget.value)}
                                class="h-8 w-8 cursor-pointer rounded border"
                            />
                            <span class="text-sm text-muted-foreground">{selectedColor}</span>
                        </div>
                    </div>

                    <Button class="w-full" onclick={handleImport} disabled={loading}>
                        {#if loading}
                            <div
                                class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                            ></div>
                        {/if}
                        Import Calendar
                    </Button>
                </div>
            {/if}
        </div>
    {/if}
</div>
