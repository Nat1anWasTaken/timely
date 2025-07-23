<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import { Input } from "$lib/components/ui/input/index.js";
    import { Label } from "$lib/components/ui/label/index.js";
    import { FileText, CheckCircle2 } from "@lucide/svelte";
    import { api } from "$lib/api.js";
    import { toast } from "svelte-sonner";

    interface Props {
        onBack: () => void;
        onSuccess: () => void;
    }

    let { onBack, onSuccess }: Props = $props();

    // Internal state
    let icsFile: File | null = $state(null);
    let calendarName = $state("");
    let loading = $state(false);
    let error = $state("");

    function handleFileUpload(event: Event) {
        const target = event.target as HTMLInputElement;
        const file = target.files?.[0] || null;

        if (file && file.type === "text/calendar") {
            icsFile = file;
            // Auto-generate calendar name from filename if not set
            if (!calendarName) {
                calendarName = file.name.replace(/\.ics$/i, "");
            }
            error = "";
        } else if (file) {
            error = "Please select a valid ICS file";
            icsFile = null;
        } else {
            icsFile = null;
        }
    }

    function handleNameChange(name: string) {
        calendarName = name;
    }

    async function handleImport() {
        if (!icsFile) return;

        loading = true;
        error = "";

        try {
            const response = await api.importICSFileUpload(icsFile, calendarName || undefined);
            
            if (response.success) {
                toast.success(`ICS calendar imported successfully! ${response.events_count || 0} events added.`);
                onSuccess();
            } else {
                toast.error(response.message || "Failed to import ICS file");
            }
        } catch (err) {
            toast.error(err instanceof Error ? err.message : "An error occurred");
        } finally {
            loading = false;
        }
    }
</script>

<div class="space-y-4">
    <div class="mx-4 flex items-center space-x-4">
        <Button variant="ghost" size="sm" onclick={onBack}>← Back</Button>
    </div>

    <div class="space-y-3 p-4">
        <Label class="text-sm font-medium">Upload ICS file</Label>

        <div class="space-y-2">
            <Label for="ics-file" class="text-sm">Select file</Label>
            <input
                id="ics-file"
                type="file"
                accept=".ics,text/calendar"
                onchange={handleFileUpload}
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
            />
        </div>

        {#if icsFile}
            <div class="flex items-center space-x-2 rounded bg-accent/50 p-2">
                <FileText class="h-4 w-4 text-green-600" />
                <span class="text-sm">{icsFile.name}</span>
                <CheckCircle2 class="h-4 w-4 text-green-600" />
            </div>
        {/if}

        {#if error}
            <div class="rounded bg-destructive/10 p-3 text-sm text-destructive">
                {error}
            </div>
        {/if}

        <div class="space-y-2">
            <Label for="calendar-name" class="text-sm">Calendar name (optional)</Label>
            <Input
                id="calendar-name"
                type="text"
                placeholder="Enter calendar name"
                value={calendarName}
                oninput={(e) => handleNameChange(e.currentTarget.value)}
            />
        </div>

        {#if icsFile}
            <Button class="w-full" onclick={handleImport} disabled={loading}>
                {#if loading}
                    <div class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"></div>
                {/if}
                Import Calendar
            </Button>
        {/if}
    </div>
</div>
