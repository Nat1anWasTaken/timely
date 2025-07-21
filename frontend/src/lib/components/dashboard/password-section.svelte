<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Card,
        CardContent,
        CardDescription,
        CardHeader,
        CardTitle
    } from "$lib/components/ui/card";
    import { Lock } from "@lucide/svelte";
</script>

<Card>
    <CardHeader>
        <CardTitle class="flex items-center gap-2">
            <Lock class="h-5 w-5" />
            Password
        </CardTitle>
        <CardDescription>Secure your account with a strong password</CardDescription>
    </CardHeader>
    <CardContent>
        <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">Password last updated 2 months ago</span>
            <Button variant="outline" size="sm">Change Password</Button>
        </div>
    </CardContent>
</Card>
