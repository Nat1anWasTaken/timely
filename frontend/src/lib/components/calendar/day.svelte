<script lang="ts">
    import { Card } from "$lib/components/ui/card/index.js";
    import { cn } from "$lib/utils";
    import type { Snippet } from "svelte";

    interface Props {
        day: number;
        isCurrentMonth?: boolean;
        children?: Snippet;
    }

    let { day, isCurrentMonth = true, children }: Props = $props();
</script>

<Card class={cn("max flex max-h-24 min-h-30 flex-col gap-0 p-4", !isCurrentMonth && "bg-muted/50")}>
    <p class={cn("text-sm", isCurrentMonth ? "text-muted-foreground" : "text-muted-foreground/50")}>{day}</p>
    <div class="flex flex-1 flex-col items-start justify-start overflow-y-scroll">
        {#if children}
            {@render children()}
        {/if}
    </div>
</Card>
