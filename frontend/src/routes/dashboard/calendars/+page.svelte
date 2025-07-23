<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.ts";
    import { Card, CardContent } from "$lib/components/ui/card/index.ts";
    import { Calendar } from "@lucide/svelte";
    import CalendarCard from "$lib/components/dashboard/calendar-card.svelte";
    import AddCalendarDialog from "$lib/components/dashboard/add-calendar-dialog.svelte";
    import { createUserDataQuery } from "$lib/globalQueries";
    import { createQuery } from "@tanstack/svelte-query";
    import { api } from "$lib/api";

    const userDataQuery = createUserDataQuery();

    let importedCalendarQuery = createQuery({
        queryKey: ["imported-calendars"],
        queryFn: () => api.getImportedCalendars()
    });
</script>

<div class="w-2xl max-w-[90vw] space-y-6">
    <div class="hidden md:block">
        <h1 class="text-2xl font-bold tracking-tight">Calendars</h1>
    </div>

    <div class="space-y-4">
        {#if $importedCalendarQuery.isLoading}
            <p class="text-sm text-muted-foreground">Loading calendars...</p>
        {:else if $importedCalendarQuery.isError}
            <p class="text-sm text-red-500">
                Error loading calendars: {$importedCalendarQuery.error.message}
            </p>
        {:else if $importedCalendarQuery.data?.calendars.length === 0 && !$userDataQuery.data?.user}
            <p class="text-sm text-muted-foreground">
                You have no calendars imported. Please import a calendar to get started.
            </p>
        {/if}

        {#each $importedCalendarQuery.data?.calendars! as calendar (calendar.id)}
            <CalendarCard {calendar} />
        {/each}
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
                <AddCalendarDialog user={$userDataQuery.data?.user}>
                    <Button size="sm">Add Calendar</Button>
                </AddCalendarDialog>
            </div>
        </CardContent>
    </Card>
</div>
