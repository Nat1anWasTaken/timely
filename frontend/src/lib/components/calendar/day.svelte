<script lang="ts">
    import { Card } from "$lib/components/ui/card/index.js";
    import type { Snippet } from "svelte";

    interface Props {
        day: number;
        children?: Snippet;
    }

    let { day, children }: Props = $props();
</script>

<Card class="h-32 gap-0 p-4">
    <p class="text-sm text-muted-foreground">{day}</p>
    <div class="flex h-full flex-col items-center justify-start">
        {#if children}
            {@render children()}
        {/if}
    </div>
</Card>
