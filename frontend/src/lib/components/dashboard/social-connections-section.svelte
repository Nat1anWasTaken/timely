<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Card,
        CardContent,
        CardDescription,
        CardHeader,
        CardTitle
    } from "$lib/components/ui/card";
    import { Link, Chrome } from "@lucide/svelte";
    import SocialConnection from "./social-connection.svelte";
    import { createUserDataQuery } from "$lib/globalQueries";
    import type { ComponentType } from "svelte";

    let userDataQuery = createUserDataQuery();

    // Map provider names to icons
    const providerIcons: Record<string, ComponentType> = {
        google: Chrome
    };

    // Map provider names to display names
    const providerNames: Record<string, string> = {
        google: "Google"
    };
</script>

<Card>
    <CardHeader>
        <CardTitle class="flex items-center gap-2">
            <Link class="h-5 w-5" />
            Social Connections
        </CardTitle>
        <CardDescription>Manage your connected social accounts</CardDescription>
    </CardHeader>
    <CardContent>
        <div class="space-y-3">
            {#if $userDataQuery.data?.user.accounts && $userDataQuery.data.user.accounts.length > 0}
                {#each $userDataQuery.data.user.accounts as account}
                    <SocialConnection
                        icon={providerIcons[account.provider] || Chrome}
                        title={providerNames[account.provider] || account.provider}
                        description={account.email || "Connected"}
                    />
                {/each}
            {:else}
                <p class="text-sm text-muted-foreground">No social connections found</p>
            {/if}
        </div>
    </CardContent>
</Card>
