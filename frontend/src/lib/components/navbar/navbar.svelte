<script>
    import ModeToggle from "$lib/components/dashboard/mode-toggle.svelte";
    import { Button } from "$lib/components/ui/button/index.ts";
    import { Card, CardTitle } from "$lib/components/ui/card/index.ts";
    import { createUserDataQuery } from "$lib/globalQueries";
    import UserAvatar from "./user-avatar.svelte";
    import { onMount } from "svelte";

    const userDataQuery = createUserDataQuery();

    onMount(() => {
        $userDataQuery.refetch();
    });
</script>

<div class="m-4 flex max-h-16">
    <Card class="w-full flex-row items-center justify-between p-4">
        <CardTitle>
            <a class="text-xl" href="/">Timely</a>
        </CardTitle>
        <div class="flex items-center">
            <ModeToggle />
            {#if $userDataQuery.data?.user}
                <UserAvatar user={$userDataQuery.data?.user} />
            {:else}
                <Button variant="outline" class="ml-4" href="/login">Login</Button>
            {/if}
        </div>
    </Card>
</div>
