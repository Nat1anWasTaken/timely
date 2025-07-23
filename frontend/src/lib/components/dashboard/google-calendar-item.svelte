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
        onSelect: (calendar: GoogleCalendar) => void;
    }

    let { calendar, isSelected, onSelect }: Props = $props();
</script>

<Card
    class="cursor-pointer transition-colors hover:bg-accent/50 {isSelected
        ? 'ring-2 ring-primary'
        : ''}"
    onclick={() => onSelect(calendar)}
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
                <div class="mt-2 flex items-center space-x-2 text-xs text-muted-foreground">
                    <span>{calendar.accessRole}</span>
                    {#if calendar.primary}
                        <span class="rounded bg-primary/10 px-1.5 py-0.5 text-primary">Primary</span
                        >
                    {/if}
                </div>
            </div>
        </div>
    </CardContent>
</Card>
