<script lang="ts">
    import type { CalendarEvent, Calendar } from "$lib/types/api";
    import { getSourceString, formatDateTime, formatDate, formatTime } from "$lib/utils";

    interface Props {
        event: CalendarEvent;
        calendar: Calendar;
    }

    let { event, calendar }: Props = $props();

    let startDate = $derived(new Date(event.start));
    let endDate = $derived(new Date(event.end));
    let isSameDay = $derived(startDate.toDateString() === endDate.toDateString());
    let showFullDescription = $state(false);
</script>

<div class="space-y-3 p-4">
    <!-- Event Title -->
    <div class="space-y-1">
        <h3 class="font-semibold text-lg leading-tight">{event.title}</h3>
        <p class="text-sm text-muted-foreground">{calendar.summary}</p>
    </div>

    <!-- Date and Time -->
    <div class="space-y-1">
        <div class="flex items-center gap-2">
            <svg class="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <div class="text-sm">
                {#if event.all_day}
                    {#if isSameDay}
                        <span>{formatDate(event.start)}</span>
                    {:else}
                        <span>{formatDate(event.start)} - {formatDate(event.end)}</span>
                    {/if}
                    <span class="text-muted-foreground ml-2">(All day)</span>
                {:else}
                    {#if isSameDay}
                        <div>{formatDate(event.start)}</div>
                        <div class="text-muted-foreground">
                            {formatTime(event.start)} - {formatTime(event.end)}
                        </div>
                    {:else}
                        <div>{formatDateTime(event.start)}</div>
                        <div class="text-muted-foreground">to</div>
                        <div>{formatDateTime(event.end)}</div>
                    {/if}
                {/if}
            </div>
        </div>
    </div>

    <!-- Location -->
    {#if event.location}
        <div class="space-y-1">
            <div class="flex items-center gap-2">
                <svg class="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span class="text-sm">{event.location}</span>
            </div>
        </div>
    {/if}

    <!-- Description -->
    {#if event.description}
        <div class="space-y-1">
            <div class="flex items-start gap-2">
                <svg class="h-4 w-4 text-muted-foreground mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
                </svg>
                <div class="text-sm text-muted-foreground leading-relaxed">
                    {#if event.description.length > 150 && !showFullDescription}
                        {@html event.description.substring(0, 150)}...
                        <button 
                            class="text-primary hover:text-primary/80 ml-1 underline text-xs"
                            onclick={() => showFullDescription = true}
                        >
                            Show more
                        </button>
                    {:else}
                        {@html event.description}
                        {#if event.description.length > 150 && showFullDescription}
                            <button 
                                class="text-primary hover:text-primary/80 ml-1 underline text-xs block mt-1"
                                onclick={() => showFullDescription = false}
                            >
                                Show less
                            </button>
                        {/if}
                    {/if}
                </div>
            </div>
        </div>
    {/if}

    <!-- Event Color Indicator -->
    <div class="flex items-center gap-2 pt-2 border-t">
        <div 
            class="w-3 h-3 rounded-full" 
            style="background-color: {event.event_color || calendar.event_color || '#3b82f6'}"
        ></div>
        <span class="text-xs text-muted-foreground">
            {getSourceString(calendar.source)}
        </span>
    </div>
</div>