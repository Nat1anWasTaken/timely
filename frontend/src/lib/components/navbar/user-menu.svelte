<script>
    import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.ts";
    import { goto } from "$app/navigation";

    let { children } = $props();
</script>

<DropdownMenu.Root>
    <DropdownMenu.Trigger>{@render children()}</DropdownMenu.Trigger>
    <DropdownMenu.Content>
        <DropdownMenu.Item
            onclick={() => {
                goto("/dashboard");
            }}>Dashboard</DropdownMenu.Item
        >
        <DropdownMenu.Item
            variant="destructive"
            onclick={() => {
                goto("/logout");
            }}>Logout</DropdownMenu.Item
        >
    </DropdownMenu.Content>
</DropdownMenu.Root>
