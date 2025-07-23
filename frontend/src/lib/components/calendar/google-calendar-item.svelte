<script lang="ts">
    import {
        Card,
        CardContent,
        CardDescription,
        CardTitle
    } from "$lib/components/ui/card/index.js";
    import { CheckCircle2 } from "@lucide/svelte";
    import type { GoogleCalendar } from "$lib/types/api.js";

    interface Props {
        calendar: GoogleCalendar;
        isSelected: boolean;
        isImported: boolean;
        onSelect: (calendar: GoogleCalendar) => void;
    }

    let { calendar, isSelected, isImported, onSelect }: Props = $props();
</script>

<Card
    class="transition-colors {isImported
        ? 'cursor-not-allowed opacity-60'
        : 'cursor-pointer hover:bg-accent/50'} {isSelected ? 'ring-2 ring-primary' : ''}"
    onclick={() => !isImported && onSelect(calendar)}
>
    <CardContent class="px-3 py-2">
        <div class="flex items-center space-x-2.5">
            <div
                class="h-2.5 w-2.5 rounded-full flex-shrink-0"
                style="background-color: {calendar.backgroundColor || '#3B82F6'}"
            ></div>
            <div class="min-w-0 flex-1">
                <CardTitle class="truncate text-sm leading-none">{calendar.summary}</CardTitle>
                <div class="mt-0.5 flex items-center text-xs text-muted-foreground leading-none">
                    <span>{calendar.timeZone}</span>
                    {#if calendar.primary}
                        <span class="mx-1">•</span>
                        <span class="text-primary">Primary</span>
                    {/if}
                    {#if isImported}
                        <span class="mx-1">•</span>
                        <span>Imported</span>
                    {/if}
                </div>
            </div>
        </div>
    </CardContent>
</Card>
