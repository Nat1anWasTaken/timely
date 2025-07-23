<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import {
        Sheet,
        SheetContent,
        SheetDescription,
        SheetHeader,
        SheetTitle,
        SheetTrigger
    } from "$lib/components/ui/sheet/index.js";
    import type { User } from "$lib/types/api.js";
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

<Sheet bind:open onOpenChange={() => resetState()}>
    <SheetTrigger>
        {@render children()}
    </SheetTrigger>
    <SheetContent class="sm:max-w-md">
        <SheetHeader>
            <SheetTitle>Add Calendar</SheetTitle>
            <SheetDescription>
                Connect a Google calendar or upload an ICS file to add events to your dashboard.
            </SheetDescription>
        </SheetHeader>

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
    </SheetContent>
</Sheet>
