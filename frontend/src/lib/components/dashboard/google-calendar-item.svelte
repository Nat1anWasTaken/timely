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
    <CardContent class="p-4">
        <div class="flex items-start space-x-3">
            <div
                class="mt-0.5 h-4 w-4 rounded-full"
                style="background-color: {calendar.backgroundColor || '#3B82F6'}"
            ></div>
            <div class="min-w-0 flex-1">
                <CardTitle class="truncate text-sm">{calendar.summary}</CardTitle>
                {#if calendar.description}
                    <CardDescription class="mt-1 line-clamp-2 text-xs">
                        {calendar.description}
                    </CardDescription>
                {/if}
                <div class="mt-2 flex items-center text-xs text-muted-foreground">
                    <span>{calendar.timeZone}</span>
                    <span class="mx-1">•</span>
                    {#if calendar.primary}
                        <span class="text-primary">Primary</span>
                    {/if}
                    {#if isImported}
                        <span>Imported</span>
                    {/if}
                </div>
            </div>
        </div>
    </CardContent>
</Card>
