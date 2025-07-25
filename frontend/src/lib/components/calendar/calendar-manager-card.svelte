<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Card, CardContent, CardHeader } from "$lib/components/ui/card";
    import type { Calendar } from "$lib/types/api";
    import { getSourceString } from "$lib/utils";
    import { Calendar as CalendarIcon, Settings } from "@lucide/svelte";
    import CalendarSettingsDialog from "./calendar-settings-dialog.svelte";

    interface Props {
        calendar: Calendar;
        onUpdate?: (calendar: Calendar) => void;
    }

    let { calendar, onUpdate }: Props = $props();
    let showSettingsDialog = $state(false);
</script>

<Card class="mb-0 justify-center">
    <CardHeader>
        <div class="flex items-center justify-between">
            <div class="flex min-w-0 flex-1 items-center space-x-3">
                <div
                    class="h-3 w-3 flex-shrink-0 rounded-full"
                    style="background-color: {calendar.event_color || '#3B82F6'}"
                ></div>
                <div class="min-w-0 flex-1">
                    <h3 class="truncate text-sm font-medium">{calendar.summary}</h3>
                    <p class="mt-0.5 text-xs text-muted-foreground">
                        {getSourceString(calendar.source)} • last synced {calendar.synced_at
                            ? new Date(calendar.synced_at).toLocaleDateString()
                            : "never"}
                    </p>
                </div>
            </div>
            <Button variant="ghost" size="sm" class="flex-shrink-0" onclick={() => showSettingsDialog = true}>
                <Settings class="h-4 w-4" />
            </Button>
        </div>
    </CardHeader>
</Card>

<CalendarSettingsDialog 
    {calendar} 
    open={showSettingsDialog} 
    onOpenChange={(open) => showSettingsDialog = open}
    {onUpdate}
/>
