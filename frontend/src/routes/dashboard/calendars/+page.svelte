<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.ts";
    import { Card, CardContent } from "$lib/components/ui/card/index.ts";
    import { Calendar } from "@lucide/svelte";
    import CalendarCard from "$lib/components/calendar-card.svelte";
</script>

<div class="w-2xl max-w-[90vw] space-y-6">
    <div class="hidden md:block">
        <h1 class="text-2xl font-bold tracking-tight">Calendars</h1>
    </div>

    <div class="space-y-4">
        <CalendarCard
            title="Work Calendar"
            email="john.doe@company.com"
            lastSync="Synced 5 minutes ago"
            color="blue-500"
            provider="google"
        />

        <CalendarCard
            title="Personal Calendar"
            email="john.doe@gmail.com"
            lastSync="Synced 2 minutes ago"
            color="green-500"
            provider="google"
        />

        <CalendarCard
            title="Team Events"
            email="team-events.ics"
            lastSync="Last updated 1 day ago"
            color="purple-500"
            provider="ics"
        />
    </div>

    <Card class="border-dashed">
        <CardContent class="pt-6">
            <div class="flex flex-col items-center justify-center space-y-3 text-center">
                <Calendar class="h-8 w-8 text-muted-foreground" />
                <div>
                    <p class="text-sm font-medium">Add a new calendar</p>
                    <p class="text-xs text-muted-foreground">
                        Connect Google, Outlook, or upload an ICS file
                    </p>
                </div>
                <Button size="sm">Add Calendar</Button>
            </div>
        </CardContent>
    </Card>
</div>
