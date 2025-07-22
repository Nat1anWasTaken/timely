<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { User, Calendar, X } from "@lucide/svelte";
    import type { DashboardPage } from "../../../routes/dashboard/types.ts";

    let { currentPage, isOpen = $bindable(false) }: { currentPage: DashboardPage; isOpen?: boolean } = $props();
</script>

<!-- Desktop Sidebar -->
<div class="hidden md:block w-64 rounded-lg border-r bg-card">
    <div class="p-6">
        <h2 class="text-lg font-semibold">Dashboard</h2>
    </div>
    <nav class="space-y-2 px-4">
        <Button
            variant={currentPage === "account" ? "default" : "ghost"}
            class="w-full justify-start"
            href="/dashboard/account"
        >
            <User class="mr-2 h-4 w-4" />
            Account
        </Button>
        <Button
            variant={currentPage === "calendars" ? "default" : "ghost"}
            class="w-full justify-start"
            href="/dashboard/calendars"
        >
            <Calendar class="mr-2 h-4 w-4" />
            Calendars
        </Button>
    </nav>
</div>

<!-- Mobile Sidebar Overlay -->
{#if isOpen}
    <div class="fixed inset-0 z-50 md:hidden">
        <!-- Backdrop -->
        <div class="fixed inset-0 bg-black/50" onclick={() => isOpen = false}></div>

        <!-- Sidebar -->
        <div class="fixed left-0 top-0 h-full w-64 bg-card border-r">
            <div class="flex items-center justify-between p-6">
                <h2 class="text-lg font-semibold">Dashboard</h2>
                <Button variant="ghost" size="sm" onclick={() => isOpen = false}>
                    <X class="h-4 w-4" />
                </Button>
            </div>
            <nav class="space-y-2 px-4">
                <Button
                    variant={currentPage === "account" ? "default" : "ghost"}
                    class="w-full justify-start"
                    href="/dashboard/account"
                    onclick={() => isOpen = false}
                >
                    <User class="mr-2 h-4 w-4" />
                    Account
                </Button>
                <Button
                    variant={currentPage === "calendars" ? "default" : "ghost"}
                    class="w-full justify-start"
                    href="/dashboard/calendars"
                    onclick={() => isOpen = false}
                >
                    <Calendar class="mr-2 h-4 w-4" />
                    Calendars
                </Button>
            </nav>
        </div>
    </div>
{/if}
