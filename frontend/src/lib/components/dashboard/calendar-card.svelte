<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Card,
        CardContent,
        CardDescription,
        CardHeader,
        CardTitle
    } from "$lib/components/ui/card";
    import { Calendar, Settings } from "@lucide/svelte";
    import type { Snippet } from "svelte";

    interface Props {
        title: string;
        email: string;
        lastSync: string;
        color: string;
        icon?: Snippet;
    }

    let { title, email, lastSync, color, icon }: Props = $props();
</script>

<Card>
    <CardHeader>
        <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
                <div class="h-4 w-4 rounded-full bg-{color}">
                    {#if icon}
                        {@render icon()}
                    {:else}
                        <Calendar class="h-4 w-4" />
                    {/if}
                </div>
                <div>
                    <CardTitle class="text-lg">{title}</CardTitle>
                    <CardDescription class="flex items-center gap-2"></CardDescription>
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
