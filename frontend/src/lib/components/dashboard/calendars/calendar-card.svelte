<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Card,
        CardContent,
        CardDescription,
        CardHeader,
        CardTitle
    } from "$lib/components/ui/card";
    import type { Calendar } from "$lib/types/api";
    import { getSourceString } from "$lib/utils";
    import { Calendar as CalendarIcon } from "@lucide/svelte";
    import { Settings } from "@lucide/svelte";
    import type { Component } from "svelte";

    interface Props {
        calendar: Calendar;
    }

    let { calendar }: Props = $props();
</script>

<Card class="gap-3">
    <CardHeader>
        <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
                <div class="h-4 w-4 rounded-full text-[{calendar.event_color}]">
                    <CalendarIcon class="h-4 w-4" />
                </div>
                <div>
                    <CardTitle class="text-lg">{calendar.summary}</CardTitle>
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
            {getSourceString(calendar.source)} • last synced {calendar.synced_at
                ? new Date(calendar.synced_at).toLocaleDateString()
                : "never"}
        </p>
    </CardContent>
</Card>
