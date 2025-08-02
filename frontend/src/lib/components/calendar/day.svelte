<script lang="ts">
    import { Card } from "$lib/components/ui/card/index.js";
    import { cn } from "$lib/utils";
    import type { Snippet } from "svelte";

    interface Props {
        day: number;
        isCurrentMonth?: boolean;
        specialBorderClass?: string;
        dayOfWeek?: string;
        children?: Snippet;
    }

    let { day, isCurrentMonth = true, specialBorderClass, children, dayOfWeek }: Props = $props();
</script>

<div class={cn("flex flex-col gap-0 rounded-none radis py-2 border max-h-24 min-h-30", specialBorderClass, isCurrentMonth ? "bg-muted/50" : "bg-muted/20")}>
    {#if dayOfWeek}
        <p class={cn("flex items-center justify-center text-[0.7rem] text-muted-foreground")}>{dayOfWeek}</p>
    {/if}
    <p class={cn("flex items-center justify-center text-sm", isCurrentMonth ? "text-muted-foreground" : "text-muted-foreground/50")}>{day}</p>
    <div class="flex flex-col items-start justify-start">
        {#if children}
            {@render children()}
        {/if}
    </div>
</div>
