<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.ts";
    import { Card, CardContent } from "$lib/components/ui/card/index.ts";
    import { Calendar } from "@lucide/svelte";
    import AddCalendarDialog from "./add-calendar-dialog.svelte";
    import type { User } from "$lib/types/api";

    export let user: User | undefined = undefined;
</script>

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
            <AddCalendarDialog {user}>
                <Button size="sm">Add Calendar</Button>
            </AddCalendarDialog>
        </div>
    </CardContent>
</Card>
