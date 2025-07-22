<script lang="ts">
    import Weekday from "$lib/components/calendar/weekday.svelte";
    import Day from "$lib/components/calendar/day.svelte";
    import calendar from "calendar-js";
    import { Card } from "$lib/components/ui/card";
    import YearMonthSelector from "$lib/components/calendar/year-month-selector.svelte";

    let currentDate = new Date();
    let year: number = $state(currentDate.getFullYear());
    let month: number = $state(currentDate.getMonth()); // 0 = January, 11 = December

    let calendarDays = $derived(calendar().of(year, month).calendar);
</script>

<div class="w-4xl flex-col">
    <!--Calendar Title-->
    <div class="mb-4 flex flex-row items-center justify-between">
        <h1 class="text-4xl font-bold">Calendar Title</h1>
        <YearMonthSelector bind:year bind:month />
    </div>

    <!--Calendar Grid-->
    <div class="grid grid-cols-7 gap-2">
        <!--Top field-->
        <Weekday specialDay>Sunday</Weekday>
        <Weekday>Monday</Weekday>
        <Weekday>Tuesday</Weekday>
        <Weekday>Wednesday</Weekday>
        <Weekday>Thursday</Weekday>
        <Weekday>Friday</Weekday>
        <Weekday specialDay>Saturday</Weekday>

        <!--Bottom field-->
        {#each calendarDays as daysInWeek, index (index)}
            {#each daysInWeek as day, index (index)}
                {#if day !== 0}
                    <Day {day}>
                        <!-- Optional content for each day -->
                        <p>Hello</p>
                    </Day>
                {:else}
                    <!-- Empty day cell for days not in the current month -->
                    <Card class="h-32 p-4" />
                {/if}
            {/each}
        {/each}
    </div>
</div>
