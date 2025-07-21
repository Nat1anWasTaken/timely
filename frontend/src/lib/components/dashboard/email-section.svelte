<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Card,
        CardContent,
        CardDescription,
        CardHeader,
        CardTitle
    } from "$lib/components/ui/card";
    import { Mail } from "@lucide/svelte";
</script>

<Card>
    <CardHeader>
        <CardTitle class="flex items-center gap-2">
            <Mail class="h-5 w-5" />
            Email
        </CardTitle>
        <CardDescription>Your email address for account access</CardDescription>
    </CardHeader>
    <CardContent>
        <div class="flex items-center justify-between">
            <span class="text-sm">john.doe@example.com</span>
            <Button variant="outline" size="sm">Change Email</Button>
        </div>
    </CardContent>
</Card>
