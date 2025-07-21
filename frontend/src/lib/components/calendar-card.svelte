<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Card,
        CardContent,
        CardDescription,
        CardHeader,
        CardTitle
    } from "$lib/components/ui/card";
    import { Settings, Chrome } from "@lucide/svelte";

    export let title: string;
    export let email: string;
    export let lastSync: string;
    export let color: string = "blue-500";
    export let provider: "google" | "ics" = "google";
</script>

<Card>
    <CardHeader>
        <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
                <div class="h-4 w-4 rounded-full bg-{color}"></div>
                <div>
                    <CardTitle class="text-lg">{title}</CardTitle>
                    <CardDescription class="flex items-center gap-2">
                        {#if provider === "google"}
                            <Chrome class="h-4 w-4" />
                            Google Calendar
                        {:else}
                            Uploaded ICS File
                        {/if}
                    </CardDescription>
                </div>
            </div>
            <Button variant="outline" size="sm">
                <Settings class="mr-2 h-4 w-4" />
                Settings
            </Button>
        </div>
    </CardHeader>
    <CardContent>
        <p class="text-sm text-muted-foreground">
            {email} • {lastSync}
        </p>
    </CardContent>
</Card>
