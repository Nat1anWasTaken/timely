<script lang="ts">
    import {
        Card,
        CardContent,
        CardDescription,
        CardTitle
    } from "$lib/components/ui/card/index.js";
    import { Label } from "$lib/components/ui/label/index.js";
    import { Calendar, Upload, AlertCircle } from "@lucide/svelte";
    import type { User } from "$lib/types/api.js";

    interface Props {
        user?: User;
        onProviderSelect: (provider: "google" | "ics") => void;
    }

    let { user, onProviderSelect }: Props = $props();

    // Check if user has Google account
    let hasGoogleAccount = $derived(
        user?.accounts?.some((account) => account.provider === "google") ?? false
    );
</script>

<div class="space-y-3 p-4">
    <Label class="text-sm font-medium">Choose calendar source</Label>

    <!-- Google Calendar Option -->
    <Card
        class="cursor-pointer transition-colors hover:bg-accent/50 {!hasGoogleAccount
            ? 'opacity-50'
            : ''}"
        onclick={() => hasGoogleAccount && onProviderSelect("google")}
    >
        <CardContent class="flex items-center space-x-3 p-4">
            <Calendar class="h-5 w-5 text-blue-600" />
            <div class="flex-1">
                <CardTitle class="text-sm">Google Calendar</CardTitle>
                <CardDescription class="text-xs">
                    {hasGoogleAccount
                        ? "Import calendars from your Google account"
                        : "Google account required - link one in your profile"}
                </CardDescription>
            </div>
            {#if !hasGoogleAccount}
                <AlertCircle class="h-4 w-4 text-muted-foreground" />
            {/if}
        </CardContent>
    </Card>

    <!-- ICS Upload Option -->
    <Card
        class="cursor-pointer transition-colors hover:bg-accent/50"
        onclick={() => onProviderSelect("ics")}
    >
        <CardContent class="flex items-center space-x-3 p-4">
            <Upload class="h-5 w-5 text-green-600" />
            <div class="flex-1">
                <CardTitle class="text-sm">Upload ICS File</CardTitle>
                <CardDescription class="text-xs">
                    Import from any calendar application (.ics file)
                </CardDescription>
            </div>
        </CardContent>
    </Card>
</div>
