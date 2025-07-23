<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Dialog,
        DialogContent,
        DialogDescription,
        DialogHeader,
        DialogTitle,
        DialogTrigger
    } from "$lib/components/ui/dialog";
    import type { User } from "$lib/types/api";
    import CalendarProviderSelection from "./calendar-provider-selection.svelte";
    import GoogleCalendarSelection from "./google-calendar-selection.svelte";
    import IcsUpload from "./ics-upload.svelte";

    interface Props {
        user?: User;
        children: import("svelte").Snippet;
    }

    let { user, children }: Props = $props();

    let open = $state(false);
    let selectedProvider: "google" | "ics" | null = $state(null);

    function resetState() {
        selectedProvider = null;
    }

    function selectProvider(provider: "google" | "ics") {
        selectedProvider = provider;
    }

    function handleSuccess() {
        open = false;
        resetState();
    }
</script>

<Dialog bind:open onOpenChange={() => resetState()}>
    <DialogTrigger>
        {@render children()}
    </DialogTrigger>
    <DialogContent class="sm:max-w-md">
        <DialogHeader>
            <DialogTitle>Import Calendar</DialogTitle>
            <DialogDescription>
                Connect a Google calendar or upload an ICS file to add events to your profile.
            </DialogDescription>
        </DialogHeader>

        <div class="space-y-4 py-4">
            {#if !selectedProvider}
                <CalendarProviderSelection {user} onProviderSelect={selectProvider} />
            {:else if selectedProvider === "google"}
                <GoogleCalendarSelection
                    onBack={() => (selectedProvider = null)}
                    onSuccess={handleSuccess}
                />
            {:else if selectedProvider === "ics"}
                <IcsUpload onBack={() => (selectedProvider = null)} onSuccess={handleSuccess} />
            {/if}
        </div>
    </DialogContent>
</Dialog>