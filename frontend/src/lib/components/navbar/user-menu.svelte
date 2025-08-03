<script>
    import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.ts";
    import { goto } from "$app/navigation";
    import { createUserDataQuery } from "$lib/globalQueries";
    import { api } from "$lib/api";
    import { QueryClient } from "@tanstack/svelte-query";

    let userDataQuery = createUserDataQuery();
    let queryClient = new QueryClient();

    let { children } = $props();

    async function handleLogout() {
        try {
            await api.logout();
            localStorage.removeItem("auth_token");
            await queryClient.invalidateQueries({ queryKey: ["user-profile"] });
            goto("/login");
        } catch (error) {
            console.error("Logout failed:", error);
            localStorage.removeItem("auth_token");
            await queryClient.invalidateQueries({ queryKey: ["user-profile"] });
            goto("/login");
        }
    }
</script>

<DropdownMenu.Root>
    <DropdownMenu.Trigger>{@render children()}</DropdownMenu.Trigger>
    <DropdownMenu.Content>
        <DropdownMenu.Item
            onclick={() => {
                goto(`/${$userDataQuery.data?.user?.username}`);
            }}>My Calendars</DropdownMenu.Item
        >
        <DropdownMenu.Item
            variant="destructive"
            onclick={handleLogout}>Logout</DropdownMenu.Item
        >
    </DropdownMenu.Content>
</DropdownMenu.Root>
