<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Sheet,
        SheetContent,
        SheetDescription,
        SheetHeader,
        SheetTitle,
        SheetTrigger
    } from "$lib/components/ui/sheet";
    import type { User } from "$lib/types/api";
    import CalendarManagerList from "./calendar-manager-list.svelte";
    import CalendarImportDialog from "./calendar-import-dialog.svelte";

    interface Props {
        user?: User;
        children: import("svelte").Snippet;
    }

    let { user, children }: Props = $props();

    let open = $state(false);
</script>

<Sheet bind:open>
    <SheetTrigger>
        {@render children()}
    </SheetTrigger>
    <SheetContent class="sm:max-w-lg">
        <SheetHeader>
            <SheetTitle>Manage My Calendars</SheetTitle>
            <SheetDescription>
                View and manage your imported calendars, or import new ones.
            </SheetDescription>
        </SheetHeader>

        <div class="space-y-6 p-4">
            <!-- Import Calendar Button -->
            <div class="flex justify-end">
                <CalendarImportDialog {user}>
                    <Button>Import Calendar</Button>
                </CalendarImportDialog>
            </div>

            <!-- Calendar List -->
            <CalendarManagerList />
        </div>
    </SheetContent>
</Sheet>
