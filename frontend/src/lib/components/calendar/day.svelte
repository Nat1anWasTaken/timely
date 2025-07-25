<script lang="ts">
    import { Card } from "$lib/components/ui/card/index.js";
    import type { Snippet } from "svelte";

    interface Props {
        day: number;
        children?: Snippet;
    }

    let { day, children }: Props = $props();
</script>

<Card class="max flex max-h-24 min-h-30 flex-col gap-0 p-4">
    <p class="text-sm text-muted-foreground">{day}</p>
    <div class="flex flex-1 flex-col items-start justify-start overflow-y-scroll">
        {#if children}
            {@render children()}
        {/if}
    </div>
</Card>
