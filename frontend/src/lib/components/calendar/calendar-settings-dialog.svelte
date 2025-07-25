<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Dialog,
        DialogContent,
        DialogHeader,
        DialogTitle,
        DialogFooter
    } from "$lib/components/ui/dialog";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import { Switch } from "$lib/components/ui/switch";
    import { Textarea } from "$lib/components/ui/textarea";
    import type { Calendar, CalendarUpdateRequest } from "$lib/types/api";
    import { api } from "$lib/api";
    import { toast } from "svelte-sonner";
    import { createMutation, useQueryClient } from "@tanstack/svelte-query";

    interface Props {
        calendar: Calendar;
        open: boolean;
        onOpenChange: (open: boolean) => void;
        onUpdate?: (calendar: Calendar) => void;
    }

    let { calendar, open, onOpenChange, onUpdate }: Props = $props();

    let formData = $state({
        visibility: calendar.visibility,
        description: calendar.description || "",
        eventColor: calendar.event_color || "#3B82F6",
        eventRedaction: calendar.event_redaction || "",
        timeZone: calendar.time_zone || "UTC"
    });

    const isExternalSource = calendar.source !== "isc";
    const queryClient = useQueryClient();

    const updateCalendarMutation = createMutation({
        mutationFn: async (updateRequest: CalendarUpdateRequest) => {
            return api.updateCalendar(calendar.id, updateRequest);
        },
        onSuccess: (response) => {
            if (response.success) {
                toast.success("Calendar settings updated successfully");
                onUpdate?.(response.calendar);
                onOpenChange(false);
                // Invalidate and refetch calendar queries
                queryClient.invalidateQueries({ queryKey: ["imported-calendars"] });
            } else {
                toast.error("Failed to update calendar settings");
            }
        },
        onError: (error) => {
            console.error("Error updating calendar:", error);
            toast.error("Failed to update calendar settings");
        }
    });

    function handleSubmit() {
        const updateRequest: CalendarUpdateRequest = {
            visibility: formData.visibility
        };

        if (!isExternalSource) {
            updateRequest.description = formData.description;
            updateRequest.event_color = formData.eventColor;
            updateRequest.event_redaction = formData.eventRedaction;
            updateRequest.time_zone = formData.timeZone;
        }

        $updateCalendarMutation.mutate(updateRequest);
    }

    function handleCancel() {
        formData = {
            visibility: calendar.visibility,
            description: calendar.description || "",
            eventColor: calendar.event_color || "#3B82F6",
            eventRedaction: calendar.event_redaction || "",
            timeZone: calendar.time_zone || "UTC"
        };
        onOpenChange(false);
    }
</script>

<Dialog {open} {onOpenChange}>
    <DialogContent class="sm:max-w-[425px]">
        <DialogHeader>
            <DialogTitle>Calendar Settings</DialogTitle>
        </DialogHeader>

        {#if isExternalSource}
            <div class="rounded-md bg-destructive p-3 text-sm text-destructive-foreground">
                <strong>Note:</strong> This calendar is from an external source ({calendar.source}).
                Only visibility settings can be modified.
            </div>
        {/if}

        <div class="space-y-6 py-4">
            <div class="space-y-2">
                <Label for="calendar-name">Name</Label>
                <Input id="calendar-name" value={calendar.summary} disabled />
            </div>

            <div class="space-y-2">
                <Label for="visibility">Visibility</Label>
                <div class="flex items-center space-x-3">
                    <Switch
                        id="visibility"
                        checked={formData.visibility === "public"}
                        onCheckedChange={(checked) =>
                            (formData.visibility = checked ? "public" : "private")}
                    />
                    <span class="text-sm text-muted-foreground">
                        {formData.visibility === "public" ? "Public" : "Private"}
                    </span>
                </div>
            </div>

            <div class="space-y-2">
                <Label for="description">Description</Label>
                <Textarea
                    id="description"
                    placeholder="Enter calendar description"
                    bind:value={formData.description}
                    disabled={isExternalSource}
                />
            </div>

            <div class="space-y-2">
                <Label for="event-color">Event Color</Label>
                <div class="flex items-center space-x-2">
                    <Input
                        id="event-color"
                        type="color"
                        bind:value={formData.eventColor}
                        class="h-10 w-16"
                        disabled={isExternalSource}
                    />
                    <Input
                        bind:value={formData.eventColor}
                        placeholder="#3B82F6"
                        class="flex-1"
                        disabled={isExternalSource}
                    />
                </div>
            </div>

            <div class="space-y-2">
                <Label for="event-redaction">Event Redaction</Label>
                <Input
                    id="event-redaction"
                    placeholder="e.g., [PRIVATE]"
                    bind:value={formData.eventRedaction}
                    disabled={isExternalSource}
                />
            </div>

            <div class="space-y-2">
                <Label for="time-zone">Time Zone</Label>
                <Input
                    id="time-zone"
                    bind:value={formData.timeZone}
                    placeholder="e.g., America/New_York"
                    disabled={isExternalSource}
                />
            </div>
        </div>

        <DialogFooter>
            <Button variant="outline" onclick={handleCancel} disabled={$updateCalendarMutation.isPending}>Cancel</Button>
            <Button onclick={handleSubmit} disabled={$updateCalendarMutation.isPending}>
                {$updateCalendarMutation.isPending ? "Saving..." : "Save Changes"}
            </Button>
        </DialogFooter>
    </DialogContent>
</Dialog>
