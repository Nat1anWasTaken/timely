<script lang="ts">
    import { Card } from "$lib/components/ui/card/index.js";
    import { cn } from "$lib/utils.ts";
    import type { Snippet } from "svelte";
    interface Props {
        children?: Snippet;
        specialDay?: boolean;
    }

    let { children, specialDay = false }: Props = $props();
</script>

<Card
    class={cn(
        "flex min-h-12 items-center justify-center",
        specialDay ? "bg-primary text-primary-foreground" : ""
    )}
>
    {#if children}
        {@render children()}
    {/if}
</Card>
