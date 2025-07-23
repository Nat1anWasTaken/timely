<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import type { Component } from "svelte";

    interface Props {
        icon: Component;
        title: string;
        description: string;
    }

    let { icon, title, description } = $props();
</script>

<div class="flex items-center justify-between rounded-lg border p-3">
    <div class="flex items-center space-x-3">
        <svelte:component this={icon} class="h-5 w-5 text-muted-foreground" />
        <div>
            <p class="text-sm font-medium">{title}</p>
            <p class="text-xs text-muted-foreground">{description}</p>
        </div>
    </div>
    <Button variant="outline" size="sm" disabled>Disconnect</Button>
</div>
