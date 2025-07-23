<script lang="ts">
    import { Button } from "$lib/components/ui/button";

    interface Props {
        icon: Snippet;
        title: string;
        description: string;
    }

    let { icon, title, description } = $props();
</script>

<div class="flex items-center justify-between rounded-lg border p-3">
    <div class="flex items-center space-x-3">
        {@render icon()}
        <div>
            <p class="text-sm font-medium">{title}</p>
            <p class="text-xs text-muted-foreground">{description}</p>
        </div>
    </div>
    <Button variant="outline" size="sm" disabled>Disconnect</Button>
</div>
