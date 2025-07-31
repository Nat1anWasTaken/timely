<script lang="ts">
    import { Card } from "$lib/components/ui/card/index.js";
    import { cn } from "$lib/utils";
    import type { Snippet } from "svelte";

    interface Props {
        day: number;
        isCurrentMonth?: boolean;
        specialBorderClass?: string;
        children?: Snippet;
    }

    let { day, isCurrentMonth = true, specialBorderClass, children }: Props = $props();
</script>

<Card class={cn("max flex max-h-24 min-h-30 flex-col gap-0 rounded-none radis py-2", specialBorderClass, !isCurrentMonth && "bg-muted/50")}>
    <p class={cn("flex items-center justify-center text-sm", isCurrentMonth ? "text-muted-foreground" : "text-muted-foreground/50")}>{day}</p>
    <div class="flex flex-1 flex-col items-start justify-start overflow-y-scroll">
        {#if children}
            {@render children()}
        {/if}
    </div>
</Card>
