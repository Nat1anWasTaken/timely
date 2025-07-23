<script lang="ts">
    import CalendarCard from "./calendar-card.svelte";
    import type { Calendar } from "$lib/types/api";

    export let calendars: Calendar[] = [];
    export let isLoading = false;
    export let isError = false;
    export let errorMessage = "";
    export let isUserLoggedIn = false;
</script>

<div class="space-y-4">
    {#if isLoading}
        <p class="text-sm text-muted-foreground">Loading calendars...</p>
    {:else if isError}
        <p class="text-sm text-red-500">
            Error loading calendars: {errorMessage}
        </p>
    {:else if calendars.length === 0 && !isUserLoggedIn}
        <p class="text-sm text-muted-foreground">
            You have no calendars imported. Please import a calendar to get started.
        </p>
    {/if}

    {#each calendars as calendar (calendar.id)}
        <CalendarCard {calendar} />
    {/each}
</div>
