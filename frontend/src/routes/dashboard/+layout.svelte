<script lang="ts">
    import DashboardSidebar from "$lib/components/dashboard/dashboard-sidebar.svelte";
    import { Button } from "$lib/components/ui/button";
    import { Menu } from "@lucide/svelte";
    import { page } from "$app/state";
    import type { DashboardPage } from "./types.ts";

    let { children } = $props();

    let currentPage: DashboardPage = $derived(page.url.pathname.split("/")[2] || "account");
    let sidebarOpen = $state(false);

    function toggleSidebar() {
        sidebarOpen = !sidebarOpen;
    }
</script>

<div class="flex h-full bg-background">
    <DashboardSidebar {currentPage} bind:isOpen={sidebarOpen} />

    <!-- Main Content -->
    <div class="flex-1 overflow-auto">
        <!-- Mobile Header with Menu Button -->
        <div class="flex items-center border-b p-4 text-muted-foreground md:hidden">
            <Button variant="ghost" size="sm" onclick={toggleSidebar}>
                <Menu class="h-5 w-5" />
            </Button>
            <h1 class="ml-3 text-lg font-semibold capitalize">{currentPage}</h1>
        </div>

        <div class="p-6">
            {@render children()}
        </div>
    </div>
</div>
